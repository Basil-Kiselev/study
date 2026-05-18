package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("20260502233151_initdb.go", upInitdb, downInitdb)
}

func upInitdb(ctx context.Context, tx *sql.Tx) error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		login TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err := tx.Exec(query)
	if err != nil {
		return err
	}

	query = `CREATE TABLE IF NOT EXISTS refresh_tokens (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		revoked_at TIMESTAMP NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func downInitdb(ctx context.Context, tx *sql.Tx) error {
	query := `DROP TABLE IF EXISTS refresh_tokens;`
	_, err := tx.Exec(query)
	if err != nil {
		return err
	}

	query = `DROP TABLE IF EXISTS users;`
	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
