package cmd

import (
	"errors"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/lade-io/go-lade"
	"github.com/lade-io/lade/config"
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

type tokenSource struct {
	*sync.Mutex
	oauthConf *oauth2.Config
	token     *oauth2.Token
}

func newTokenSource(oauthConf *oauth2.Config) *tokenSource {
	mutex := new(sync.Mutex)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stopChan
		mutex.Lock()
		os.Exit(0)
	}()
	return &tokenSource{
		Mutex:     mutex,
		oauthConf: oauthConf,
		token:     conf.GetToken(),
	}
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	t.Lock()
	defer t.Unlock()
	if t.token.Valid() {
		return t.token, nil
	}
	if err := config.Load(conf); err != nil {
		return nil, err
	}
	t.token = conf.GetToken()
	if t.token.Valid() {
		return t.token, nil
	}
	ts := t.oauthConf.TokenSource(oauth2.NoContext, t.token)
	token, err := ts.Token()
	if err != nil {
		return nil, err
	}
	t.token = token
	return token, conf.StoreToken(token)
}

func getClient() (*lade.Client, error) {
	oauthConf := getOAuthConfig()
	ts := newTokenSource(oauthConf)
	if !ts.token.Valid() {
		_, err := ts.Token()
		var e *oauth2.RetrieveError
		if errors.As(err, &e) && e.Response.StatusCode >= 500 {
			lines := strings.SplitN(e.Error(), "\n", 2)
			return nil, errors.New(lines[0])
		} else if errors.Is(err, syscall.ECONNREFUSED) {
			return nil, err
		} else if err != nil {
			ts.token, err = loginRun(oauthConf)
			if err != nil {
				return nil, err
			}
		}
	}
	httpClient := oauth2.NewClient(oauth2.NoContext, ts)
	client := lade.NewClient(httpClient)
	if conf.APIURL != "" {
		client.SetAPIURL(conf.APIURL)
	}
	return client, nil
}
