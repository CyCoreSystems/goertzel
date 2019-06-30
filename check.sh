#!/bin/bash -e
go build ./...
golangci-lint run
go test -race -coverprofile=coverage.txt -covermode=atomic
