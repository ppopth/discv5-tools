allpackages := $(shell go list ./...)

.PHONY: all
all: build

# Add all projects in the cmd directory
CMDS := $(notdir $(wildcard $(CURDIR)/cmd/*))
.PHONY: %.gocmd
%.gocmd:
	go build -o bin/$* github.com/ppopth/discv5-tools/cmd/$*

.PHONY: build
build: $(CMDS:%=%.gocmd)

.PHONY: clean
clean:
	rm -rf bin

.PHONY: test
test:
	@go vet $(allpackages)
	@go test -v $(allpackages)
