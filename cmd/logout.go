package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/lade-io/go-lade"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of your Lade account",
	RunE: func(cmd *cobra.Command, args []string) error {
		oauthConf := getOAuthConfig()
		return logoutRun(oauthConf)
	},
}

func logoutRun(oauthConf *oauth2.Config) error {
	if conf.RefreshToken == "" {
		fmt.Println("Already logged out")
		return nil
	}
	v := url.Values{}
	v.Set("token", conf.RefreshToken)
	v.Set("token_type_hint", "refresh_token")
	revokeURL := strings.Replace(oauthConf.Endpoint.TokenURL, "token", "revoke", 1)
	req, err := http.NewRequest("POST", revokeURL, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(lade.DefaultClientID, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return lade.ErrServerError
	}
	fmt.Println("Logout successful")
	return conf.StoreToken(new(oauth2.Token))
}
