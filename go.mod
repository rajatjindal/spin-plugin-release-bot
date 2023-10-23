module github.com/rajatjindal/spin-plugin-release-bot

go 1.20

require (
	github.com/fermyon/spin/sdk/go v1.5.1
	github.com/google/go-github/v56 v56.0.0
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.6.0
)

replace github.com/fermyon/spin/sdk/go v1.5.1 => ../../fermyon/spin/sdk/go

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
