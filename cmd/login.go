package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lade-io/go-lade"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your Lade account",
	RunE: func(cmd *cobra.Command, args []string) error {
		oauthConf := getOAuthConfig()
		_, err := loginRun(oauthConf)
		return err
	},
}

type loginOpts struct {
	Username, Password string
}

func loginRun(oauthConf *oauth2.Config) (*oauth2.Token, error) {
	opts := &loginOpts{}
	ctx := context.Background()
	qs := []*survey.Question{
		{
			Name:     "username",
			Prompt:   &survey.Input{Message: "Username or email:"},
			Validate: survey.Required,
		},
		{
			Name:     "password",
			Prompt:   &survey.Password{Message: "Password:"},
			Validate: survey.Required,
		},
	}
	for {
		fmt.Println("Enter your Lade credentials:")
		if err := survey.Ask(qs, opts); err != nil {
			return nil, askError(err)
		}
		token, err := oauthConf.PasswordCredentialsToken(ctx, opts.Username, opts.Password)
		if err != nil {
			if oautherr, ok := err.(*oauth2.RetrieveError); ok {
				apierr := &lade.APIError{}
				if err = json.Unmarshal(oautherr.Body, apierr); err == nil {
					switch apierr.Type {
					case "invalid_grant":
						fmt.Println("Invalid username or password. Please try again.")
						continue
					case "email_not_verified":
						fmt.Println("Your email is not verified. Please check your inbox.")
						continue
					}
				}
			}
			return nil, lade.ErrServerError
		}
		fmt.Println("Logged in as " + opts.Username)
		return token, conf.StoreToken(token)
	}
}
