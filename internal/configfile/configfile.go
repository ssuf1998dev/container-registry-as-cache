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
	dir    string
	file   string
	Config *Config
}

type NewOptions struct {
	File string
}

func NewConfigFile(opts *NewOptions) *ConfigFile {
	if opts == nil || len(opts.File) == 0 {
		dir := appdir.New(utils.Crac).UserConfig()
		return &ConfigFile{
			dir:  dir,
			file: filepath.Join(dir, "config.json"),
		}
	}
	return &ConfigFile{
		dir:  filepath.Dir(opts.File),
		file: opts.File,
	}
}

func (cf *ConfigFile) ready() error {
	if _, err := os.Stat(cf.dir); err != nil {
		if err := os.MkdirAll(cf.dir, 0755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(cf.file, os.O_CREATE|os.O_RDWR, 0644)
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
		return nil
	}
	return nil
}

func (cf *ConfigFile) Read() error {
	err := cf.ready()
	if err != nil {
		return err
	}

	b, err := os.ReadFile(cf.file)
	if err != nil {
		return err
	}

	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return err
	}
	cf.Config = &config
	return nil
}

func (cf *ConfigFile) Write() error {
	err := cf.ready()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(cf.Config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cf.file, b, 0644)
}

func (cf *ConfigFile) Reset() error {
	err := os.Remove(cf.file)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
