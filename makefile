export GOWORK=off

.PHONY: all test lint tidy tools-update tools-tidy upgrade-direct-dependencies build install uninstall clean

all: tidy build lint test

install:
	go install .

uninstall:
	go clean -i .
	rm -f $(shell go env GOBIN)/antigravity-quota $(shell go env GOPATH)/bin/antigravity-quota

build:
	go build -o build/antigravity-quota ./...

clean:
	rm -rf build/

test:
	go test -v -race ./...

lint:
	go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...

tidy:
	go mod tidy
	git diff --exit-code go.mod go.sum

tools-update:
	for d in tools/* ; do \
		if [ -d $$d ] && [ -f $$d/go.mod ] ; then \
			echo "Updating $$d" ; \
			cd $$d ; \
			go get -u $$(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all) ; \
			go mod tidy ; \
			cd - >/dev/null; \
		fi ; \
	done

tools-tidy:
	for d in tools/* ; do \
		if [ -d $$d ] && [ -f $$d/go.mod ] ; then \
			echo "tidy $$d" ; \
			cd $$d ; \
			go mod tidy ; \
			cd - >/dev/null; \
		fi ; \
	done

upgrade-direct-dependencies:
	echo "Upgrading direct dependencies..."
	go get -u $$(go list -m -f '{{if and (not .Indirect) (not .Main)}}{{.Path}}@latest{{end}}' all)
	go mod tidy
