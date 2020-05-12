# Release Notes

## Version 1.0.0-beta.4 - 2020-05-11

Improvements:

* Publish Docker image, split Dockerfile into base and default [#90](https://github.com/google/triage-party/pull/90)
* Add CloudSQL support for Postgres [#89](https://github.com/google/triage-party/pull/89)
* Support GitHub Enterprise [#64](https://github.com/google/triage-party/pull/64)
* Improve example Kubernetes manifests [#91](https://github.com/google/triage-party/pull/91)
* Make stale data warning more informative and fire less often [#88](https://github.com/google/triage-party/pull/88)
* Add custom.css override file, minor UI tweaks [#81](https://github.com/google/triage-party/pull/81)
* Include timeline metadata: new 'prioritized' rule, new 'new-commits' tag [#72](https://github.com/google/triage-party/pull/72)
* Log an error when rate limited by GitHub [#71](https://github.com/google/triage-party/pull/71)

Bugfixes:

* add tikv to persist [#73](https://github.com/google/triage-party/pull/73)
* Build similarity info on cached data, add example 'similar' example page [#87](https://github.com/google/triage-party/pull/87)
* Stale notification: use save time instead of item update time [#84](https://github.com/google/triage-party/pull/84)
* Only download closed issues & PR's when required [#74](https://github.com/google/triage-party/pull/74)

Thank you to our most recent contributors!

- Mahmoud
- Shingo Omura
- Thomas Strömberg
- Travis Tomsu

## Version 1.0.0-beta.3 - 2020-05-06

* Add 'postgres' persistence backend for PostgreSQL & CockroachDB [#65](https://github.com/google/triage-party/pull/65)
* Improve examples: tighten similarity, fix yaml errors [#70](https://github.com/google/triage-party/pull/70)
* UI: add titles to tags, improve similarity/omit display [#68](https://github.com/google/triage-party/pull/68)
* Improve refresh performance through better caching  [#67](https://github.com/google/triage-party/pull/67)
* Improve similarity scoring by removing junk words [#66](https://github.com/google/triage-party/pull/66)
* Separate persist loop from content update loop [#60](https://github.com/google/triage-party/pull/60)

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
