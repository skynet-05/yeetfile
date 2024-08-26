package globals

import (
	"log"
	"yeetfile/cli/api"
	"yeetfile/cli/config"
	"yeetfile/cli/crypto"
)

var API *api.Context
var Config *config.Config

func init() {
	Config = config.InitConfig()
	session := Config.ReadSession()

	if session == nil || len(session) == 0 {
		API = api.InitContext(Config.Server, "")
		return
	}

	cliKey := crypto.ReadCLIKey()
	if cliKey == nil || len(cliKey) == 0 {
		log.Println("missing YEETFILE_CLI_KEY to decrypt session")
		API = api.InitContext(Config.Server, "")
	} else {
		sessionVal, err := crypto.DecryptChunk(cliKey, session)
		if err != nil {
			API = api.InitContext(Config.Server, "")
		} else {
			API = api.InitContext(Config.Server, string(sessionVal))
		}
	}

}
