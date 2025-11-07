package configfile

import (
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
)

type ConfigAuth struct {
	Username string `yaml:"usename,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Config struct {
	Auths map[string]ConfigAuth `yaml:"auths,omitempty"`
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
			file: filepath.Join(dir, "config.yaml"),
		}
	}
	return &ConfigFile{
		dir:  filepath.Dir(opts.File),
		file: opts.File,
	}
}

func (cf *ConfigFile) ready() error {
	if _, err := os.Stat(cf.dir); err != nil {
		if err := os.MkdirAll(cf.dir, 0766); err != nil {
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
		b, err := yaml.Marshal(config)
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
	err = yaml.Unmarshal(b, &config)
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

	b, err := yaml.Marshal(cf.Config)
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
