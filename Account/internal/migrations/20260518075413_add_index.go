package migrations

import (
	"context"
	"database/sql"
)

func upAddIndex(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	query := `CREATE INDEX ix_user_login
	ON users (login);`
	_, err := tx.Exec(query)

	return err
}

func downAddIndex(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	query := `DROP INDEX ix_user_login ON users;`
	_, err := tx.Exec(query)

	return err
}
