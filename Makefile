doc-dev:
	@uv run -m mkdocs serve
.PHONY: doc-dev

doc-build:
	@uv run -m mkdocs build
.PHONY: doc-build

# Test coverage targets
coverage:
	@echo "Running tests with coverage..."
	@go test ./... -race -coverprofile=coverage.out
	@echo "\nCoverage by function:"
	@go tool cover -func=coverage.out
.PHONY: coverage

coverage-html:
	@echo "Generating HTML coverage report..."
	@go test ./... -race -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "HTML report generated: coverage.html"
.PHONY: coverage-html

coverage-summary:
	@echo "Generating coverage summary..."
	@mkdir -p coverage
	@go test ./... -race -covermode=atomic -coverprofile=coverage/coverage.out
	@go tool cover -func=coverage/coverage.out | tee coverage/summary.txt
	@echo "Coverage summary saved to coverage/summary.txt"
.PHONY: coverage-summary

clean-coverage:
	@rm -f coverage.out coverage.html
	@rm -rf coverage
	@echo "Coverage files cleaned"
.PHONY: clean-coverage
