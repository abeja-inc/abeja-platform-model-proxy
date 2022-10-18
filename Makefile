clean:
	rm -rf runtime abeja-runner

runtime: clean
	statik -src=./runtime_src -p=runtime

.PHONY: build
build: runtime
	go build -o abeja-runner

.PHONY: init
init:
	go mod init

.PHONY: clean-test
clean-test:
	go clean -testcache

.PHONY: test
test: clean-test
	go test ./...

TARGET = Test
.PHONY: test-target
test-target: clean-test
	go test -run ${TARGET} ./...

.PHONY: download
download:
	go mod download

.PHONY: verify
verify:
	go mod verify

.PHONY: lint
lint:
	golangci-lint run

.PHONY: linux
linux:
	GOOS=linux GOARCH=amd64 go build -o abeja-runner
