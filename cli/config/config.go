package config

import (
	_ "embed"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
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
}

var UserConfig Config
var UserConfigPaths Paths
var Session string

var baseConfigPath = filepath.Join(".config", "yeetfile")

const configFileName = "config.yml"
const gitignoreName = ".gitignore"
const sessionName = "session"
const encPrivateKeyName = "enc-priv-key"
const publicKeyName = "pub-key"

//go:embed config.yml
var defaultConfig string

// SetupConfigDir ensures that the directory necessary for yeetfile's config
// have been created. This path defaults to $HOME/.config/yeetfile.
func SetupConfigDir() (Paths, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	localConfig, err := makeConfigDirectories(dirname)
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

// setupTempConfigDir creates a config directory for the current user in the
// OS's temporary directory. Used for testing.
func setupTempConfigDir() (Paths, error) {
	dirname := os.TempDir()
	localConfig, err := makeConfigDirectories(dirname)
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
func makeConfigDirectories(dirname string) (string, error) {
	localConfig := filepath.Join(dirname, baseConfigPath)
	err := os.MkdirAll(localConfig, os.ModePerm)
	if err != nil {
		return "", err
	}

	return localConfig, nil
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

		// Strip trailing slash
		if strings.HasSuffix(config.Server, "/") {
			config.Server = config.Server[0 : len(config.Server)-1]
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

// setupDefaultConfig copies default config files from the repo to the user's
// config directory
func setupDefaultConfig(paths Paths) error {
	err := utils.CopyToFile(defaultConfig, paths.config)
	if err != nil {
		return err
	}

	defaultGitignore := fmt.Sprintf(`
%s
%s
%s`, sessionName, encPrivateKeyName, publicKeyName)

	err = utils.CopyToFile(defaultGitignore, paths.gitignore)
	if err != nil {
		return err
	}

	err = utils.CopyToFile("", paths.session)
	if err != nil {
		return err
	}

	return nil
}

// SetSession sets the session to the value returned by the server when signing
// up or logging in, and saves it to a (gitignored) file in the config directory
func (paths Paths) SetSession(sessionVal string) error {
	Session = sessionVal
	err := utils.CopyToFile(sessionVal, paths.session)
	if err != nil {
		return err
	}

	return nil
}

// ReadSession reads the value in $config_path/session
func (paths Paths) ReadSession() string {
	if _, err := os.Stat(paths.session); err == nil {
		session, err := os.ReadFile(paths.session)
		if err != nil {
			return ""
		}

		return string(session)
	} else {
		return ""
	}
}

func (paths Paths) Reset() error {
	if _, err := os.Stat(paths.session); err == nil {
		err := os.Remove(paths.session)
		if err != nil {
			log.Println("error removing session file")
			return err
		}
	}

	if _, err := os.Stat(paths.encPrivateKey); err == nil {
		err = os.Remove(paths.encPrivateKey)
		if err != nil {
			log.Println("error removing private key")
			return err
		}
	}

	if _, err := os.Stat(paths.publicKey); err == nil {
		err = os.Remove(paths.publicKey)
		if err != nil {
			log.Println("error removing public key")
			return err
		}
	}

	return nil
}

// SetKeys writes the encrypted private key bytes and the (unencrypted) public
// key bytes to their respective file paths
func (paths Paths) SetKeys(encPrivateKey, publicKey []byte) error {
	err := utils.CopyBytesToFile(encPrivateKey, paths.encPrivateKey)
	if err != nil {
		return err
	}

	err = utils.CopyBytesToFile(publicKey, paths.publicKey)
	return err
}

// GetKeys returns the user's encrypted private key and their public key from
// the config directory. Returns private key, public key, and error.
func (paths Paths) GetKeys() ([]byte, []byte, error) {
	var privateKey []byte
	var publicKey []byte

	_, privKeyErr := os.Stat(paths.encPrivateKey)
	_, pubKeyErr := os.Stat(paths.publicKey)

	if privKeyErr != nil || pubKeyErr != nil {
		return nil, nil, errors.New("key files do not exist in config dir")
	}

	privateKey, privKeyErr = os.ReadFile(paths.encPrivateKey)
	publicKey, pubKeyErr = os.ReadFile(paths.publicKey)

	if privKeyErr != nil || pubKeyErr != nil {
		errMsg := fmt.Sprintf("error reading key files:\n"+
			"privkey: %v\n"+
			"pubkey: %v", privKeyErr, pubKeyErr)
		return nil, nil, errors.New(errMsg)
	}

	return privateKey, publicKey, nil
}

func init() {
	var err error

	// Setup config dir
	UserConfigPaths, err = SetupConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	UserConfig, err = ReadConfig(UserConfigPaths)
	if err != nil {
		log.Fatal(err)
	}

	Session = UserConfigPaths.ReadSession()
}
