package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	fmt.Println("Successfully created account! You are now logged in.")
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

	_, err := sendSignup(signupData)
	if err != nil {
		return err
	}

	// Verify user email
	fmt.Printf("A verification code has been sent to %s, please enter "+
		"it below to finish signing up.\n", signupData.Email)
	resp, err := verifyEmail(signupData.Email)
	if err != nil {
		return nil
	}

	// Use verification response to set initial user session
	err = SetSessionFromCookies(resp)
	if err != nil {
		return err
	}

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

	// We can use an empty signup struct here, since we aren't using an
	// email or password
	resp, _ := sendSignup(shared.Signup{})
	fmt.Println(resp)

	return nil
}

// sendSignup handles signing a user up based on the information provided in the
// signup struct. Returns the response body if successful.
func sendSignup(signupData shared.Signup) (*http.Response, error) {
	reqData, err := json.Marshal(signupData)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/signup", userConfig.Server)
	resp, err := PostRequest(url, reqData)
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

	return resp, nil
}

// verifyEmail prompts the user for the code sent to their email and uses it
// to finish verifying their account.
func verifyEmail(email string) (*http.Response, error) {
	code := utils.StringPrompt("Enter Verification Code:")
	url := fmt.Sprintf(
		"%s/verify?email=%s&code=%s",
		userConfig.Server,
		email,
		code)

	resp, err := GetRequest(url)
	if err != nil {
		return nil, err
	} else if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
		fmt.Println("Incorrect verification code, please try again.")
		return verifyEmail(email)
	}

	return resp, nil
}
