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

.PHONY: test-all
test-all: clean-test
	go test -tags=extra ./...

TARGET = Test
.PHONY: test-target
test-target: clean-test
	go test -run ${TARGET} ./...

.PHONY: test-target-all
test-target-all: clean-test
	go test -run ${TARGET} -tags=extra ./...

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
