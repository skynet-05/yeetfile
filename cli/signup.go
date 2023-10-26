package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"yeetfile/cli/config"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

const acctNumOnlyWarning string = `
Warning: By creating an account with only an account number, you will be unable
to log back in if you lose the account number provided to you.`

// CreateAccount walks the user through creating a new account with either
// and email+password or just an account number.
func CreateAccount() {
	fmt.Println("1) Email + password")
	fmt.Println("2) Account # only")
	acct := utils.StringPrompt("How do you want to set up your account? (1 or 2):")

	if acct != "1" && acct != "2" {
		CreateAccount()
	}

	if acct == "1" && createEmailAccount("") != nil {
		fmt.Println("Error creating email account")
	} else if acct == "2" && createNumericAccount() != nil {
		fmt.Println("Error creating account #")
	}
}

// createEmailAccount creates a new account using an email and password. The
// user's email must be verified before account creation is finalized.
func createEmailAccount(email string) error {
	if len(email) == 0 {
		email = utils.StringPrompt("Email:")
	}

	pw := utils.RequestPassword()
	if len(pw) < 5 {
		fmt.Println("Password must be > 5 characters")
		return createEmailAccount(email)
	} else if !utils.ConfirmPassword(pw) {
		fmt.Println("Error: passwords don't match")
		return createEmailAccount(email)
	}

	signupData := shared.Signup{
		Email:    email,
		Password: string(pw),
	}

	resp, _ := sendSignup(signupData)
	fmt.Println(string(resp))

	// TODO: Validate email
	return nil
}

// createNumericAccount creates a new account with only a numeric account ID
// for logging in.
func createNumericAccount() error {
	fmt.Println(acctNumOnlyWarning)
	confirm := utils.StringPrompt("Confirm (enter \"y\" or \"n\"):")

	if strings.ToLower(confirm) != "y" {
		fmt.Println("Account creation aborted")
		return errors.New("user didn't confirm acct warning")
	}

	// We can use an empty signup struct here, since we
	resp, _ := sendSignup(shared.Signup{})
	fmt.Println(resp)

	return nil
}

// sendSignup handles signing a user up based on the information provided in the
// signup struct. Returns the response body if successful.
func sendSignup(signupData shared.Signup) ([]byte, error) {
	client := &http.Client{}
	reqData, err := json.Marshal(signupData)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/signup", userConfig.Server)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Error %d: %s\n", resp.StatusCode, string(body))
		return nil, errors.New("error creating account")
	}

	cookies := resp.Cookies()
	if len(cookies) > 0 {
		session = cookies[0].Value
		err = config.SetSession(configPaths, session)
		if err != nil {
			fmt.Printf("Failed to save user session: %v\n", err)
			return nil, err
		}
	}

	return body, nil
}
