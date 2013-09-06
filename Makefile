all: test

dependencies:
	go get -d .

coverage:
	gocov test ./... | gocov-html > coverage.html
	open coverage.html

bench:
	go test -bench=. ./...

fmt:
	go fmt  ./...

proto: 
	find . -iname *.proto | xargs -J %  protoc --go_out=. %

test:
	go test ./...

testv:
	go test -v ./...

install:
	go install ./...

.PHONY: coverage dependencies test
