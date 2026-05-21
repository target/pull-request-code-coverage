GO_FILES=$(shell find . -type f -iregex '.*\.go')
GO_PKGS=$(shell go list ./... | grep -v -e "/resources")

format: check-gofmt


check-gofmt:
	@echo "Checking formatting..."
	@FMT="0"; \
	for pkg in $(GO_FILES); do \
		OUTPUT=`gofmt -l $$pkg`; \
		if [ -n "$$OUTPUT" ]; then \
			echo "$$OUTPUT"; \
			FMT="1"; \
		fi; \
	done ; \
	if [ "$$FMT" -eq "1" ]; then \
		echo "Problem with formatting in files above."; \
		exit 1; \
	else \
		echo "Success - way to run gofmt!"; \
	fi

GOLANGCI_LINT_VERSION=v2.12.2

bin/golangci-lint:
	mkdir -p bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b ./bin $(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: bin/golangci-lint
	bin/golangci-lint run ./...