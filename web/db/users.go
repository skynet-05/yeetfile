package db

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"yeetfile/shared"
	"yeetfile/web/utils"
)

type User struct {
	ID           string
	Email        string
	PasswordHash []byte
	ProtectedKey []byte
	Meter        int
	PaymentID    string
	MemberExp    time.Time
}

var defaultExp time.Time
var defaultMeter = utils.GetEnvVarInt("YEETFILE_METER", 1024*1024*1024*5) // 5gb

var UserAlreadyExists = errors.New("user already exists")

// NewUser creates a new user in the "users" table, ensuring that the email
// provided is not already in use.
func NewUser(user User) (string, error) {
	if len(user.Email) > 0 {
		rows, err := db.Query(`SELECT * from users WHERE email = $1`, user.Email)
		if err != nil {
			return "", err
		} else if rows.Next() {
			return "", UserAlreadyExists
		}
	}

	if len(user.ID) == 0 {
		user.ID = CreateUniqueUserID()
	} else {
		if UserIDExists(user.ID) {
			return "", UserAlreadyExists
		}
	}

	paymentID := CreateUniquePaymentID()

	s := `INSERT INTO users
	      (id, email, pw_hash, payment_id, meter, member_expiration, last_upgraded_month, protected_key)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := db.Exec(s, user.ID, user.Email, user.PasswordHash, paymentID, 0, defaultExp, -1, user.ProtectedKey)
	if err != nil {
		return "", err
	}

	return user.ID, nil
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

// GetUserByID retrieves a User struct for given user ID.
func GetUserByID(id string) (User, error) {
	rows, err := db.Query(`
		SELECT email, meter, payment_id, member_expiration
		FROM users
		WHERE id = $1`, id)
	if err != nil {
		log.Printf("Error querying for user by id: %s\n", id)
		return User{}, err
	}

	if rows.Next() {
		var email string
		var meter int
		var paymentID string
		var expiration time.Time
		err = rows.Scan(&email, &meter, &paymentID, &expiration)
		if err != nil {
			return User{}, err
		}

		return User{
			Email:     email,
			Meter:     meter,
			PaymentID: paymentID,
			MemberExp: expiration,
		}, nil
	}

	return User{}, errors.New("unexpected error fetching user by id")
}

func GetUserByPaymentID(paymentID string) (User, error) {
	rows, err := db.Query(`
		SELECT email, meter, member_expiration
		FROM users
		WHERE payment_id = $1`, paymentID)
	if err != nil {
		log.Printf("Error querying for user by payment_id: %s\n", paymentID)
		return User{}, err
	}

	if rows.Next() {
		var email string
		var meter int
		var expiration time.Time
		err = rows.Scan(&email, &meter, &expiration)
		if err != nil {
			return User{}, err
		}

		return User{
			Email:     email,
			Meter:     meter,
			PaymentID: paymentID,
			MemberExp: expiration,
		}, nil
	}

	return User{}, errors.New("unexpected error fetching user by payment_id")
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

// GetUserMeter returns a user's meter given their user ID.
func GetUserMeter(id string) (int, error) {
	rows, err := db.Query(`
		SELECT meter
		FROM users
		WHERE id = $1`, id)
	if err != nil {
		log.Printf("Error querying for user by id: %s\n", id)
		return 0, err
	}

	if rows.Next() {
		var meter int
		err = rows.Scan(&meter)
		if err != nil {
			log.Printf("Error reading meter for user %s\n", id)
			return 0, err
		}

		return meter, nil
	}

	return 0, errors.New("unable to find user by id")
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

// SetUserMembershipExpiration updates a user's membership expiration to be
// one year from the current payment time.
func SetUserMembershipExpiration(paymentID string, exp time.Time) error {
	s := `UPDATE users
              SET member_expiration=$1,
                  meter=CASE WHEN meter < $2 THEN $2 ELSE meter END,
                  last_upgraded_month=$3
              WHERE payment_id=$3`

	_, err := db.Exec(s, exp, defaultMeter, paymentID, time.Now().Month())
	if err != nil {
		return err
	}

	return nil
}

// AddUserStorage adds amount to the meter column for a user with the matching
// payment ID. Once the payment ID is used here, it should be replaced by calling
// RotateUserPaymentID.
func AddUserStorage(paymentID string, amount int) error {
	s := `UPDATE users
	      SET meter=meter + $1
	      WHERE payment_id=$2`

	_, err := db.Exec(s, amount, paymentID)
	if err != nil {
		return err
	}

	return nil
}

// ReduceUserStorage subtracts an amount of bytes (size) from a user's meter
// given their user ID.
func ReduceUserStorage(id string, size int) error {
	s := `UPDATE users
          SET meter=meter - $2
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
	s := `SELECT id, member_expiration, last_upgraded_month FROM users
              WHERE member_expiration >= $1::date`
	rows, err := db.Query(s, time.Now())
	if err != nil {
		log.Printf("Error retrieving user memberships: %v", err)
		return
	}

	var upgradeIDs []string
	now := time.Now()

	for rows.Next() {
		var id string
		var exp time.Time
		var lastUpgradedMonth int

		err = rows.Scan(&id, &exp, &lastUpgradedMonth)

		if err != nil {
			log.Printf("Error scanning user rows: %v", err)
			return
		}

		if int(exp.Month()) == lastUpgradedMonth {
			// User has already had their transfer limit upgraded
			// this month
			continue
		}

		if now.Day() == exp.Day() || ExpDateRollover(now, exp) {
			// User
			upgradeIDs = append(upgradeIDs, id)
		}
	}

	if len(upgradeIDs) > 0 {
		// Add the default meter amount to all users whose memberships
		// are still active
		u := `UPDATE users
		      SET meter=meter + $1,
		          last_upgraded_month=$2
		      WHERE id=ANY($3)`

		ids := fmt.Sprintf("{%s}", strings.Join(upgradeIDs, ","))
		_, err = db.Exec(u, defaultMeter, int(now.Month()), ids)
		if err != nil {
			panic(err)
		}
	}

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

func init() {
	var err error
	defaultExp, err = time.Parse(time.RFC1123, time.RFC1123)
	if err != nil {
		panic(err)
	}
}
