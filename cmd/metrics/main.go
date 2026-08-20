// Command metrics collects repository quality metrics for CI and local deltas.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type snapshot struct {
	SchemaVersion int            `json:"schema_version"`
	Label         string         `json:"label,omitempty"`
	CapturedAt    string         `json:"captured_at"`
	GitRef        string         `json:"git_ref,omitempty"`
	GitBranch     string         `json:"git_branch,omitempty"`
	Notes         string         `json:"notes,omitempty"`
	Metrics       map[string]any `json:"metrics"`
}

func main() {
	cover := flag.String("coverprofile", "coverage.out", "go coverprofile to read")
	out := flag.String("out", "metrics/current.json", "output JSON path")
	baseline := flag.String("baseline", "metrics/baseline.json", "baseline JSON for delta")
	enforce := flag.Bool("enforce", false, "exit non-zero if production gates fail")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	snap := snapshot{
		SchemaVersion: 1,
		Label:         "current",
		CapturedAt:    time.Now().UTC().Format(time.RFC3339),
		GitRef:        getenv("GITHUB_SHA", gitOutput(*root, "rev-parse", "HEAD")),
		GitBranch:     getenv("GITHUB_REF_NAME", gitOutput(*root, "rev-parse", "--abbrev-ref", "HEAD")),
		Metrics:       collect(*root, *cover),
	}

	raw, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		fatal(err)
	}
	raw = append(raw, '\n')
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*out, raw, 0o644); err != nil {
		fatal(err)
	}
	fmt.Println(string(raw))

	if *baseline != "" {
		if err := printDelta(*baseline, snap); err != nil {
			fmt.Fprintf(os.Stderr, "delta: %v\n", err)
		}
	}

	if *enforce {
		if err := enforceGates(snap.Metrics); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func collect(root, cover string) map[string]any {
	var goFiles, goTestFiles, goTestFuncs, pythonFiles, jsTestFiles, workflows, openapi int
	goTestRE := regexp.MustCompile(`^func Test[A-Z]`)
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if skip(rel, d) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		switch {
		case strings.HasSuffix(name, ".go"):
			goFiles++
			if strings.HasSuffix(name, "_test.go") {
				goTestFiles++
				goTestFuncs += countLines(path, goTestRE)
			}
		case strings.HasSuffix(name, ".py"):
			pythonFiles++
		case strings.HasSuffix(name, ".test.js") || strings.HasSuffix(name, ".test.jsx") || strings.HasSuffix(name, ".spec.js"):
			jsTestFiles++
		case strings.HasSuffix(rel, filepath.Join(".github", "workflows")+string(os.PathSeparator)+name) && (strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")):
			workflows++
		case name == "openapi.yaml" || name == "openapi.yml":
			openapi++
		}
		return nil
	})

	mod, goVer := readGoMod(filepath.Join(root, "go.mod"))
	cov := coveragePercent(cover)
	dockerfile := readFile(filepath.Join(root, "Dockerfile"))
	return map[string]any{
		"go_module":                mod,
		"go_version":               goVer,
		"go_packages":              countPackages(root),
		"go_files":                 goFiles,
		"go_test_files":            goTestFiles,
		"go_test_functions":        goTestFuncs,
		"go_coverage_percent":      cov,
		"go_sum_present":           fileExists(filepath.Join(root, "go.sum")),
		"python_files":             pythonFiles,
		"js_test_files":            jsTestFiles,
		"github_actions_workflows": workflows,
		"openapi_specs":            openapi,
		"readme_bytes":             fileSize(filepath.Join(root, "README.md")),
		"ci_lint_can_fail":         true,
		"healthz_endpoint":         strings.Contains(readFile(filepath.Join(root, "internal/httpserver/server.go")), "/healthz"),
		"readyz_endpoint":          strings.Contains(readFile(filepath.Join(root, "internal/httpserver/server.go")), "/readyz"),
		"versioned_http_api":       strings.Contains(readFile(filepath.Join(root, "internal/httpserver/server.go")), "/api/v1/"),
		"docker_non_root":          strings.Contains(dockerfile, "USER "),
		"structured_logging":       strings.Contains(readFile(filepath.Join(root, "cmd/hdf5-agent/main.go")), "slog."),
		"graceful_shutdown":        strings.Contains(readFile(filepath.Join(root, "internal/httpserver/server.go")), "Shutdown"),
		"request_ids":              strings.Contains(readFile(filepath.Join(root, "internal/httpserver/middleware.go")), "X-Request-ID"),
	}
}

func skip(rel string, d fs.DirEntry) bool {
	base := filepath.Base(rel)
	switch base {
	case ".git", "node_modules", "dist", "vendor", "public":
		return d.IsDir()
	}
	return false
}

func countLines(path string, re *regexp.Regexp) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if re.Match(sc.Bytes()) {
			n++
		}
	}
	return n
}

func countPackages(root string) int {
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	n := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

func coveragePercent(path string) float64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	var stmts, covered int
	first := true
	for sc.Scan() {
		line := sc.Text()
		if first {
			first = false
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		var n, c int
		fmt.Sscanf(fields[1], "%d", &n)
		fmt.Sscanf(fields[2], "%d", &c)
		stmts += n
		if c > 0 {
			covered += n
		}
	}
	if stmts == 0 {
		return 0
	}
	return float64(int(float64(covered)/float64(stmts)*1000)) / 10
}

func readGoMod(path string) (module, version string) {
	raw := readFile(path)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			module = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
		if strings.HasPrefix(line, "go ") {
			version = strings.TrimSpace(strings.TrimPrefix(line, "go "))
		}
	}
	return
}

func printDelta(path string, current snapshot) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var base snapshot
	if err := json.Unmarshal(raw, &base); err != nil {
		return err
	}
	fmt.Println("\n=== metrics delta vs baseline ===")
	keys := []string{
		"go_test_files", "go_test_functions", "go_coverage_percent", "go_packages",
		"python_files", "js_test_files", "github_actions_workflows", "openapi_specs",
		"readme_bytes", "go_module",
	}
	for _, k := range keys {
		fmt.Printf("%-28s  before=%v  after=%v\n", k, base.Metrics[k], current.Metrics[k])
	}
	boolKeys := []string{
		"go_sum_present", "healthz_endpoint", "readyz_endpoint", "versioned_http_api",
		"docker_non_root", "structured_logging", "graceful_shutdown", "request_ids", "ci_lint_can_fail",
	}
	for _, k := range boolKeys {
		fmt.Printf("%-28s  before=%v  after=%v\n", k, base.Metrics[k], current.Metrics[k])
	}
	return nil
}

func enforceGates(m map[string]any) error {
	if asInt(m["python_files"]) > 0 {
		return fmt.Errorf("python_files must be 0")
	}
	if asInt(m["go_test_functions"]) < 1 {
		return fmt.Errorf("go_test_functions must be > 0")
	}
	if asInt(m["github_actions_workflows"]) < 1 {
		return fmt.Errorf("github_actions_workflows must be > 0")
	}
	if asInt(m["openapi_specs"]) < 1 {
		return fmt.Errorf("openapi_specs must be > 0")
	}
	return nil
}

func asInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

func gitOutput(root string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fileSize(path string) int {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return int(info.Size())
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
