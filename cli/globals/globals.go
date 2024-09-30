package globals

import (
	"log"
	"yeetfile/cli/api"
	"yeetfile/cli/config"
	"yeetfile/cli/crypto"
	"yeetfile/shared"
)

var API *api.Context
var Config *config.Config
var ServerInfo shared.ServerInfo

func init() {
	Config = config.LoadConfig()

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
			log.Println("failed to decrypt session with YEETFILE_CLI_KEY value")
			API = api.InitContext(Config.Server, "")
		} else {
			API = api.InitContext(Config.Server, string(sessionVal))
		}
	}

	var err error
	ServerInfo, err = API.GetServerInfo()
	if err != nil {
		log.Fatalf("Error fetching server info: %v\n", err)
	}
}
