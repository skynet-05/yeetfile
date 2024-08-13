package db

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"yeetfile/backend/config"
	"yeetfile/backend/server/subscriptions"
	"yeetfile/shared"
)

type User struct {
	ID                 string
	Email              string
	PasswordHash       []byte
	ProtectedKey       []byte
	PublicKey          []byte
	PaymentID          string
	MemberExp          time.Time
	StorageAvailable   int
	StorageUsed        int
	SendAvailable      int
	SendUsed           int
	SubscriptionMethod string
}

type UserStorage struct {
	StorageAvailable int
	StorageUsed      int
}

type UserSend struct {
	SendAvailable int
	SendUsed      int
}

var defaultExp time.Time

var UserAlreadyExists = errors.New("user already exists")
var UserLimitReached = errors.New("user limit has been reached")

// NewUser creates a new user in the "users" table, ensuring that the email
// provided is not already in use.
func NewUser(user User) (string, error) {
	if config.YeetFileConfig.MaxUserCount > 0 {
		count, err := GetUserCount()
		if err != nil {
			return "", err
		} else if count == config.YeetFileConfig.MaxUserCount {
			return "", UserLimitReached
		}

		config.YeetFileConfig.CurrentUserCount = count
	}

	if len(user.Email) > 0 {
		rows, err := db.Query(`SELECT * from users WHERE email = $1`, user.Email)
		if err != nil {
			return "", err
		} else if rows.Next() {
			return "", UserAlreadyExists
		}

		rows.Close()
	}

	if len(user.ID) == 0 {
		user.ID = CreateUniqueUserID()
	} else {
		if UserIDExists(user.ID) {
			return "", UserAlreadyExists
		}
	}

	paymentID := CreateUniquePaymentID()

	s := `INSERT INTO users (
                   id,
                   email,
                   pw_hash,
                   payment_id,
                   send_available,
                   storage_available,
                   member_expiration,
                   last_upgraded_month,
                   protected_key,
                   public_key)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := db.Exec(
		s,
		user.ID,
		user.Email,
		user.PasswordHash,
		paymentID,
		config.YeetFileConfig.DefaultUserSend,
		config.YeetFileConfig.DefaultUserStorage,
		defaultExp,
		-1,
		user.ProtectedKey,
		user.PublicKey)
	if err != nil {
		return "", err
	}

	if config.YeetFileConfig.MaxUserCount > 0 {
		config.YeetFileConfig.CurrentUserCount += 1
	}

	return user.ID, nil
}

// GetUserCount returns the total number of users in the table
func GetUserCount() (int, error) {
	rows, err := db.Query(`SELECT COUNT(*) from users`)
	if err != nil || !rows.Next() {
		return 0, err
	}

	defer rows.Close()

	var count int
	err = rows.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CreateUniqueUserID creates a 16 digit user ID that is not already being used
// in the user database.
func CreateUniqueUserID() string {
	id := shared.GenRandomNumbers(16)
	for UserIDExists(id) {
		id = shared.GenRandomNumbers(16)
	}

	return id
}

// CreateUniquePaymentID creates a 16 character payment ID that is not already
// being used in the user database.
func CreateUniquePaymentID() string {
	paymentID := shared.GenRandomString(16)
	for PaymentIDExists(paymentID) {
		paymentID = shared.GenRandomString(16)
	}

	return paymentID
}

// GetUserStorage returns UserStorage and UserSend struct containing the user's
// available and used limits for storing and sending files
func GetUserStorage(id string) (UserStorage, UserSend, error) {
	rows, err := db.Query(`
	    SELECT storage_available, storage_used, send_available, send_used 
	    FROM users 
	    WHERE id = $1`, id)
	if err != nil {
		return UserStorage{}, UserSend{}, err
	} else if !rows.Next() {
		errorStr := fmt.Sprintf("unable to find user with id '%s'", id)
		return UserStorage{}, UserSend{}, errors.New(errorStr)
	}

	defer rows.Close()

	var storageAvailable int
	var storageUsed int
	var sendAvailable int
	var sendUsed int
	err = rows.Scan(&storageAvailable, &storageUsed, &sendAvailable, &sendUsed)
	if err != nil {
		return UserStorage{}, UserSend{}, err
	}

	return UserStorage{
			StorageAvailable: storageAvailable, StorageUsed: storageUsed,
		},
		UserSend{
			SendAvailable: sendAvailable, SendUsed: sendUsed,
		},
		nil
}

// UpdateStorageUsed updates the amount of storage used by the user. Can be a
// negative number to remove storage space.
func UpdateStorageUsed(userID string, amount int) error {
	s := `UPDATE users 
	      SET storage_used = CASE 
	                           WHEN storage_used + $1 < 0 THEN 0
	                           ELSE storage_used + $1
	                         END
	      WHERE id=$2`
	_, err := db.Exec(s, amount, userID)
	return err
}

// RotateUserPaymentID overwrites the previous payment ID once a transaction is
// completed.
func RotateUserPaymentID(paymentID string) error {
	rows, err := db.Query(`SELECT id from users WHERE payment_id = $1`, paymentID)
	if err != nil {
		return err
	} else if !rows.Next() {
		errorStr := fmt.Sprintf("unable to find user with payment id '%s'", paymentID)
		return errors.New(errorStr)
	}

	defer rows.Close()

	newID := shared.GenRandomString(16)
	for PaymentIDExists(newID) {
		newID = shared.GenRandomString(16)
	}

	// Read in account ID for the user
	var accountID string
	err = rows.Scan(&accountID)

	// Replace payment ID
	s := `UPDATE users
	      SET payment_id=$1
	      WHERE id=$2`

	_, err = db.Exec(s, newID, accountID)
	if err != nil {
		return err
	}

	return nil
}

func SetNewPassword(email string, pwHash []byte) error {
	rows, err := db.Query(`SELECT id from users WHERE email = $1`, email)
	if err != nil {
		return err
	} else if !rows.Next() {
		errorStr := fmt.Sprintf("unable to find user with email '%s'", email)
		return errors.New(errorStr)
	}

	defer rows.Close()

	// Read in account ID for the user
	var accountID string
	err = rows.Scan(&accountID)

	// Replace payment ID
	s := `UPDATE users
	      SET pw_hash=$1
	      WHERE id=$2`

	_, err = db.Exec(s, pwHash, accountID)
	if err != nil {
		return err
	}

	return nil
}

// UserIDExists checks the users table to see if the provided id is already
// being used for another user.
func UserIDExists(id string) bool {
	rows, err := db.Query(`SELECT id FROM users WHERE id = $1`, id)
	if err != nil {
		log.Fatalf("Error querying user id: %v", err)
		return true
	}

	defer rows.Close()

	// If any rows are returned, the id exists
	if rows.Next() {
		return true
	}

	return false
}

// PaymentIDExists checks the user table to see if the provided payment ID
// (for Stripe + BTCPay) already exists for another user.
func PaymentIDExists(paymentID string) bool {
	rows, err := db.Query(`SELECT * FROM users WHERE payment_id = $1`, paymentID)
	if err != nil {
		log.Fatalf("Error querying user payment id: %v", err)
		return true
	}

	defer rows.Close()

	// If any rows are returned, the id exists
	if rows.Next() {
		return true
	}

	return false
}

// GetUserPasswordHashByEmail retrieves the password hash for a user with the
// provided email address.
func GetUserPasswordHashByEmail(email string) ([]byte, error) {
	rows, err := db.Query(`
		SELECT pw_hash
		FROM users 
		WHERE email = $1`, email)
	if err != nil {
		log.Printf("Error querying for user by email: %v", err)
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		var pwHash []byte
		err = rows.Scan(&pwHash)
		if err != nil {
			return nil, err
		}

		return pwHash, nil
	}

	return nil, errors.New("unable to find user")
}

// GetUserPasswordHashByID retrieves the password hash for a user with the
// provided ID.
func GetUserPasswordHashByID(id string) ([]byte, error) {
	rows, err := db.Query(`
		SELECT pw_hash
		FROM users 
		WHERE id = $1`, id)
	if err != nil {
		log.Printf("Error querying for user by id: %v", err)
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		var pwHash []byte
		err = rows.Scan(&pwHash)
		if err != nil {
			return nil, err
		}

		return pwHash, nil
	}

	return nil, errors.New("unable to find user")
}

// GetUserKeys retrieves the user's public key and their private key, the latter
// is encrypted with their user key (which is generated client side and never stored)
func GetUserKeys(id string) ([]byte, []byte, error) {
	rows, err := db.Query(`
		SELECT protected_key, public_key
		FROM users 
		WHERE id = $1`, id)
	if err != nil {
		log.Printf("Error querying for user by id: %v", err)
		return nil, nil, err
	}

	defer rows.Close()
	if rows.Next() {
		var publicKey []byte
		var protectedKey []byte
		err = rows.Scan(&protectedKey, &publicKey)
		if err != nil {
			return nil, nil, err
		}

		return protectedKey, publicKey, nil
	}

	return nil, nil, errors.New("unable to find user")
}

// GetUserByID retrieves a User struct for given user ID.
func GetUserByID(id string) (User, error) {
	rows, err := db.Query(`
		SELECT email, payment_id, member_expiration,
		       send_available, send_used, 
		       storage_available, storage_used,
		       sub_method
		FROM users
		WHERE id = $1`, id)
	if err != nil {
		log.Printf("Error querying for user by id: %s\n", id)
		return User{}, err
	}

	defer rows.Close()
	if rows.Next() {
		var email string
		var paymentID string
		var expiration time.Time
		var sendAvailable int
		var sendUsed int
		var storageAvailable int
		var storageUsed int
		var subMethod string
		err = rows.Scan(
			&email, &paymentID, &expiration,
			&sendAvailable, &sendUsed,
			&storageAvailable, &storageUsed,
			&subMethod)
		if err != nil {
			return User{}, err
		}

		return User{
			Email:              email,
			PaymentID:          paymentID,
			MemberExp:          expiration,
			SendAvailable:      sendAvailable,
			SendUsed:           sendUsed,
			StorageAvailable:   storageAvailable,
			StorageUsed:        storageUsed,
			SubscriptionMethod: subMethod,
		}, nil
	}

	return User{}, errors.New("error fetching user by id")
}

func GetUserPubKey(userID string) ([]byte, error) {
	rows, err := db.Query(`SELECT public_key FROM users WHERE id=$1`, userID)
	if err != nil {
		log.Printf("Error querying for public key by user id: %v\n", err)
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		var publicKey []byte
		err = rows.Scan(&publicKey)
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	}

	return nil, errors.New("user public key not found")
}

func GetUserPublicName(userID string) (string, error) {
	rows, err := db.Query(`SELECT email FROM users WHERE id=$1`, userID)
	if err != nil {
		log.Printf("Error querying for user's public name")
		return "", err
	}

	defer rows.Close()
	if rows.Next() {
		var email string
		err = rows.Scan(&email)
		if err != nil {
			return "", err
		}

		if len(email) == 0 {
			idTail := userID[len(userID)-4:]
			return fmt.Sprintf("*%s", idTail), nil
		}

		return email, nil
	}

	return "", errors.New("unable to fetch user's public name")
}

// GetUserIDByEmail returns a user's ID given their email address.
func GetUserIDByEmail(email string) (string, error) {
	rows, err := db.Query(`
		SELECT id
		FROM users 
		WHERE email = $1`, email)
	if err != nil {
		log.Printf("Error querying for user by email: %v", err)
		return "", err
	}

	defer rows.Close()
	if rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return "", err
		}

		return id, nil
	}

	return "", errors.New("unable to find user")
}

// GetUserSendLimits returns the amount of used and available bytes for
// sending files
func GetUserSendLimits(id string) (int, int, error) {
	rows, err := db.Query(`
		SELECT send_used, send_available
		FROM users
		WHERE id = $1`, id)
	if err != nil {
		log.Printf("Error querying for user by id: %s\n", id)
		return 0, 0, err
	}

	defer rows.Close()
	if rows.Next() {
		var sendUsed int
		var sendAvailable int
		err = rows.Scan(&sendUsed, &sendAvailable)
		if err != nil {
			log.Printf("Error reading limits for user %s\n", id)
			return 0, 0, err
		}

		return sendUsed, sendAvailable, nil
	}

	return 0, 0, errors.New("unable to find user by id")
}

func GetPaymentIDByUserID(userID string) (string, error) {
	rows, err := db.Query(`
		SELECT payment_id
		FROM users
		WHERE id = $1`, userID)
	if err != nil {
		log.Println("Error querying for payment_id")
		return "", err
	}

	defer rows.Close()
	if rows.Next() {
		var paymentID string
		err = rows.Scan(&paymentID)
		if err != nil {
			log.Println("Error fetching payment ID")
			return "", err
		}

		return paymentID, nil
	}

	return "", errors.New("unable to find payment id by user id")
}

func GetUserEmailByPaymentID(paymentID string) (string, error) {
	rows, err := db.Query(`
		SELECT email
		FROM users
		WHERE payment_id = $1`, paymentID)
	if err != nil {
		log.Printf("Error querying for user by payment_id: %s\n", paymentID)
		return "", err
	}

	defer rows.Close()
	if rows.Next() {
		var email string
		err = rows.Scan(&email)
		if err != nil {
			log.Printf("Error fetching email for user with payment id %s\n", paymentID)
			return "", err
		}

		return email, nil
	}

	return "", errors.New("unable to find user by payment id")
}

// SetUserSubscription updates a user's subscription to have the correct amount
// of storage and sending available
func SetUserSubscription(
	paymentID, subTag, subMethod string,
	exp time.Time,
	storage, send int,
) error {
	subDuration, err := subscriptions.GetSubscriptionDuration(subTag)
	if err != nil {
		return err
	}

	subType, err := subscriptions.GetSubscriptionType(subTag)
	if err != nil {
		return err
	}

	s := `UPDATE users
              SET member_expiration=$1,
                  storage_available=$2, send_available=$3,
                  sub_duration=$4, sub_type=$5, sub_method=$6,
                  last_upgraded_month=$7
              WHERE payment_id=$7`

	_, err = db.Exec(s,
		exp,
		storage, send,
		subDuration, subType, subMethod,
		time.Now().Month(), paymentID)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserSendUsed adds an amount of bytes (size) to a user's send_used
// given their user ID.
func UpdateUserSendUsed(id string, size int) error {
	s := `UPDATE users
          SET send_used=send_used + $2
          WHERE id=$1`

	_, err := db.Exec(s, id, size)
	if err != nil {
		return err
	}

	return nil
}

// CheckMemberships inspects each user's membership and updates their available
// transfer if their membership is still valid
func CheckMemberships() {
	s := `SELECT id, member_expiration FROM users
              WHERE last_upgraded_month != $1`
	rows, err := db.Query(s, int(time.Now().Month()))
	if err != nil {
		log.Printf("Error retrieving user memberships: %v", err)
		return
	}

	var upgradeIDs []string
	var revertIDs []string
	now := time.Now()

	defer rows.Close()
	for rows.Next() {
		var id string
		var exp time.Time

		err = rows.Scan(&id, &exp)

		if err != nil {
			log.Printf("Error scanning user rows: %v", err)
			return
		}

		if exp.Before(time.Now()) {
			// User doesn't have an active membership, set send to
			// default amount
			revertIDs = append(revertIDs, id)
			continue
		} else if now.Day() == exp.Day() || ExpDateRollover(now, exp) {
			// User has an active membership
			upgradeIDs = append(upgradeIDs, id)
		}
	}

	if len(revertIDs) > 0 {
		// Add the default send and storage amount to all users whose
		// memberships are no longer active
		u := `UPDATE users
		      SET send_used=0,
		          send_available=$1,
		          storage_available=$2,
		          last_upgraded_month=$3
		      WHERE id=ANY($4)`

		ids := fmt.Sprintf("{%s}", strings.Join(revertIDs, ","))
		_, err = db.Exec(u,
			config.YeetFileConfig.DefaultUserSend,
			config.YeetFileConfig.DefaultUserStorage,
			int(now.Month()),
			ids)
		if err != nil {
			panic(err)
		}
	}

	// TODO: Implement membership check and update storage/send appropriately

	time.Sleep(3600 * time.Second)
	CheckMemberships()
}

// ExpDateRollover checks to see if the user's membership expiration date takes
// place on a day that doesn't exist in other months. If so, the user's transfer
// limit should be upgraded "early". For example:
//
// - Expiration: Dec 31
// - Today: June 30
//
// In this scenario, the membership should be upgraded today. The 31st will
// never occur in June, but the following day would be a new month.
func ExpDateRollover(now time.Time, exp time.Time) bool {
	if exp.Day() <= 28 {
		// Skip check, the expiration date is within the bounds of all
		// monthly days
		return false
	}

	return exp.Day() > now.Day() && now.AddDate(0, 0, 1).Month() > now.Month()
}

func DeleteUser(id string) error {
	s := `DELETE FROM users WHERE id=$1`
	_, err := db.Exec(s, id)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	var err error
	defaultExp, err = time.Parse(time.RFC1123, time.RFC1123)
	if err != nil {
		panic(err)
	}
}
