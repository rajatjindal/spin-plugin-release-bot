module github.com/rajatjindal/spin-plugin-release-bot

go 1.25.5

require (
	github.com/bytecodealliance/wit-bindgen v0.0.0-00010101000000-000000000000 // indirect
	github.com/google/go-github/v66 v66.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.4-0.20230606125235-dd1b4c2e81af
	golang.org/x/oauth2 v0.6.0
)

replace (
	github.com/bytecodealliance/wit-bindgen => github.com/bytecodealliance/wit-bindgen/crates/go/src/package v0.51.0
	github.com/spinframework/spin-go-sdk/v3 => ../../spinframework/spin-go-sdk/v3
)

require github.com/spinframework/spin-go-sdk/v3 v3.0.0-00010101000000-000000000000

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
