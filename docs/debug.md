# Debugging Triage Party

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Server](#server)
- [Tester](#tester)
- [Disabling persistent cache](#disabling-persistent-cache)
- [Making RAW JSON requests](#making-raw-json-requests)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Server

For basic debugging, add `-v=<level>` to the server arguments. The verbosity goes up as you increase the number:

* `-v=1`: Particularly tricky sections will log here. Cache misses too.
* `-v=2`: Noisy. Enough to debug most matching issues.
* `-v=3`: Very noisy, and usually not very useful.

## Tester

For pin-point debugging, Triage Party includes a separate `tester` tool to run a specific rule and dump raw JSON data from GitHub on a particular PR or issue number.

Here is real-world example for using the tester to debug why a PR `#4126` was listed in a particular rule (`pr-reviewable`).

`go run cmd/tester/main.go --github-token-file ~/.github-personal-read --config config/examples/skaffold.yaml --rule pr-reviewable -v=3 --num 4126`

Here is some example output:

```json
I0521 07:33:27.139885    5503 match.go:17] pre-matching item #4126 against filter: state: open
E0521 07:33:27.139960    5503 search.go:279] debug comments: null
E0521 07:33:27.142985    5503 search.go:291] *** Debug PR timeline #4126:
[
  {
    "event": "committed",
    "url": "https://api.github.com/repos/GoogleContainerTools/skaffold/git/commits/cf551ff40453965987933f80b4b662ac604eb158"
  },
  {
    "event": "committed",
    "url": "https://api.github.com/repos/GoogleContainerTools/skaffold/git/commits/89b2527cf11c730fdf7a07820945f811f23a711a"
  },
```

From this I was able to see the `debug comments: null` hint that allowed me to investigate why no comments were fetched for this PR.

If you find it useful to add debugging that only triggers on a particular issue or PR number, this code block is useful:

```go
if h.debug[pr.GetNumber()] {
    klog.Infof("debug response: %+v", formatStruct(resp))
}
```

## Disabling persistent cache

For both the server and tester: `--persist-backend=memory`

## Making RAW JSON requests

See the [GitHub v3 JSON API](https://developer.github.com/v3/) documentation. Here is an example, which returns an empty list:

`curl -H "Authorization: token <token value>" https://api.github.com/repos/GoogleContainerTools/skaffold/pulls/4126/comments`
