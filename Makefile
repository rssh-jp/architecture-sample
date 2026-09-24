.PHONY: run build check test vet fmt setup doc

DOC_ADDR ?= localhost:7777

run:
	go run ./cmd/check

build:
	go build ./...

check:
	go build ./... && go run ./cmd/check

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

setup:
	go install golang.org/x/tools/cmd/godoc@latest

doc:
	@echo "ドキュメントを http://$(DOC_ADDR)/pkg/sample で公開します"
	godoc -http=$(DOC_ADDR)

