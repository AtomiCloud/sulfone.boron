---
id: commit-conventions
title: Commit Conventions
---
This project uses [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) loosely as the specification
for our commits.

Commit message will be in the format:

```
type(scope): title

body
```

This page will document the types and scopes used.

# Types

| Type                  | Description                                                                                            |
| --------------------- | ------------------------------------------------------------------------------------------------------ |
| [build](#build)       | Changes that affect the build system or external dependencies (example scopes: nix)                    |
| [config](#config)     | Changes to the configuration files (example scopes: nix, ci)                                           |
| [ci](#ci)             | Changes to our CI configuration files and scripts (example scopes: github_workflow, github_action)     |
| [docs](#docs)         | Documentation only changes                                                                             |
| [feat](#feat)         | A new feature                                                                                          |
| [fix](#fix)           | A bug fix                                                                                              |
| [perf](#perf)         | A code change that improves performance                                                                |
| [refactor](#refactor) | A code change that neither fixes a bug nor adds a feature                                              |
| [style](#style)       | Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc) |
| [test](#test)         | Adding missing tests or correcting existing tests                                                      |
| [amend](#amend)       | Any amendment, uncategorized, or small mistakes (like typos)                                           |

## build

Changes that affect the build system or external dependencies (example scopes: nix)

| Scope   | Description  | Bump  |
| ------- | ------------ | ----- |
| default | Fixes in nix | `nil` |

## config

Changes to the configuration files (example scopes: nix, ci)

| Scope   | Description                | Bump  |
| ------- | -------------------------- | ----- |
| default | Update configuration files | `nil` |

## ci

Changes to our CI configuration files and scripts (example scopes: github_workflow, github_action)

| Scope   | Description             | Bump  |
| ------- | ----------------------- | ----- |
| default | Update CI configuration | `nil` |

## docs

Documentation only changes

| Scope   | Description                              | Bump  |
| ------- | ---------------------------------------- | ----- |
| default | Update documentation such as README file | `nil` |

## feat

A new feature

| **V.A.E**       | V.A.E values                                                                      |
| --------------- | --------------------------------------------------------------------------------- |
| verb            | add                                                                               |
| application     | when this commit is applied, it will _add_ `<scope>, <title>`                     |
| example         | feat: new withdraw api for rapid                                                  |
| example applied | when this commit is applied, it will _add_ `null`, **new withdraw api for rapid** |

| Scope   | Description            | Bump    |
| ------- | ---------------------- | ------- |
| default | Release a new features | `minor` |

## fix

A bug fix

| **V.A.E**       | V.A.E values                                                         |
| --------------- | -------------------------------------------------------------------- |
| verb            | fix                                                                  |
| application     | when this commit is applied, it will _fix_ `<title>`                 |
| example         | fix: deposit api for rapid                                           |
| example applied | when this commit is applied, it will _fix_ **deposit api for rapid** |

| Scope   | Description   | Bump    |
| ------- | ------------- | ------- |
| default | Generic fixes | `patch` |

## perf

A code change that improves performance

| **V.A.E**       | V.A.E values                                                                          |
| --------------- | ------------------------------------------------------------------------------------- |
| verb            | improve                                                                               |
| application     | when this commit is applied, it will _improve_ `<title>`                              |
| example         | perf: alpaca api callback speed increase                                              |
| example applied | when this commit is applied, it will _improve_ **alpaca api callback speed increase** |

| Scope   | Description                   | Bump    |
| ------- | ----------------------------- | ------- |
| default | Generic improvement statement | `patch` |

## refactor

A code change that neither fixes a bug nor adds a feature

| **V.A.E**       | V.A.E values                                                                                   |
| --------------- | ---------------------------------------------------------------------------------------------- |
| verb            | refactor                                                                                       |
| application     | when this commit is applied, it will _refactor_ `<scope>, <title>`                             |
| example         | refactor: make withdrawal code more reusable                                                   |
| example applied | when this commit is applied, it will _refactor_ `null`, **make withdrawal code more reusable** |

| Scope   | Description               | Bump    |
| ------- | ------------------------- | ------- |
| default | Refactor existing feature | `minor` |

## style

Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)

| **V.A.E**       | V.A.E values                                                            |
| --------------- | ----------------------------------------------------------------------- |
| verb            | style                                                                   |
| application     | when this commit is applied, it will _style_ `<title>`                  |
| example         | style: add missing semi colon                                           |
| example applied | when this commit is applied, it will _style_ **add missing semi colon** |

| Scope   | Description                   | Bump    |
| ------- | ----------------------------- | ------- |
| default | Generic improvement statement | `patch` |

## test

Adding missing tests or correcting existing tests

| **V.A.E**       | V.A.E values                                                                        |
| --------------- | ----------------------------------------------------------------------------------- |
| verb            | test                                                                                |
| application     | when this commit is applied, it will _test_ `<scope>, <title>`                      |
| example         | test: update test for deposit api                                                   |
| example applied | when this commit is applied, it will _test_ `null`, **update test for deposit api** |

| Scope   | Description          | Bump    |
| ------- | -------------------- | ------- |
| default | Made changes to test | `minor` |

## amend

Any amendment, uncategorized, or small mistakes (like typos)

| Scope   | Description | Bump  |
| ------- | ----------- | ----- |
| default | amendments  | `nil` |

# Special Scopes

| Scope        | Description                    | Bump  |
| ------------ | ------------------------------ | ----- |
| `no-release` | Prevent release from happening | `nil` |
