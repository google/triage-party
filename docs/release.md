# Release instructions
   

1. Update `VERSION` in `pkg/site/site.go`
2. In upstream repository, run `git fetch && git pull && hacks/release_note.sh`
3. Update `CHANGELOG.md`
4. Get PR's merged
5. In upstream repository, run Docker example from `README.md`
6. Run:

```
 docker build -t=triageparty/triage-party -f release.Dockerfile .
 docker push triageparty/triage-party:latest 
 ```

7. Run Kubernetes example: https://github.com/google/triage-party/blob/master/docs/deploy.md#kubernetes
8. Run:

```
export TAG=v1.1.0
git tag -a $TAG -m "$TAG release"
git push --tags
```

9. Run:

```
docker tag triageparty/triage-party:latest triageparty/triage-party:1.1.0
docker push triageparty/triage-party:1.1.0
```

10. Make GitHub release