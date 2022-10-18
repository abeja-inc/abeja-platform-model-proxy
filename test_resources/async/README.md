# Test for Async Request

This directory contains test codes for async request.
(Since there was no way to test asynchronous requests well, I wrote a code to check the operation manually.)

### run docker

```
$ docker run -it -v $(root of this repository):/tmp/work -w /tmp/work abeja/all-cpu:19.04 bash
```

### install golang

read [Getting Started](https://golang.org/doc/install) in golang.org.

### build SAMPv2

```
$ cd /tmp/work
$ make build
```

### build ARMS dummy

```
$ cd /tmp/work/test_resources/async/
$ go build arms_dummy.go
```

### start ARMS dummy

```
$ ./arms_dummy
```

### start SAMPv2

```
(from another terminal)
$ docker ps
$ docker exec -it <docker container id> bash
--- (on docker) ---
$ cd /tmp/work/test_resources/async
$ ABEJA_API_URL=http://localhost:8080 HANDLER=main:handler ABEJA_DEPLOYMENT_ID=2222222222222 ABEJA_ORGANIZATION_ID=1111111111111 ../../abeja-runner service run
```

### request to SAMPv2

```
(from another terminal)
$ docker exec -it <docker container id> bash
--- (on docker) ---
$ curl http://localhost:5000/ -X POST -H 'x-abeja-arms-async-request-id: req-id-1' -H 'x-abeja-arms-async-request-token: token' -H 'Content-Type: application/json' -d '{"key":"value"}'
```
