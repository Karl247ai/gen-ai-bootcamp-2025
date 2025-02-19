.PHONY: test coverage coverage-html

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run all tests and generate coverage report
test-all: test coverage coverage-html

# Run integration tests only
test-integration:
	go test -v ./test/integration/...

# Run unit tests only
test-unit:
	go test -v ./internal/...

# Clean test cache and coverage files
clean:
	go clean -testcache
	rm -f coverage.out 