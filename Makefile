.PHONY: run build check test vet fmt doc

DOC_ADDR ?= localhost:7777

run:
	go run .

build:
	go build ./...

check:
	go build ./... && go run .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

doc:
	@echo "ドキュメントを http://$(DOC_ADDR)/ で公開します"
	go run ./cmd/docserver -addr=$(DOC_ADDR)

