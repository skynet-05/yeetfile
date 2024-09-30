package api

import (
	"encoding/json"
	"log"
	"net/http"
	"yeetfile/cli/requests"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

// GetServerInfo returns information about the current YeetFile instance/server
func (ctx *Context) GetServerInfo() (shared.ServerInfo, error) {
	url := endpoints.ServerInfo.Format(ctx.Server)
	resp, err := requests.GetRequest(ctx.Session, url)
	if err != nil {
		return shared.ServerInfo{}, err
	} else if resp.StatusCode != http.StatusOK {
		return shared.ServerInfo{}, utils.ParseHTTPError(resp)
	}

	var serverInfo shared.ServerInfo
	err = json.NewDecoder(resp.Body).Decode(&serverInfo)
	if err != nil {
		log.Println("Error decoding server response: ", err)
		return shared.ServerInfo{}, err
	}

	return serverInfo, nil
}
