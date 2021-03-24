<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Release instructions](#release-instructions)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Release instructions


1. Update `VERSION` in `pkg/site/site.go`
2. In upstream repository, run `git fetch && git pull && hacks/release_notes.sh`
3. Update `CHANGELOG.md`
4. Get PR's merged
5. In upstream repository, run Docker example from `README.md`
6. Run:

```
 docker build -t=triageparty/triage-party -f release.Dockerfile .
 docker push triageparty/triage-party:latest
 ```

7. Run Kubernetes example: https://github.com/google/triage-party/blob/master/docs/deploy.md#kubernetes
8. Run `./hacks/release.sh`
10. Make GitHub release
