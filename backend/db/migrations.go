package db

import "time"

func getMigrationVersion() (int, error) {
	var version int
	s := `SELECT version FROM migrations ORDER BY date DESC LIMIT 1`
	err := db.QueryRow(s).Scan(&version)

	return version, err
}

func setMigrationVersion(version int) error {
	s := `INSERT INTO migrations (version, date) VALUES ($1, $2)`
	_, err := db.Exec(s, version, time.Now().UTC())
	return err
}
