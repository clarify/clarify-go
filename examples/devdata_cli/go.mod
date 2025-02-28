module github.com/clarify/clarify-go/devdata_cli

go 1.23

toolchain go1.23.2

require (
	github.com/clarify/clarify-go v0.3.0
	github.com/peterbourgon/ff/v3 v3.4.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/oauth2 v0.13.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace github.com/clarify/clarify-go v0.3.0 => ../../
