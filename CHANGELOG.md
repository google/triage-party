# Release Notes

## Version 1.4.0 - 2021-05-25

* Parallelize Pull Request & Issue ingestion (2X faster load time) [#260](https://github.com/google/triage-party/pull/260)
* Minor stylesheet improvements [#261](https://github.com/google/triage-party/pull/261)
* Fix a bug listing pull requests [#259](https://github.com/google/triage-party/pull/259)
* Fix "fonud" typo in readme [#254](https://github.com/google/triage-party/pull/254)
* GH vars: Store pointers [#258](https://github.com/google/triage-party/pull/258)
* Make persist easier to use by other programs [#257](https://github.com/google/triage-party/pull/257)
* Rip out vestigial remains of old persistence layer [#256](https://github.com/google/triage-party/pull/256)
* Refactor persist for a bi-level readthrough cache: memory & persistent [#255](https://github.com/google/triage-party/pull/255)

Thank you to our contributors!

- Josh Hunt
- Stephen Wayne
- Steven Powell
- Thomas Strömberg

## Version 1.4.0-beta.2 - 2021-03-02

* Bump github library from v31 to v33 [#250](https://github.com/google/triage-party/pull/250)
* Default to earliest due milestone [#245](https://github.com/google/triage-party/pull/245)
* Return false if matchDuration is passed an empty t [#244](https://github.com/google/triage-party/pull/244)
* Implement provider.GetClosedIssues [#243](https://github.com/google/triage-party/pull/243)

Thank you for this release towards our contributors:

- Thomas Strömberg
- dw1
- gsquared94

## Version 1.4.0-beta.1 - 2021-02-22

Changes:

* Use name parameter from settings file if present [#240](https://github.com/google/triage-party/pull/240)
* Resolve GHE host on getting git provider [#232](https://github.com/google/triage-party/pull/232)
* GitHub workflows: Update actions/setup to v2 [#239](https://github.com/google/triage-party/pull/239)
* Refactor provider library with Go best practices [#238](https://github.com/google/triage-party/pull/238)
* Add static checks [#218](https://github.com/google/triage-party/pull/218)
* Validate configured repos [#234](https://github.com/google/triage-party/pull/234)
* Avoid panics due to Github and Gitlab API error results [#237](https://github.com/google/triage-party/pull/237)
* Prevent endless loop of loading same page [#214](https://github.com/google/triage-party/pull/214)
* minikube: improve daily/weekly ordering/rules [#230](https://github.com/google/triage-party/pull/230)
* Run gofumports -w across code base [#225](https://github.com/google/triage-party/pull/225)
* Convert dupe rule panic to error message [#224](https://github.com/google/triage-party/pull/224)
* Update base version to v1.4.0-DEV [#223](https://github.com/google/triage-party/pull/223)
* Improve gathering data message [#222](https://github.com/google/triage-party/pull/222)
* Update Kubernetes support label to kind/support [#221](https://github.com/google/triage-party/pull/221)
* Increase reload wait time to 5 seconds [#220](https://github.com/google/triage-party/pull/220)
* Wait 4 seconds to refresh [#219](https://github.com/google/triage-party/pull/219)
* Improve "Try it" section in README.md [#217](https://github.com/google/triage-party/pull/217)
* Build application on CI [#216](https://github.com/google/triage-party/pull/216)
* Add gitlab support [#193](https://github.com/google/triage-party/pull/193)

Huge thank you for this release towards our contributors:

- Aaron Ogle
- Brian de Alwis
- Jim Kalafut
- Kamil Breguła
- Teplov
- Thomas Strömberg
- disktnk

## Version 1.3.0 - 2020-08-24

* Defer comment tags until all data is fetched, use cached metadata when available [#207](https://github.com/google/triage-party/pull/207)
* Make the Tags a set to avoid issues of DeDup [#203](https://github.com/google/triage-party/pull/203)
* Use internal mtime tracking to determine in-memory issue validatity  [#205](https://github.com/google/triage-party/pull/205)
* Add svg for future scaling image uses [#202](https://github.com/google/triage-party/pull/202)
* fix startup panic [#199](https://github.com/google/triage-party/pull/199)
* cursor:pointer for collapsible [#198](https://github.com/google/triage-party/pull/198)
* Allow collections to define their own velocity stats [#191](https://github.com/google/triage-party/pull/191)

Huge thank you for this release towards our contributors:

- Grant Mccloskey
- Thomas Strömberg
- balopat

## Version 1.2.1 - 2020-07-17

* Return stale results if GitHub cannot be queried [#189](https://github.com/google/triage-party/pull/189)
* Fix Kanban ETA estimation, add ETA for non-milestone pages [#188](https://github.com/google/triage-party/pull/188)
* Automatically pick a contrasting label text color [#187](https://github.com/google/triage-party/pull/187)

## Version 1.2.0 - 2020-07-14

* Don't block page-loads if missing content, add healthz [#175](https://github.com/google/triage-party/pull/175)
* Add --warn-age flag instead of determining automatically [#184](https://github.com/google/triage-party/pull/184)
* Persist items on the fly rather than periodically in bulk [#183](https://github.com/google/triage-party/pull/183)
* Base data age on oldest query time rather than data age [#181](https://github.com/google/triage-party/pull/181)
* optimization: use cached conversations instead of re-parsing [#178](https://github.com/google/triage-party/pull/178)
* Add status to /healthz, build similarity tables in the background [#177](https://github.com/google/triage-party/pull/177)
* Add /threadz handler, fix data age calculation bug [#176](https://github.com/google/triage-party/pull/176)
* Add comment cross-reference parsing, support multiple debug numbers [#169](https://github.com/google/triage-party/pull/169)
* Fetch timelines for all issues within a milestone [#166](https://github.com/google/triage-party/pull/166)
* Fetch timelines for issues that have zero comments [#165](https://github.com/google/triage-party/pull/165)
* make comment fetching optional for the initial data cycle [#162](https://github.com/google/triage-party/pull/162)

## Version 1.2.0-beta.3 - 2020-06-19

* Allow stale comment/timeline/review data on initial cycle [#155](https://github.com/google/triage-party/pull/155)
* Delay persistence until second cycle [#153](https://github.com/google/triage-party/pull/153)
* stale warning: Add link to Shift-Reload documentation  [#158](https://github.com/google/triage-party/pull/158)
* Remove obsolete tag warning [#157](https://github.com/google/triage-party/pull/157)
* Fix filtered view count, remove obsolete embedded mode artifacts [#156](https://github.com/google/triage-party/pull/156)
* Golangci linting cleanup [#150](https://github.com/google/triage-party/pull/150)

Thank you to our contributors for this release:

- Ken Sipe
- Thomas Strömberg

## Version 1.2.0-beta.2 - 2020-06-17

Multiple improvements to the new Kanban display feature:

* Improve reaction count display [#147](https://github.com/google/triage-party/pull/147)
* Improve how Kanban is handled for unconfigured collections [#146](https://github.com/google/triage-party/pull/146)
* Improve Kanban dashboard milestone handling & UI [#142](https://github.com/google/triage-party/pull/142)
* Make timeline cache date calculation smarter [#141](https://github.com/google/triage-party/pull/141)
* site: Make relative times more specific [#137](https://github.com/google/triage-party/pull/137)
* Ensure that referenced PR's are the same age of parent issue [#136](https://github.com/google/triage-party/pull/136)

## Version 1.2.0-beta.1 - 2020-06-10

* Add Kanban visualization support (display: kanban) [#125](https://github.com/google/triage-party/pull/125)
* Improve PR review tags, bump to v1.2.0-beta.1 [#130](https://github.com/google/triage-party/pull/130)
* Respect 'dedup: false' collection state [#123](https://github.com/google/triage-party/pull/123)
* Add MaxSaveAge/MaxLoadAge to avoid persisting stale data [#126](https://github.com/google/triage-party/pull/126)
* Display statuses of similar conversations [#124](https://github.com/google/triage-party/pull/124)
* Clarify recv-q, make it easier to debug [#120](https://github.com/google/triage-party/pull/120)
* Include issue-like comments for PullRequests [#119](https://github.com/google/triage-party/pull/119)

Thanks to:

* Michael Plump
* Shingo Omura
* Thomas Stromberg

## Version 1.1.0 - 2020-05-19

* Change Dockerfile to build source and make build-args optional [#110](https://github.com/google/triage-party/pull/110)
* Fix 'tag: assigned' filter [#109](https://github.com/google/triage-party/pull/109)
* Make members and member-roles configurable [#108](https://github.com/google/triage-party/pull/108)
* Remove org membership fetching (no longer necessary) [#107](https://github.com/google/triage-party/pull/107)

Thank you to our contributors:

* Shingo Omura
* Thomas Stromberg

## Version 1.0.0 - 2020-05-13

No functional changes since beta 4:

* Remove reference to unused moment library [#97](https://github.com/google/triage-party/pull/97)

Special thanks to all the Kubernetes users and contributors who opened issues and shared their feedback before the initial release!

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
