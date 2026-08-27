## Description
<!-- What changed and why. -->

## Metrics
<!-- Paste `make metrics` delta vs metrics/baseline.json if this PR touches quality gates. -->

## Testing
- [ ] `go test ./...`
- [ ] `cd frontend && npm test && npm run lint && npm run build`
- [ ] API health: `curl localhost:8080/healthz`

## Checklist
- [ ] No Python added
- [ ] Go docs on exported APIs
- [ ] OpenAPI updated if HTTP API changed
