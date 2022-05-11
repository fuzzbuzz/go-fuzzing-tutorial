package handlers

import "database/sql"

func MigrateDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE [users] (
email TEXT NOT NULL PRIMARY KEY,
name TEXT NOT NULL,
phone_number TEXT NOT NULL
);
   `)
	return err
}
