.PHONY: build test coverage lint check clean
.DELETE_ON_ERROR:

COVERAGE_THRESHOLD ?= 20
VERSION    := $(shell cat VERSION)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

GO_FILES   := $(shell find . -name '*.go')
SOURCES    := $(GO_FILES) go.mod go.sum VERSION

tool-builder: $(SOURCES)
	@go build -ldflags "$(LDFLAGS)" -o $@ .

build: tool-builder

coverage.out: $(SOURCES)
	@go test -coverprofile=$@ ./...

test: coverage.out

coverage: coverage.out
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@TOTAL=$$(go tool cover -func=coverage.out | grep '^total:' | awk '{print $$3}' | tr -d '%'); \
	echo "Coverage: $${TOTAL}%  (threshold: $(COVERAGE_THRESHOLD)%)"; \
	awk -v t="$${TOTAL}" -v th="$(COVERAGE_THRESHOLD)" \
	  'BEGIN { if (t+0 < th+0) { print "FAIL: coverage below threshold"; exit 1 } }'
	@echo "Full report: coverage.html"

lint:
	@golangci-lint run ./...

check:
ifndef SKIP_LINTERS
	$(MAKE) lint
endif
ifndef SKIP_TESTS
	$(MAKE) coverage
endif

clean:
	@rm -f tool-builder coverage.out coverage.html
