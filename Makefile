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
	# Pretty hacky way to add BSON annotations to our Protobuf structs
	find . -iname *.pb.go | xargs -J % perl -pi -e 's|json:"(.*)"|json:\"$$1\" bson:\"$$1\"|g' %
	find . -iname *.proto | xargs -J % protoc --proto_path=. --python_out=ui/app %

test:
	go test ./...

testv:
	go test -v ./...

install:
	go install ./...

lint:
	find . -iname '*.go' | grep -v protobufs | xargs golint 

.PHONY: coverage dependencies test
