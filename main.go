package main

import (
	"net/http"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/variables"
	"github.com/rajatjindal/spin-plugin-release-bot/pkg/releaser"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		token, err := variables.Get("gh_token")
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		rh := releaser.New(r.Context(), token)
		rh.HandleActionWebhook(w, r)
	})
}

func main() {}
