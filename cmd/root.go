package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/kelproject/kel-go"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	flagURI string
)

// RootCmd is ...
var RootCmd = &cobra.Command{
	Use:   "kel",
	Short: "Kel end-user command-line tool",
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&flagURI, "uri", "", "", "URI for this invocation")
}

func getClusterAuthClient() *http.Client {
	oauth2.RegisterBrokenAuthHeaderProvider("https://identity.gondor.io/")
	conf := &oauth2.Config{
		ClientID: "KtcICiPMAII8FAeArUoDB97zmjqltllyUDev8HOS",
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://identity.gondor.io/oauth/token/",
		},
	}
	ctx := oauth2.NoContext
	var token *oauth2.Token
	var ok bool
	token, ok = config.Tokens["identity.gondor.io"]
	if !ok {
		var err error
		// ask for username
		var username string
		fmt.Printf("Username: ")
		fmt.Scan(&username)
		// ask for password safely
		var password string
		password, err = speakeasy.Ask("Password: ")
		if err != nil {
			fatal(err.Error())
		}
		token, err = conf.PasswordCredentialsToken(ctx, username, password)
		if err != nil {
			fatal(err.Error())
		}
		config.Tokens["identity.gondor.io"] = token
		config.Save()
	}
	return conf.Client(ctx, token)
}

func setupAuth() *http.Client {
	var hc *http.Client
	switch config.Auth {
	case AuthCluster:
		hc = getClusterAuthClient()
	case AuthNone:
		hc = http.DefaultClient
	}
	return hc
}

func setupKelClient(uri *URI) *kel.Client {
	parts := []string{}
	if uri.Insecure {
		parts = append(parts, "http://")
	} else {
		parts = append(parts, "https://")
	}
	parts = append(parts, uri.Host)
	parts = append(parts, "/v1/self")
	kc, err := kel.New(setupAuth(), strings.Join(parts, ""))
	if err != nil {
		fatal(err.Error())
	}
	return kc
}
