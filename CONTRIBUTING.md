<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [How to Contribute](#how-to-contribute)
  - [Contributor License Agreement](#contributor-license-agreement)
  - [Code reviews](#code-reviews)
  - [Community Guidelines](#community-guidelines)
  - [Static checks](#static-checks)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# How to Contribute

We'd love to accept your patches and contributions to this project. There are
just a few small guidelines you need to follow.

## Contributor License Agreement

Contributions to this project must be accompanied by a Contributor License
Agreement. You (or your employer) retain the copyright to your contribution;
this simply gives us permission to use and redistribute your contributions as
part of the project. Head over to <https://cla.developers.google.com/> to see
your current agreements on file or to sign a new one.

You generally only need to submit a CLA once, so if you've already submitted one
(even if it was for a different project), you probably don't need to do it
again.

## Code reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult
[GitHub Help](https://help.github.com/articles/about-pull-requests/) for more
information on using pull requests.

## Community Guidelines

This project follows [Google's Open Source Community
Guidelines](https://opensource.google.com/conduct/).

## Static checks

This project is using [pre-commits](https://pre-commit.com) to ensure the quality of the code.
We encourage you to use pre-commits, but it's not a required to contribute. Every change is checked
on CI and if it does not pass the tests it cannot be accepted. If you want to check locally then
you should install Python3.6 or newer together and run:
```bash
pip install pre-commit
# or
brew install pre-commit
```
For more installation options visit the [pre-commits](https://pre-commit.com).

To turn on pre-commit checks for commit operations in git, run:
```bash
pre-commit install
```
To run all checks on your staged files, run:
```bash
pre-commit run
```
To run all checks on all files, run:
```bash
pre-commit run --all-files
```
