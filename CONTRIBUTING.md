# Contributing

This document is mainly aimed at audience within Searis AS, and describes information about how to contribute to this repository.

## Configure git

After first check out, run the following command to use our recommended git hooks:

```sh
git config core.hooksPath .githooks
```

Also ensure that the git author configuration correctly, using your _work email_, or an email (if you contribute on behalf of a company). See the [Git set-up guide](https://docs.github.com/en/get-started/quickstart/set-up-git#setting-up-git) from GitHub for details. We also recommend that you set up [commit verification](https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification).

## Updating copyright

Whenever updating a file, ensure that the Copyright notice is up-to-date. I.e. if the notice says "Copyright 2022 Searis AS", and the current year is "2028", then update the notice to "Copyright 2022-2028 Searis AS".

If you are not making the contribution on behalf of Searis AS, you should add a new copyright line instead:

```
// Copyright 2028 <Your company> <contact email for legal issues>
// Copyright 2020 Searis AS
```

The copyright notice containing the current year always needs to be on top.

## Atomic commits

Avoid unrelated changes within a commit, and retain linear history. This requires good knowledge of the Git rebase feature.

## Commit message

Let your commit message follow this format:

```txt
<header line preferably less than 50 chars>
<BLANK LINE>
<Body text, wrap at 72 chars.>
```

With more details following below.

### Commit header format

```txt
<component>: <lower case change summary>
```

Use imperative mode in the header change summary. It should not capitalize the first letter and it should not include a trailing period. This means that if the first word does not naturally start with a large letter (e.g. `change`), the summary should start with a lower-case. Words that always start with an upper-cas (e.g `ID` or `MyClass`), should keep their case, also when used as the first word.

When relevant, the header should also include a _component_ prefix. Generally, a relevant portion of the directory path that's changed is a good candidate as a component name, but make sure it's not too long or you will have less space left to describe your changes. Abstract terms like `docs:` or `ci:` are accepted, but path based alternatives such as `README.md:` and `.github:` are preferred.

### Commit body format

The commit message body should consist of full sentences, including capitalized letters and periods. It may also contain lists.

```txt
<Message containing full sentences and potentially lists.>
```
