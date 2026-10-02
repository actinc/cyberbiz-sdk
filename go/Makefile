.PHONY: test lint fmt integration docs console golden

test:
	go test -race -cover ./...
	if [ -d console ]; then cd console && go test -race ./cmd/... ./internal/...; fi

lint:
	test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }
	go vet ./...
	if [ -d console ]; then cd console && go vet ./cmd/... ./internal/...; fi

fmt:
	gofmt -w .

# Read-only calls against the live API. Needs CYBERBIZ_API_TOKEN.
integration:
	go test -tags integration -run TestLive -v ./cyberbiz

# Regenerate docs/api/en from docs/references and testdata/golden.
docs:
	cd console && go run ./cmd/gendocs

# Import raw responses (e.g. from a sweep) as redacted Golden Files.
golden:
	go run ./internal/tools/goldenimport -in $(IN) -out ../testdata/golden

console:
	cd console && go run ./cmd/console
