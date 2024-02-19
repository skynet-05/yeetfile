package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"yeetfile/cli/config"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

var ErrorAccountDoesNotExist = errors.New("account does not exist")
var ErrorIncorrectPassword = errors.New("incorrect password")

func LoginUser(loop bool) bool {
	fmt.Println("1) Email + password")
	fmt.Println("2) Account # only")
	loginOpt := utils.StringPrompt("How do you want to login? (1 or 2):")

	if loginOpt != "1" && loginOpt != "2" {
		return LoginUser(loop)
	}

	var err error
	if loginOpt == "1" {
		err = loginWithEmail("")
	} else {
		err = loginWithAccountID()
	}

	if loop && err != nil {
		return LoginUser(loop)
	}

	return err == nil
}

func loginWithEmail(email string) error {
	if len(email) == 0 {
		email = utils.StringPrompt("Email:")
	}

	pw := utils.RequestPassword()
	err := sendLogin(shared.Login{
		Identifier: email,
		Password:   string(pw),
	})

	if err != nil {
		if errors.Is(err, ErrorAccountDoesNotExist) {
			fmt.Println("Error: Account does not exist or incorrect password")
			return loginWithEmail("")
		} else if errors.Is(err, ErrorIncorrectPassword) {
			fmt.Println("Error: Account does not exist or incorrect password")
			return loginWithEmail(email)
		} else {
			return errors.New("failed to log in")
		}
	}

	fmt.Println("Successfully logged in!")
	return nil
}

func loginWithAccountID() error {
	account := utils.StringPrompt("Account #:")
	err := sendLogin(shared.Login{
		Identifier: account,
		Password:   "",
	})

	if err != nil {
		if errors.Is(err, ErrorAccountDoesNotExist) {
			fmt.Println("Error: Account does not exist")
			return loginWithAccountID()
		} else {
			return errors.New("failed to log in")
		}
	}

	fmt.Println("Successfully logged in!")
	return nil
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
		fmt.Printf("An account was not found for %s\n", login.Identifier)
		return ErrorAccountDoesNotExist
	case http.StatusUnauthorized:
		// 401 - Account credentials incorrect
		fmt.Println("Account password is incorrect")
		return ErrorIncorrectPassword
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
