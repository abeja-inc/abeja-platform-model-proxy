# abeja-platform-model-proxy

SAMPv2 Proxy

## Development

### Prerequisite

You need Go 1.17 or later.
You should set environment variable `GOPATH`, and add `$GOPATH/bin` to `PATH`.
(If you don't set `GOPATH`, Go use `$HOME/go` as `$GOPATH`)

### Getting started

#### Setup python runtime

```
$ git clone git@github.com:abeja-inc/abeja-platform-model-proxy.git
$ cd abeja-platform-model-proxy

# install tools
$ go install golang.org/x/tools/cmd/goimports@latest
$ go install github.com/rakyll/statik@latest

# build samp-v2 runner
$ make build


# Install python runtime
$ pip install abejaruntime==1.1.0

# Put python handler
$ vim main.py

# Run samp-v2 with python runtime without downloading model source
./abeja-runner service run

# Send test request
$ curl localhost:5000

```

**`main.py`**

```python

import http

def handler(request, context):

    # some inference process...

    content = {
        'transaction_id': 1234567890,
        'category_id': 10,
        'predictions': [],
    }
    return {
        'status_code': http.HTTPStatus.OK,
        'content_type': 'application/json; charset=utf8',
        'content': content
    }
```

### Formatting

Using [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports).
And use `-local` option.

```
$ goimports -local github.com/abeja-inc main.go
```

### Lint

Using [GolangCI-Lint](https://github.com/golangci/golangci-lint).

```
$ make lint
```

### Build

```
$ make build
```

### Run Tests

```
$ make test
```

## Release

Use git-flow.

Synchronize master and develop branch.

```
$ git checkout master
$ git pull --rebase origin master
$ git checkout develop
$ git pull --rebase origin develop
```

Create release branch and prepare for release.

```
$ git flow release start X.X.X
# update to new version
$ vim CHANGELOG.md
$ vim version/version.go
$ git add CHANGELOG.md version/version.go
$ git commit -m "bump version"
```

Release.

```
$ git flow release finish X.X.X
$ git push origin develop
$ git push origin X.X.X
$ git push origin master
```

Pushing the tag to Github will automatically release the binary to Github Releases.

## Contribution

> Feel free to ask the team directly about the best way to contribute!

[gitflow](https://github.com/nvie/gitflow) branching model is used, if you have a feature you want to contribute create a feature/FEATURE-NAME branch from the "develop" branch, and issue a Pull-Request to have your feature integrated.
