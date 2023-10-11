# Clarify Go SDK

[![GitHub Actions](https://github.com/clarify/clarify-go/workflows/Go/badge.svg?branch=master)](https://github.com/clarify/clarify-go/actions?query=workflow%3AGo+branch%main)
[![Go Reference](https://pkg.go.dev/badge/github.com/clarify/clarify-go.svg)](https://pkg.go.dev/github.com/clarify/clarify-go)
[![Clarify Docs](https://img.shields.io/badge/%7CC%7C-docs-blue)][docs]

This repository contains a Go(lang) SDK from [Clarify][clarify]. Clarify is a cloud service for active collaboration on time-series data. It includes a free plan to get started, a plan for small businesses as well as an Enterprise plan for our most demanding customers. If you are interested in SDKs for a different language, or perhaps a different way of integration altogether, please check our [documentation][docs].

## Risk of breaking changes

This SDK is following [Semantic Versioning][semver]. The SDK is currently in a v0 state, which means that we reserver the right to introduce breaking changes without increasing the MAJOR number.

## Features

This SDK is currently written for [Clarify API v1.1][docs-v1.1], and include the following features, based on your integration access configuration:

Always possible:

- Compose Clarify data frames using the `data` sub-package.
- Write signal meta-data to Clarify with `client.SaveSignals` (scoped to the current integration). See [examples/save_signals](examples/save_signals/).
- Write data frames to Clarify with `client.Insert` (scoped to the current integration). See [examples/insert](examples/insert/).

When access to the Admin namespace is granted in Clarify ` (scoped to entire organization):

- Read signal meta-data to Clarify with `client.Admin().SelectSignals`. Allows side-loading related items. See [examples/select_signals](examples/select_signals/).
- Publish signals as items directly with `client.Admin().PublishSignals`, or more conveniently via the `automation` package. See [examples/publish_signals](examples/publish_signals/) for the latter.

When access to the Clarify namespace is granted in Clarify (scoped to entire organization):

- Read item meta-data from Clarify via `client.Clarify().SelectItems`. See [examples/select_items](examples/select_items/).
- Read time-series data from Clarify via `client.Clarify().DataFrame`. See [examples/data_frame](examples/select_items/).
- Do calculations with data from Clarify via `client.Clarify().Evaluate`. See [examples/evaluate](examples/evaluate/).

[clarify]: https://clarify.io/
[semver]: https://semver.org/
[docs]: https://docs.clarify.io
[docs-v1.1]: https://docs.clarify.io/1.1

## Setting up automation routines

To quickly set-up your own automation routines, you can get started with our [automation template repository](https://github.com/clarify/template-clarify-automation). This template let's you customize and build your own automation binary and easily run it inside GitHub Actions; no external hosting environment is required (unless you want to).

## Copyright

Copyright 2022-2023 Searis AS

Licensed under the Apache License, Version 2.0 (the "License");
you may not use the content in this repo except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
