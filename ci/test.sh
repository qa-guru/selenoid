#!/bin/bash

set -e

export GOTOOLCHAIN=go1.26.6+auto
export GO111MODULE="on"
go test -tags 's3 metadata' -v -race -coverprofile=coverage.txt -covermode=atomic -coverpkg github.com/qa-guru/selenoid,github.com/qa-guru/selenoid/session,github.com/qa-guru/selenoid/config,github.com/qa-guru/selenoid/protect,github.com/qa-guru/selenoid/service,github.com/qa-guru/selenoid/upload,github.com/qa-guru/selenoid/info,github.com/qa-guru/selenoid/jsonerror

GOTOOLCHAIN=go1.26.6 go run golang.org/x/vuln/cmd/govulncheck@v1.5.0 -tags production ./...
