.PHONY: help setup build test vet lint fmt metrics docker clean testdata

GREEN := \033[0;32m
BLUE := \033[0;34m
NC := \033[0m

help:
	@echo "$(BLUE)HDF5 Agent$(NC)"
	@echo "  $(GREEN)make setup$(NC)     Install Go and frontend dependencies"
	@echo "  $(GREEN)make test$(NC)      Run Go and frontend tests (fails on failure)"
	@echo "  $(GREEN)make vet$(NC)       go vet"
	@echo "  $(GREEN)make lint$(NC)      Go vet + frontend eslint"
	@echo "  $(GREEN)make metrics$(NC)   Write metrics/current.json and print baseline delta"
	@echo "  $(GREEN)make build$(NC)     Build service binary and frontend"
	@echo "  $(GREEN)make testdata$(NC)  Write data/sample.h5"
	@echo "  $(GREEN)make docker$(NC)    Build and run docker compose"
	@echo "  $(GREEN)make clean$(NC)     Remove build artifacts"

setup:
	go mod download
	cd frontend && npm ci

vet:
	go vet ./...

fmt:
	gofmt -w .
	cd frontend && npm run lint

lint: vet
	test -z "$$(gofmt -l .)"
	cd frontend && npm run lint

test:
	CGO_ENABLED=1 go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic
	cd frontend && npm test

metrics: test
	go run ./cmd/metrics -coverprofile=coverage.out -out=metrics/current.json -baseline=metrics/baseline.json -enforce

build:
	mkdir -p public data
	cd frontend && npm run build && rm -rf ../public && cp -R dist ../public
	CGO_ENABLED=1 go build -o hdf5-agent ./cmd/hdf5-agent

testdata:
	mkdir -p data
	go run ./cmd/create-testdata -out data/sample.h5

docker:
	mkdir -p data
	docker compose up --build

clean:
	rm -f hdf5-agent coverage.out
	rm -rf frontend/dist public

.DEFAULT_GOAL := help
