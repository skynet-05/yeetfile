package config

import (
	_ "embed"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"yeetfile/cli/utils"
)

type Paths struct {
	config    string
	gitignore string
	session   string
}

type Config struct {
	Server  string
	session string
}

var baseConfigPath = filepath.Join(".config", "yeetfile")

const configFileName = "config.yml"
const gitignoreName = ".gitignore"
const sessionName = "session"

//go:embed config.yml
var defaultConfig string

//go:embed .gitignore
var defaultGitignore string

// SetupConfigDir ensures that the directory necessary for yeetfile's config
// have been created. This path defaults to $HOME/.config/yeetfile.
func SetupConfigDir() (Paths, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	localConfig := filepath.Join(dirname, baseConfigPath)
	err = os.MkdirAll(localConfig, os.ModePerm)
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		config:    filepath.Join(localConfig, configFileName),
		gitignore: filepath.Join(localConfig, gitignoreName),
		session:   filepath.Join(localConfig, sessionName),
	}, nil
}

// ReadConfig reads the config file (config.yml) for current configuration
func ReadConfig(paths Paths) (Config, error) {
	if _, err := os.Stat(paths.config); err == nil {
		config := Config{}
		data, err := os.ReadFile(paths.config)
		if err != nil {
			return config, err
		}

		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return config, err
		}

		return config, nil
	} else {
		err := setupDefaultConfig(paths)
		if err != nil {
			return Config{}, err
		}
		return ReadConfig(paths)
	}
}

// setupDefaultConfig
func setupDefaultConfig(paths Paths) error {
	err := utils.CopyToFile(defaultConfig, paths.config)
	if err != nil {
		return err
	}

	err = utils.CopyToFile(defaultGitignore, paths.gitignore)
	if err != nil {
		return err
	}

	return nil
}

// SetSession sets the session to the value returned by the server when signing
// up or logging in, and saves it to a (gitignored) file in the config directory
func SetSession(config *Config, paths Paths, sessionVal string) error {
	err := utils.CopyToFile(sessionVal, paths.session)
	if err != nil {
		return err
	}

	config.session = sessionVal
	return nil
}
