# devdata_cli

This example implements a command-based CLI with multiple sub-commands for generating or consuming example data for testing purposes. Each command can be configured either through:

- Command-line flags (`-log-only`, `-credentials clarify-credentials.json`).
- Environment variables (`DEVDATA_LOG_ONLY=true`, `DEVDATA_CREDENTIALS=clarify-credentials.json`).
- A config file with one config parameter per line (`log-only`, `clarify-credentials clarify-credentials.json`).

## Getting started

```sh
go run . -help
go run . <COMMAND> -help
```

Example (bash shell):

```sh
go run . -credentials "$HOME/.secrets/clarify-credentials.json" save-signals -s 100 -n 5
```

Is equivalent to:

```sh
export DEVDATA_CREDENTIALS="$HOME/.secrets/clarify-credentials.json"
go run . save-signals -s 100 -n 5
```
