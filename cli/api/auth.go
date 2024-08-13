package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"yeetfile/cli/requests"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

// Login logs a user into a YeetFile server, returning the server response,
// the session cookie, and any errors.
func (ctx *Context) Login(login shared.Login) (shared.LoginResponse, string, error) {
	url := endpoints.Login.Format(ctx.Server)
	reqData, err := json.Marshal(login)
	if err != nil {
		return shared.LoginResponse{}, "", err
	}

	resp, err := requests.PostRequest(ctx.Session, url, reqData)
	if err != nil {
		return shared.LoginResponse{}, "", err
	}

	var loginResponse shared.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		return shared.LoginResponse{}, "", err
	} else if resp.StatusCode != http.StatusOK {
		return shared.LoginResponse{}, "", utils.ParseHTTPError(resp)
	}

	var session string
	cookies := resp.Cookies()
	if len(cookies) > 0 {
		ctx.Session = cookies[0].Value
		session = cookies[0].Value
	}

	return loginResponse, session, nil
}

// VerifyAccount finalizes account for account-id-only accounts by verifying
// the N-digit verification code and submitting their keys
func (ctx *Context) VerifyAccount(account shared.VerifyAccount) error {
	reqData, err := json.Marshal(account)
	if err != nil {
		return err
	}

	url := endpoints.VerifyAccount.Format(ctx.Server)
	response, err := requests.PostRequest(ctx.Session, url, reqData)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusUnauthorized {
			return errors.New("incorrect verification code")
		}

		return utils.ParseHTTPError(response)
	}

	return nil
}

// SubmitSignup initiates the signup process for an account-ID-only signup,
// returning their new account ID and allowing the user to proceed with verifying
// their new account.
func (ctx *Context) SubmitSignup(signup shared.Signup) (shared.SignupResponse, error) {
	reqData, err := json.Marshal(signup)
	if err != nil {
		return shared.SignupResponse{}, err
	}

	url := endpoints.Signup.Format(ctx.Server)
	response, err := requests.PostRequest(ctx.Session, url, reqData)
	if err != nil {
		return shared.SignupResponse{}, err
	}

	decoder := json.NewDecoder(response.Body)
	var signupResponse shared.SignupResponse
	err = decoder.Decode(&signupResponse)
	if err != nil {
		return shared.SignupResponse{}, err
	} else if len(signupResponse.Error) > 0 {
		return shared.SignupResponse{}, errors.New(signupResponse.Error)
	}

	return signupResponse, nil
}

// VerifyEmail verifies a new user's email using their email and the code sent
// to their email address
func (ctx *Context) VerifyEmail(email, code string) error {
	url := endpoints.VerifyEmail.Format(ctx.Server)
	url += fmt.Sprintf("?email=%s&code=%s", email, code)

	response, err := requests.GetRequest(ctx.Session, url)
	if err != nil {
		return err
	} else if response.StatusCode >= http.StatusBadRequest {
		if response.StatusCode == http.StatusUnauthorized {
			return errors.New("incorrect verification code")
		}
		return errors.New("server error")
	}

	return nil
}

// GetSession returns the current session info.
func (ctx *Context) GetSession() (shared.SessionInfo, error) {
	url := endpoints.Session.Format(ctx.Server)
	resp, err := requests.GetRequest(ctx.Session, url)

	if err != nil {
		return shared.SessionInfo{}, err
	} else if resp.StatusCode != http.StatusOK {
		return shared.SessionInfo{}, utils.ParseHTTPError(resp)
	}

	//var sessionInfo shared.SessionInfo
	//err = json.NewDecoder(resp.Body).Decode(&sessionInfo)
	//if err != nil {
	//	return shared.SessionInfo{}, err
	//}

	return shared.SessionInfo{}, nil
}

// LogOut invalidates the current session for the logged-in user
func (ctx *Context) LogOut() error {
	url := endpoints.Logout.Format(ctx.Server)
	response, err := requests.GetRequest(ctx.Session, url)
	if err != nil {
		return err
	} else if response.StatusCode >= http.StatusBadRequest {
		return utils.ParseHTTPError(response)
	}

	return nil
}

// DeleteAccount removes the current user's YeetFile account.
// NOTE: Currently available only in debug-mode.
func (ctx *Context) DeleteAccount() error {
	url := endpoints.Account.Format(ctx.Server)
	response, err := requests.DeleteRequest(ctx.Session, url)
	if err != nil {
		return err
	} else if response.StatusCode != http.StatusOK {
		return utils.ParseHTTPError(response)
	}

	return nil
}
