GOPATH=$(shell git rev-parse --show-toplevel)

PACKAGES=aoc/...
PACKAGES+=github.com/jcorbin/anansi/...

DAYS=$(shell \
	find src/aoc -type d -mindepth 1 -maxdepth 1 -name 'day*' \
	| xargs -n1 basename \
	| sort )

.PHONY: days
days: $(DAYS)

.PHONY: $(DAYS)
$(DAYS):
	go build aoc/$@

.PHONY: test
test: lint
	export GOPATH=$(GOPATH)
	go test $(PACKAGES)

.PHONY: lint
lint:
	export GOPATH=$(GOPATH)
	./bin/go_list_sources.sh $(PACKAGES) | xargs gofmt -e -d
	golint $(PACKAGES)
	go vet $(PACKAGES)

.PHONY: fmt
fmt:
	export GOPATH=$(GOPATH)
	./bin/go_list_sources.sh $(PACKAGES) | xargs gofmt -w
