//go:build server_test

package api

import (
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"yeetfile/cli/crypto"
	"yeetfile/shared"
)

var userPassword = "password"
var server string

type TestUser struct {
	id      string
	privKey []byte
	pubKey  []byte
	context *Context
}

var (
	UserA TestUser
	UserB TestUser
)

var userFileIDs map[string][]string

func setupTestUser() TestUser {
	ctx := InitContext(server, "")
	signup, err := ctx.SubmitSignup(shared.Signup{
		Identifier:              "",
		LoginKeyHash:            nil,
		PublicKey:               nil,
		ProtectedPrivateKey:     nil,
		ProtectedVaultFolderKey: nil,
	})

	if _, ok := err.(*net.OpError); ok {
		log.Fatalf("Unable to connect to %s -- is it running?", server)
	} else if err != nil {
		log.Fatal("Failed to sign up test user")
	}

	signupKeys, err := crypto.GenerateSignupKeys(signup.Identifier, userPassword)

	verifyAcct := shared.VerifyAccount{
		ID:                      signup.Identifier,
		Code:                    "123456", // Requires YEETFILE_DEBUG=1 to succeed
		LoginKeyHash:            signupKeys.LoginKeyHash,
		ProtectedPrivateKey:     signupKeys.ProtectedPrivateKey,
		PublicKey:               signupKeys.PublicKey,
		ProtectedVaultFolderKey: signupKeys.ProtectedRootFolderKey,
	}

	err = ctx.VerifyAccount(verifyAcct)
	if err != nil {
		log.Fatalf("Failed to verify account")
	}

	_, _, err = ctx.Login(shared.Login{
		Identifier:   signup.Identifier,
		LoginKeyHash: signupKeys.LoginKeyHash,
	})

	if err != nil {
		log.Fatalf("Err logging in: %v\n", err)
	}

	privKey, _ := crypto.DecryptChunk(
		signupKeys.UserKey,
		signupKeys.ProtectedPrivateKey)

	return TestUser{
		id:      signup.Identifier,
		context: ctx,
		privKey: privKey,
		pubKey:  signupKeys.PublicKey,
	}
}

func cleanUpUserAccount(user TestUser) {
	err := user.context.DeleteAccount(user.id)
	if err != nil {
		log.Printf("ERROR - failed to delete user account: %v\n", err)
	}
}

func TestMain(m *testing.M) {
	userFileIDs = make(map[string][]string)
	port, exists := os.LookupEnv("YEETFILE_PORT")
	if !exists {
		port = "8090"
	}

	server = fmt.Sprintf("http://localhost:%s", port)

	UserA = setupTestUser()
	UserB = setupTestUser()

	if UserA.id == UserB.id {
		log.Fatal("User IDs are identical")
	} else if len(UserA.id) == 0 || len(UserB.id) == 0 {
		log.Fatal("User IDs are empty")
	}

	// Run tests
	exitCode := m.Run()

	// TODO: Teardown (remove users, files, etc)
	// Remove UserA files
	fmt.Println("Cleaning up...")
	cleanUpUserAccount(UserA)
	cleanUpUserAccount(UserB)

	os.Exit(exitCode)
}
