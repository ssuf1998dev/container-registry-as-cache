package api

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/gopasspw/gopass/pkg/appdir"
)

var configDir = appdir.New(Crac).UserConfig()
var configFile = filepath.Join(configDir, "config.json")

func readConfig() (*Config, error) {
	if _, err := os.Stat(configDir); err != nil {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, err
		}
	}

	f, err := os.OpenFile(configFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if len(b) == 0 {
		config := &Config{Auths: map[string]ConfigAuth{}}
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, err
		}
		_, err = f.Write(b)
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func writeConfig(config *Config) error {
	_, err := readConfig()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, b, 0644)
}

type ConfigAuth struct {
	Username string `json:"usename,omitempty"`
	Password string `json:"password,omitempty"`
}

type Config struct {
	Auths map[string]ConfigAuth `json:"auths,omitempty"`
}

func Login(opts ...Option) error {
	o := &options{
		context: context.Background(),
	}
	for _, option := range opts {
		option(o)
	}

	config, err := readConfig()
	if err != nil {
		return err
	}

	ref, err := name.ParseReference(o.repo)
	if err != nil {
		return err
	}
	key := ref.Context().RegistryStr()
	if config.Auths == nil {
		config.Auths = map[string]ConfigAuth{}
	}
	config.Auths[key] = ConfigAuth{
		Username: o.username,
		Password: o.password,
	}
	return writeConfig(config)
}

func Logout(opts ...Option) error {
	o := &options{
		context: context.Background(),
	}
	for _, option := range opts {
		option(o)
	}

	config, err := readConfig()
	if err != nil {
		return err
	}

	ref, err := name.ParseReference(o.repo)
	if err != nil {
		return err
	}
	key := ref.Context().RegistryStr()
	delete(config.Auths, key)

	return writeConfig(config)
}
