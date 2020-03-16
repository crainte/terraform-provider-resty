VERSION=v0.0.2
TEST?=$$(go list ./...)
GOFMT_FILES?=$$(gofmt -l `find . -name '*.go'`)

default: build

build: fmtcheck
	go build -o build/bin/terraform-provider-resty

release: fmtcheck
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/bin/terraform-provider-resty_$(VERSION)-linux-amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/bin/terraform-provider-resty_$(VERSION)-darwin-amd64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/bin/terraform-provider-resty_$(VERSION)-windows-amd64

test: fmtcheck
	@go test -i $(TEST) || exit 1
	@gotestsum --format testname $(TEST)

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@echo " >> Checking that code follows gofmt"
ifeq ($(GOFMT_FILES),)
	@echo "gofmt needs to be run on the following files:"
	@echo "$(GOFMT_FILES)"
	@echo "You can use the command 'make fmt' to reformat code."
	@exit 1
endif

clean:
	rm build/bin/*

.PHONY: build release fmt fmtcheck clean
