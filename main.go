package main

import (
	"fmt"
	"io"
	"net/http"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/variables"
	"github.com/rajatjindal/spin-plugin-release-bot/pkg/releaser"
	"github.com/sirupsen/logrus"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("starting spin-plugin-release-bot handler %s", string(raw))
		token, err := variables.Get("gh_token")
		if err != nil {
			http.Error(w, fmt.Sprintf("internal server error %v", err), http.StatusInternalServerError)
			return
		}

		rh := releaser.New(r.Context(), token)
		rh.HandleActionWebhook(w, r)
		logrus.Info("done spin-plugin-release-bot handler")
	})
}

func main() {}
