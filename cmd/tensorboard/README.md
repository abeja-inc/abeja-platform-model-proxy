# TensorBoard Command

> command line tools for TensorBoard application.
> Download training job results for TensorBoard to visualize them.

## Specification

### Download Training Job Results

- download specified training job's results to a directory for the tensorboard.
  - tensorboard' directory path : `/mnt/tensorboards/<tensorboard_id>`
  - training job result's directory path : `/mnt/tensorboards/<tensorboard_id>/training_jobs/<training_job_id>`
- **only training job results that belongs to the same TrainingJobDefinition of the tensorboard** are allowed to download.
- retry if fail to download training job results.
- (_T.B.D_) notify as failed to model_api, if failed after retrying.

## Usage

### Prerequisite

need to set following environment variables.

```
export ABEJA_API_URL=https://api.dev.abeja.io
export PLATFORM_AUTH_TOKEN=xxxxx
export ABEJA_ORGANIZATION_ID=1230000000000
export TRAINING_JOB_DEFINITION_NAME=training-job-testing
export TENSORBOARD_ID=1800000000000
# NOTE: these training jobs should be of the same TrainingJobDefinition where tensorboard belongs
export TRAINING_JOB_IDS=1500000000000,1500000000001
# For Debugging
export ABEJA_MOUNT_TARGET_DIR=/tmp/mnt
```

### Command Usage

```
$ ./abeja-runner tensorboard -h
download user training-results

Usage:
  abeja-runner tensorboard [flags]

Flags:
      --abeja_api_url string                  base url of abeja-api (default "https://api.abeja.io")
      --abeja_organization_id string          identifier of organization
  -h, --help                                  help for tensorboard
      --mount_target_dir string               directory to mount shared file system (default "/mnt")
      --platform_auth_token string            authentication token for platform
      --tensorboard_id string                 identifier of tensorboard
      --training_job_definition_name string   name of training job definition
      --training_job_ids string               comma separated list of identifier of training job
```

## Development

use Docker container environment.

```
$ docker run --rm -it -v `pwd`:/work -w /work golang:1.12.16-stretch bash
```

**build binary**

```
# prerequisite
$ go install github.com/rakyll/statik@latest

$ make build
```
