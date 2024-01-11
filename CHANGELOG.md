# 2.6.5
- Retry strategy changed to EXPONENTIAL BACKOFF
- The number of retries for the download function is 10 and the maximum wait time is 17 minutes
- Added log output

# 2.6.4
- fix dependabot alerts (#71)

# 2.6.3
- wait 10 minutes on download something (#68)

# 2.6.2
- fix to terminate correctly on dial error (#67)

# 2.6.1
- Modify Downloader to use RetryClient instead of net/http/client (#66)

# 2.6.0
- fix returncode when the user model raise exception on batch (#63)
- implement metadata on batch (#64)

# 2.5.2
- change the command name of the sub-process in python-3 (#62)

# 2.5.1
- flush subprocess's log before exit (#61)

# 2.5.0
- disable keepalive connection (#60)

# 2.4.3
- log output to file (#59)

# 2.4.2
- allow python runtime to exit with exit-code 120 (#58)

# 2.4.1
- implement access log (#57)

# 2.4.0
- Feature/tensorboard command proto (#50)
- Fix statik version v0.1.6 (#53, #54)
- fix: log output cannot be used more than 64kB (#56)

# 2.3.1
- fix panic on runtime error (#49)

# 2.3.0
- remove expantion of environment variables (#48)

# 2.2.3
- scrape blank lines of runtime log (#47)

# 2.2.2
- fix timestamp format for logging (#46)

# 2.2.1
- fix validator of environments for training-model (#45)

# 2.2.0
- change error message depending on the method (#43)
- implement batch command (#44)

# 2.1.1
- add handler-func of health-check to ServiceHandler (#41)

# 2.1.0
- change log format (#37), (#38), (#39)
- separate health check server from service server (#40)

# 2.0.1
- fix: the runtime terminates normally but ends abnormally (#36)

# 2.0.0
- rename command name to abeja-runner(#32)
- unify training command (#31)
- re-structure commands (#31)

# 1.0.7
- fix datadog settings (#30)

# 1.0.6
- set up datadog (#29)

# 1.0.5
- send error response on async when signal received (#28)

# 1.0.4
- change status code on async error (#27)

# 1.0.3
- implement error response for async (#26)

# 1.0.2
- logging runtime's exit code (#25)

# 1.0.1
- fix logging (#24)

# 1.0.0
- implement async request (#22)

# 0.9.1
- fix: receive error of subprocess through channel

# 0.9.0
- Use deep-copy (#20)

# 0.8.0
- Add trainingjob response serializer
- Add new env for downloading resources

# 0.7.1
- add other resource ids in response header

# 0.7.0
- load user-env.

# 0.6.0
- add `Content-Length` to response header.

# 0.5.0
- add `x-abeja-model-version` to response header.
- add request-header to ContentList.

# 0.4.2
- update logging settings.

# 0.4.1
- Use environment variables ABEJA_USER_MODEL_ROOT, ABEJA_TRAINING_RESULT_DIR.

# 0.4.0
- separate download and run.

# 0.3.4
- add proxy version to log and response header.

# 0.3.3
- corresponding to the lint tools.

# 0.3.2
- use random file path of unix-domain-socket for inter communication to runtime.
- use Go 1.12.1 for building binary.

# 0.3.1
- fix error handling

# 0.3.0
- support Content-Type `multipart/form-data` of the Request.

# 0.2.0
- support binary type Content-Type ( like `image/jpeg` ) of the Request.
  - NOTE: `multipart/form-data` doesn't supported yet.

# 0.1.0
- first release
- It supports to the following Content-Type of the Request:
  - text/*
  - application/json
  - query-string of GET Method

