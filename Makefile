.DEFAULT_GOAL := all

.PHONY: all
all: test lint

# the TEST_FLAGS env var can be set to eg run only specific tests
TEST_COMMAND = go test -v -coverprofile cover.out -count=1 -race -cover $(TEST_FLAGS)

.PHONY: test
test:
	$(TEST_COMMAND)

.PHONY: test_with_coverage
test_with_coverage: test
	go tool cover -html=cover.out

.PHONY: lint
lint:
	golangci-lint run
