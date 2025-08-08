module github.com/clarify/clarify-go/devdata_cli

go 1.23.0

toolchain go1.24.5

require (
	github.com/clarify/clarify-go v0.4.0
	github.com/peterbourgon/ff/v3 v3.4.0
)

require golang.org/x/oauth2 v0.30.0 // indirect

replace github.com/clarify/clarify-go v0.4.0 => ../../
