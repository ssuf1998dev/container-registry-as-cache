package configfile

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
)

type ConfigAuth struct {
	Username string `json:"usename,omitempty"`
	Password string `json:"password,omitempty"`
}

type Config struct {
	Auths map[string]ConfigAuth `json:"auths,omitempty"`
}

type ConfigFile struct {
	Dir    string
	File   string
	Config *Config
}

func NewConfigFile() (*ConfigFile, error) {
	dir := appdir.New(utils.Crac).UserConfig()
	cf := &ConfigFile{
		Dir:  dir,
		File: filepath.Join(dir, "config.json"),
	}
	if err := cf.ReadConfig(); err != nil {
		return nil, err
	}
	return cf, nil
}

func (cf *ConfigFile) ReadConfig() error {
	if _, err := os.Stat(cf.Dir); err != nil {
		if err := os.MkdirAll(cf.Dir, 0755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(cf.File, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	if len(b) == 0 {
		config := &Config{Auths: map[string]ConfigAuth{}}
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		if err != nil {
			return err
		}
		cf.Config = config
		return nil
	}

	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return err
	}
	cf.Config = &config
	return nil
}

func (cf *ConfigFile) WriteConfig() error {
	err := cf.ReadConfig()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(cf.Config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cf.File, b, 0644)
}
