package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/jinzhu/configor"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AccessToken  string    `yaml:"access_token,omitempty" env:"LADE_ACCESS_TOKEN"`
	RefreshToken string    `yaml:"refresh_token,omitempty" env:"LADE_REFRESH_TOKEN"`
	Expiry       time.Time `yaml:"expiry,omitempty" env:"LADE_EXPIRY"`
	APIURL       string    `yaml:"api_url,omitempty" env:"LADE_API_URL"`
	AuthURL      string    `yaml:"auth_url,omitempty" env:"LADE_AUTH_URL"`
	TokenURL     string    `yaml:"token_url,omitempty" env:"LADE_TOKEN_URL"`
}

func (c *Config) StoreToken(token *oauth2.Token) error {
	c.AccessToken = token.AccessToken
	c.RefreshToken = token.RefreshToken
	c.Expiry = token.Expiry
	return writeConfig(c)
}

const (
	configDirName  = ".lade"
	configFileName = "config.yaml"
)

func init() {
	os.Setenv("CONFIGOR_ENV_PREFIX", "LADE")
}

func Load(conf *Config) (err error) {
	_, configFile, err := configPaths()
	if err != nil {
		return err
	}
	if _, err = os.Stat(configFile); err == nil {
		err = configor.Load(conf, configFile)
	} else {
		err = configor.Load(conf)
	}
	if err != nil {
		return fmt.Errorf("Config error: %s", err)
	}
	return nil
}

func configPaths() (string, string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", "", err
	}
	configDir := filepath.Join(home, configDirName)
	configFile := filepath.Join(configDir, configFileName)
	return configDir, configFile, nil
}

func writeConfig(c *Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	configDir, configFile, err := configPaths()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, data, 0600)
}
