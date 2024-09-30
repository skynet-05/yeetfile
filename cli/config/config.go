package config

import (
	_ "embed"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"yeetfile/cli/utils"
)

type Paths struct {
	config        string
	gitignore     string
	session       string
	encPrivateKey string
	publicKey     string
}

type Config struct {
	Server      string `yaml:"server,omitempty"`
	DefaultView string `yaml:"default_view,omitempty"`
	Paths       Paths
}

var baseConfigPath = filepath.Join(".config", "yeetfile")

const configFileName = "config.yml"
const gitignoreName = ".gitignore"
const sessionName = "session"
const encPrivateKeyName = "enc-priv-key"
const publicKeyName = "pub-key"

//go:embed config.yml
var defaultConfig string

// setupConfigDir ensures that the directory necessary for yeetfile's config
// have been created. This path defaults to $HOME/.config/yeetfile.
func setupConfigDir() (Paths, error) {
	var localConfig string
	var configErr error
	if runtime.GOOS == "windows" {
		baseDir, err := os.UserConfigDir()
		if err != nil {
			return Paths{}, err
		}

		localConfig, configErr = makeConfigDirectories(baseDir, "yeetfile")
	} else {
		baseDir, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, err
		}

		localConfig, configErr = makeConfigDirectories(baseDir, baseConfigPath)
	}

	if configErr != nil {
		return Paths{}, configErr
	}

	return Paths{
		config:        filepath.Join(localConfig, configFileName),
		gitignore:     filepath.Join(localConfig, gitignoreName),
		session:       filepath.Join(localConfig, sessionName),
		encPrivateKey: filepath.Join(localConfig, encPrivateKeyName),
		publicKey:     filepath.Join(localConfig, publicKeyName),
	}, nil
}

// setupTempConfigDir creates a config directory for the current user in the
// OS's temporary directory. Used for testing.
func setupTempConfigDir() (Paths, error) {
	dirname := os.TempDir()
	localConfig, err := makeConfigDirectories(dirname, baseConfigPath)
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		config:        filepath.Join(localConfig, configFileName),
		gitignore:     filepath.Join(localConfig, gitignoreName),
		session:       filepath.Join(localConfig, sessionName),
		encPrivateKey: filepath.Join(localConfig, encPrivateKeyName),
		publicKey:     filepath.Join(localConfig, publicKeyName),
	}, nil
}

// makeConfigDirectories creates the necessary directories for storing the
// user's local yeetfile config
func makeConfigDirectories(baseDir, configPath string) (string, error) {
	localConfig := filepath.Join(baseDir, configPath)
	err := os.MkdirAll(localConfig, os.ModePerm)
	if err != nil {
		return "", err
	}

	return localConfig, nil
}

// ReadConfig reads the config file (config.yml) for current configuration
func ReadConfig(p Paths) (Config, error) {
	if _, err := os.Stat(p.config); err == nil {
		config := Config{Paths: p}
		data, err := os.ReadFile(p.config)
		if err != nil {
			return config, err
		}

		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return config, err
		}

		// Strip trailing slash
		if strings.HasSuffix(config.Server, "/") {
			config.Server = config.Server[0 : len(config.Server)-1]
		}

		return config, nil
	} else {
		err = setupDefaultConfig(p)
		if err != nil {
			return Config{}, err
		}
		return ReadConfig(p)
	}
}

// setupDefaultConfig copies default config files from the repo to the user's
// config directory
func setupDefaultConfig(p Paths) error {
	err := utils.CopyToFile(defaultConfig, p.config)
	if err != nil {
		return err
	}

	defaultGitignore := fmt.Sprintf(`
%s
%s
%s`, sessionName, encPrivateKeyName, publicKeyName)

	err = utils.CopyToFile(defaultGitignore, p.gitignore)
	if err != nil {
		return err
	}

	err = utils.CopyToFile("", p.session)
	if err != nil {
		return err
	}

	return nil
}

// SetSession sets the session to the value returned by the server when signing
// up or logging in, and saves it to a (gitignored) file in the config directory
func (c Config) SetSession(sessionVal string) error {
	err := utils.CopyToFile(sessionVal, c.Paths.session)
	if err != nil {
		return err
	}

	return nil
}

// ReadSession reads the value in $config_path/session
func (c Config) ReadSession() []byte {
	if _, err := os.Stat(c.Paths.session); err == nil {
		session, err := os.ReadFile(c.Paths.session)
		if err != nil {
			return nil
		}

		return session
	} else {
		return nil
	}
}

func (c Config) Reset() error {
	if _, err := os.Stat(c.Paths.session); err == nil {
		err := os.Remove(c.Paths.session)
		if err != nil {
			log.Println("error removing session file")
			return err
		}
	}

	if _, err := os.Stat(c.Paths.encPrivateKey); err == nil {
		err = os.Remove(c.Paths.encPrivateKey)
		if err != nil {
			log.Println("error removing private key")
			return err
		}
	}

	if _, err := os.Stat(c.Paths.publicKey); err == nil {
		err = os.Remove(c.Paths.publicKey)
		if err != nil {
			log.Println("error removing public key")
			return err
		}
	}

	return nil
}

// SetKeys writes the encrypted private key bytes and the (unencrypted) public
// key bytes to their respective file paths
func (c Config) SetKeys(encPrivateKey, publicKey []byte) error {
	err := utils.CopyBytesToFile(encPrivateKey, c.Paths.encPrivateKey)
	if err != nil {
		return err
	}

	err = utils.CopyBytesToFile(publicKey, c.Paths.publicKey)
	return err
}

// GetKeys returns the user's encrypted private key and their public key from
// the config directory. Returns private key, public key, and error.
func (c Config) GetKeys() ([]byte, []byte, error) {
	var privateKey []byte
	var publicKey []byte

	_, privKeyErr := os.Stat(c.Paths.encPrivateKey)
	_, pubKeyErr := os.Stat(c.Paths.publicKey)

	if privKeyErr != nil || pubKeyErr != nil {
		return nil, nil, errors.New("key files do not exist in config dir")
	}

	privateKey, privKeyErr = os.ReadFile(c.Paths.encPrivateKey)
	publicKey, pubKeyErr = os.ReadFile(c.Paths.publicKey)

	if privKeyErr != nil || pubKeyErr != nil {
		errMsg := fmt.Sprintf("error reading key files:\n"+
			"privkey: %v\n"+
			"pubkey: %v", privKeyErr, pubKeyErr)
		return nil, nil, errors.New(errMsg)
	}

	return privateKey, publicKey, nil
}

func LoadConfig() *Config {
	var err error

	// Setup config dir
	userConfigPaths, err := setupConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	userConfig, err := ReadConfig(userConfigPaths)
	if err != nil {
		log.Fatal(err)
	}

	return &userConfig
}
