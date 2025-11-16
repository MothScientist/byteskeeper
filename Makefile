# The default target when calling make without arguments
.DEFAULT_GOAL := run

run:
	go vet .
	go run -race .

test:
	go test -v -race .

vet:
	go vet .

build:
	go build .