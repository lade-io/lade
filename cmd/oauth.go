package cmd

import (
	"errors"
	"syscall"

	"github.com/lade-io/go-lade"
	"golang.org/x/oauth2"
)

func getOAuthConfig() *oauth2.Config {
	oauth := &oauth2.Config{
		ClientID: lade.DefaultClientID,
		Scopes:   lade.DefaultScopes,
		Endpoint: lade.Endpoint,
	}
	if conf.AuthURL != "" {
		oauth.Endpoint.AuthURL = conf.AuthURL
	}
	if conf.TokenURL != "" {
		oauth.Endpoint.TokenURL = conf.TokenURL
	}
	return oauth
}

func getClient() (*lade.Client, error) {
	token := &oauth2.Token{
		AccessToken:  conf.AccessToken,
		RefreshToken: conf.RefreshToken,
		Expiry:       conf.Expiry,
	}
	ctx := oauth2.NoContext
	oauthConf := getOAuthConfig()
	tokenSource := oauthConf.TokenSource(ctx, token)
	if !token.Valid() {
		token, err := tokenSource.Token()
		if err == nil {
			conf.StoreToken(token)
		} else if !errors.Is(err, syscall.ECONNREFUSED) {
			token, err = loginRun(oauthConf)
			if err != nil {
				return nil, err
			}
			tokenSource = oauthConf.TokenSource(ctx, token)
		} else {
			return nil, err
		}
	}
	httpClient := oauth2.NewClient(ctx, tokenSource)
	client := lade.NewClient(httpClient)
	if conf.APIURL != "" {
		client.SetAPIURL(conf.APIURL)
	}
	return client, nil
}
