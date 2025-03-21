package main

import (
	"fmt"
	"net/http"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
	"github.com/fermyon/spin-go-sdk/variables"
	"github.com/rajatjindal/spin-plugin-release-bot/pkg/releaser"
	"github.com/sirupsen/logrus"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		// raw, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	logrus.Error(err)
		// 	http.Error(w, fmt.Sprintf("internal server error %v", err), http.StatusInternalServerError)
		// 	return
		// }

		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("starting spin-plugin-release-bot handler")
		token, err := variables.Get("gh_token")
		if err != nil {
			logrus.Error(err)
			http.Error(w, fmt.Sprintf("internal server error %v", err), http.StatusInternalServerError)
			return
		}

		rh, err := releaser.New(r.Context(), token)
		if err != nil {
			logrus.Errorf("failed to create new release object: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rh.HandleActionWebhook(w, r)
		logrus.Info("done spin-plugin-release-bot handler")
	})
}

func main() {}
