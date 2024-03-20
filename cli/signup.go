package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qeesung/image2ascii/convert"
	"image"
	"io"
	"net/http"
	"strings"
	"yeetfile/cli/crypto"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

const acctNumOnlyWarning string = `
Warning: By creating an account with only an account number, you will be unable
to log back in if you lose the account number provided to you.`

// CreateAccount walks the user through creating a new account with either
// and email+password or just an account number.
func CreateAccount() {
	fmt.Println("1) Email address")
	fmt.Println("2) Account # only")
	acct := utils.StringPrompt("How do you want to set up your account? (1 or 2):")

	if acct != "1" && acct != "2" {
		CreateAccount()
	}

	if acct == "1" {
		err := createEmailAccount("")
		if err != nil {
			fmt.Printf("Error creating email account: %v\n", err)
			return
		}
	} else if acct == "2" {
		err := createNumericAccount()
		if err != nil {
			fmt.Printf("Error creating account #: %v\n", err)
			return
		}
	}

	fmt.Println("Successfully created account! You can now log in with " +
		"`yeetfile login`.")
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

	userKey := crypto.GenerateUserKey([]byte(email), pw)
	loginKeyHash := crypto.GenerateLoginKeyHash(userKey, pw)

	signupData := shared.Signup{
		Identifier:   email,
		LoginKeyHash: loginKeyHash,
	}

	_, err := sendSignup(signupData)
	if err != nil {
		return err
	}

	// Verify user email
	fmt.Printf("A verification code has been sent to %s, please enter "+
		"it below to finish signing up.\n", signupData.Identifier)
	resp, err := verifyEmail(signupData.Identifier)
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

	pw := utils.RequestPassword()
	for len(pw) < 5 {
		fmt.Println("Password must be > 5 characters")
		pw = utils.RequestPassword()
	}

	for !utils.ConfirmPassword(pw) {
		fmt.Println("Error: passwords don't match")
		pw = utils.RequestPassword()
	}

	// We can use an empty value for Identifier, since we're just wanting
	// to use an account ID to log in
	resp, _ := sendSignup(shared.Signup{})

	decoder := json.NewDecoder(resp.Body)
	var signupResponse shared.SignupResponse
	err := decoder.Decode(&signupResponse)
	if err != nil {
		return err
	}

	err = verifyAccountID(signupResponse, pw)
	if err != nil {
		return nil
	}

	fmt.Printf("Your account ID is: %s\n", signupResponse.Identifier)

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

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

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

// verifyAccountID verifies the user's new account using the captcha image
// data sent back from the server on signup.
func verifyAccountID(signupResponse shared.SignupResponse, password []byte) error {
	captcha, err := base64.StdEncoding.DecodeString(signupResponse.Captcha)
	if err != nil {
		fmt.Println("Error displaying captcha")
		return err
	}

	verificationCode := runCLICaptcha(captcha)
	url := fmt.Sprintf("%s/verify-account", userConfig.Server)

	userKey := crypto.GenerateUserKey([]byte(signupResponse.Identifier), password)
	loginKeyHash := crypto.GenerateLoginKeyHash(userKey, password)
	storageKey, _ := crypto.GenerateStorageKey()
	protectedKey := crypto.EncryptChunk(userKey, storageKey)

	reqData, err := json.Marshal(shared.VerifyAccount{
		ID:           signupResponse.Identifier,
		Code:         verificationCode,
		ProtectedKey: protectedKey,
		LoginKeyHash: []byte(loginKeyHash),
	})

	resp, err := PostRequest(url, reqData)
	if err != nil {
		return err
	} else if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
		fmt.Println("Incorrect verification code, please try again.")
		return verifyAccountID(signupResponse, password)
	}

	return nil
}

// runCLICaptcha displays the multi-digit verification code image sent by
// the server as ASCII art in the terminal, and returns the value that the
// user enters.
func runCLICaptcha(imageBytes []byte) string {
	img, _, _ := image.Decode(bytes.NewReader(imageBytes))

	converter := convert.NewImageConverter()
	options := convert.DefaultOptions
	options.Colored = false
	fmt.Print(converter.Image2ASCIIString(img, &options))

	codePrompt := fmt.Sprintf("Enter the %d-digit code above:",
		shared.VerificationCodeLength)
	var code string

	for len(code) != shared.VerificationCodeLength {
		code = utils.StringPrompt(codePrompt)
	}

	return code
}
