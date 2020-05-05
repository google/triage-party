# Release Notes

## Version 1.0.0-beta.2 - 2020-05-05

Improvements:

* Persistent cache refactor with MySQL support [#55](https://github.com/google/triage-party/pull/55)
* Similarity rewrite to improve latency and hit rate [#49](https://github.com/google/triage-party/pull/49)
* Show age mouseover as time.Duration instead of static date [#47](https://github.com/google/triage-party/pull/47)
* Average last 2 colection requests for refresh rate [#46](https://github.com/google/triage-party/pull/46)
* Refactor cache interfaces to accept stale data during startup [#33](https://github.com/google/triage-party/pull/33)
* Add configuration validation [#31](https://github.com/google/triage-party/pull/31)

Bug fixes:

* Exclude ourselves and dupe URL's from similarity list [#53](https://github.com/google/triage-party/pull/53)
* Fix infinite cache regression, simplify flags [#44](https://github.com/google/triage-party/pull/44)

Thank to you our contributors:

* Ruth Cheesley
* Thomas Strömberg
* Yuki Okushi

## Version 1.0.0-beta.1 - 2020-04-27

Improvements:

* Add 'title' filter regexp [#21](https://github.com/google/triage-party/pull/21)
* Add 'draft' tag to draft PRs [#19](https://github.com/google/triage-party/pull/19)
* Increase player count to 20, preserve GET variables on page changes [#17](https://github.com/google/triage-party/pull/17)
* Simplify terminology: strategy is now collection, tactic is now rule [#14](https://github.com/google/triage-party/pull/14)
* Massive refactor: split triage and hubbub packages [#15](https://github.com/google/triage-party/pull/15)

Bug fxes:

* Refactor average/total durations to not overflow [#18](https://github.com/google/triage-party/pull/18)

Thank you to our contributors:

* James Munnelly
* Martin Pool
* Medya Gh
* Teppei Fukuda

## Version v2020-04-22.1 - 2020-04-22

Second alpha release.

Fixes Docker build script to not leak GITHUB_TOKEN into environment.

## Version v2020-04-22.0 - 2020-04-22

Initial alpha release
