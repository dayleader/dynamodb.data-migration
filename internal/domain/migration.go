package domain

import (
	"strconv"
	"time"
)

// Migration consts.
const (
	MigrationFilePattern = `(.*\/)?((\d+).(\d+).(\d+).*\.json)$`
)

// Version - version struct.
type Version struct {
	Major int
	Minor int
	Patch int
}

// String - returns a string representation.
func (ver Version) String() string {
	return strconv.Itoa(ver.Major) + "." + strconv.Itoa(ver.Minor) + "." + strconv.Itoa(ver.Patch)
}

// ID - returns a string representation of ID.
func (ver Version) ID() string {
	return strconv.Itoa(ver.Major) + "." + strconv.Itoa(ver.Minor) + "." + strconv.Itoa(ver.Patch)
}

// Metadata - migrations metadata.
type Metadata struct {
	StartTime     int64
	ExecutionTime int64
}

// MigrationRecord - migration record.
type MigrationRecord struct {
	Version  Version
	Name     string
	Metadata Metadata
}

// Migration - migration struct.
type Migration struct {
	MigrationRecord
	Content []byte
}

// String - returns a string representation.
func (mig Migration) String() string {
	return mig.Version.String() + ": " + mig.Name
}

// SetExecutionTime - sets the migration execution time.
func (mig *Migration) SetExecutionTime(start, end time.Time) {
	mig.Metadata.StartTime = start.Unix()
	mig.Metadata.ExecutionTime = end.Unix() - start.Unix()
}

// QueryExecutor - query executor.
type QueryExecutor interface {

	// ExecuteQueries - execute migration queries.
	ExecuteQueries(queries []*DynamoDBQuery) error
}

// MigrationRepository - migration repository interface.
type MigrationRepository interface {
	QueryExecutor

	// IsMigrationRecordExist - checks if the migration record exist.
	IsMigrationRecordExist(ver Version) (bool, error)

	// CreateMigrationRecord - creates migration record.
	CreateMigrationRecord(migrationRecord MigrationRecord) error
}

// MigrationStorage - migration storage.
type MigrationStorage interface {

	// GetExecutableMigrations - returns executable migrations.
	GetExecutableMigrations() ([]*Migration, error)
}

// MigrationService - migration service.
type MigrationService interface {

	// RunMigrations - runs migrations.
	Migrate() (applied int, err error)
}
