package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"yeetfile/cli/config"
	"yeetfile/cli/crypto"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

var ErrorAccountDoesNotExist = errors.New("account does not exist")

func LoginUser(loop bool) bool {
	identifier := utils.StringPrompt("Email or Account ID:")

	pw := utils.RequestPassword()
	userKey := crypto.GenerateUserKey([]byte(identifier), pw)
	loginKeyHash := crypto.GenerateLoginKeyHash(userKey, pw)

	err := sendLogin(shared.Login{
		Identifier:   identifier,
		LoginKeyHash: []byte(loginKeyHash),
	})

	var loginError error

	if err != nil {
		if errors.Is(err, ErrorAccountDoesNotExist) {
			fmt.Println("Error: Account does not exist or incorrect password")
			return LoginUser(loop)
		} else {
			loginError = errors.New("failed to log in")
		}
	}

	if loop && (err != nil || loginError != nil) {
		return LoginUser(loop)
	}

	fmt.Println("Successfully logged in!")

	return err == nil && loginError == nil
}

// sendLogin sends a POST request containing the shared.Login struct to log
// a user into the service.
func sendLogin(login shared.Login) error {
	reqData, err := json.Marshal(login)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}

	url := fmt.Sprintf("%s/login", userConfig.Server)
	resp, err := PostRequest(url, reqData)
	if err != nil {
		fmt.Printf("Error logging in: %v\n", err)
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		fallthrough
	case http.StatusMovedPermanently:
		// 200->301 - Successful login
		cookies := resp.Cookies()
		if len(cookies) > 0 {
			session = cookies[0].Value
			_ = config.SetSession(configPaths, session)
		}
		return nil
	case http.StatusNotFound:
		// 404 - Account not found
		return ErrorAccountDoesNotExist
	default:
		// ??? - Unexpected response
		fmt.Printf("Server error: %d\n", resp.StatusCode)
		return ServerError
	}
}

func LogoutUser() {
	url := fmt.Sprintf("%s/logout", userConfig.Server)
	resp, err := GetRequest(url)
	if err != nil {
		fmt.Printf("Error logging out: %v\n", err)
		return
	} else if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error %d while logging out", resp.StatusCode)
		return
	}

	fmt.Println("Logged out!")
}
