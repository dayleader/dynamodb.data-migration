package domain

import "errors"

// MigrationContext - describes migration context.
type MigrationContext struct {
	MigrationsDir   string
	MigrationsTable string
}

// NewMigrationContext - constructs a new migration context.
func NewMigrationContext() *MigrationContext {
	return &MigrationContext{}
}

// Validate - checks if the migration context properties are valid.
func (m MigrationContext) Validate() error {
	if len(m.MigrationsDir) == 0 {
		return errors.New("Migrations directory required")
	}
	if len(m.MigrationsTable) == 0 {
		return errors.New("Migrations table name required")
	}
	return nil
}
