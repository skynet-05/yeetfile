package db

import (
	"errors"
	"strconv"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

var IncorrectPassIndexChangeIDErr = errors.New("incorrect pass index change id")

// InitPassIndex initializes an entry in the pass_index table with the user's
// ID and an initial change ID. The encrypted data stays empty, since the user
// hasn't added anything that can be indexed yet.
func InitPassIndex(userID string) error {
	s := `INSERT INTO pass_index (user_id, change_id) VALUES ($1, $2)`
	changeID := shared.GenRandomNumbers(constants.ChangeIDLength)
	changeIDNum, _ := strconv.Atoi(changeID)

	_, err := db.Exec(s, userID, changeIDNum)
	return err
}

// UpdatePassIndex updates the user's password index with the new encrypted
// shared.PassIndex data. If the provided change ID doesn't match, the affected
// row count will return 0, indicating that the user needs to fetch an updated
// pass index before continuing.
func UpdatePassIndex(userID string, changeID int, encData []byte) error {
	s := `UPDATE pass_index SET enc_data=$3 WHERE user_id=$1 AND change_id=$2`
	result, err := db.Exec(s, userID, changeID, encData)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return IncorrectPassIndexChangeIDErr
	}

	return nil
}
