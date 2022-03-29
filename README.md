# Clarify Go SDK

[![GitHub Actions](https://github.com/clarify/clarify-go/workflows/Go/badge.svg?branch=master)](https://github.com/clarify/clarify-go/actions?query=workflow%3AGo+branch%main)
[![Go Reference](https://pkg.go.dev/badge/github.com/clarify/clarify-go.svg)](https://pkg.go.dev/github.com/clarify/clarify-go)
[![Clarify Docs](https://img.shields.io/badge/%7CC%7C-docs-blue)][docs]

This repository contains a Go(lang) SDK from [Clarify][clarify]. Clarify is a cloud service for active collaboration on time-series data. It includes a free plan to get started, a plan for small businesses as well as an Enterprise plan for our most demanding customers. If you are interested in SDKs for a different language, or perhaps a different way of integration altogether, please check our [documentation][docs].

## Risk of breaking changes

This SDK is following [semantic versioning][semver]. The SDK is currently in a v0 state, which means that we reserver the right to introduce breaking changes in MINOR releases.

## Features

This SDK is currently written for [Clarify API v1][docs-v1], and include the following features:

- Compose Clarify data frames using the `data` sub-package.
- Write signal meta-data to Clarify with `client.SaveSignals`. See [example](examples/save_signals/).
- Write data frames to Clarify with `client.Insert`. See [example](examples/save_signals/).

[clarify]: https://clarify.io/
[semver]: https://semver.org/
[docs]: https://docs.clarify.io
[docs-v1]: https://docs.clarify.io/api/1.0/release-notes
