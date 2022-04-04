package migration

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"dynamodb.data-migration/internal/domain"
)

const (
	statusOK = iota
	statusError
	statusExist
)

type service struct {
	repository  domain.MigrationRepository
	storage     domain.MigrationStorage
	queryParser domain.QueryParser
}

// NewMigrationService creates a service with necessary dependencies.
func NewMigrationService(
	repository domain.MigrationRepository,
	storage domain.MigrationStorage,
	queryParser domain.QueryParser,
) domain.MigrationService {
	return &service{
		repository:  repository,
		storage:     storage,
		queryParser: queryParser,
	}
}

func (s *service) Migrate() (applied int, err error) {

	// Get executable migrations.
	//
	migrations, err := s.storage.GetExecutableMigrations()
	if err != nil {
		return applied, err
	}

	// Sort migrations in the correct order.
	//
	sort.Slice(migrations, func(i, j int) bool {
		if migrations[i].Version.Major == migrations[j].Version.Major {
			return migrations[i].Version.Minor < migrations[j].Version.Minor
		}
		return migrations[i].Version.Major < migrations[j].Version.Major
	})

	// Run migrations.
	//
	for _, migration := range migrations {
		status, err := s.runMigration(migration)
		if err != nil {
			return applied, fmt.Errorf("Migration failed: %s, error: %v", migration.Name, err)
		}
		switch status {
		case statusOK:
			applied++
			log.Println("Migration applied:", migration.Name)
		case statusExist:
			log.Println("Migration exists:", migration.Name)
		default:
			log.Println("Unknown status code:", status, "migration:", migration.Name)
		}
	}

	// No errors.
	//
	return applied, nil
}

func (s *service) runMigration(m *domain.Migration) (status int, err error) {

	// Nil check.
	//
	if m == nil {
		return statusError, errors.New("Migration record cannot be nil")
	}

	// Check if the migration record exist.
	//
	isExist, err := s.repository.IsMigrationRecordExist(m.Version)
	if err != nil {
		return statusError, err
	}
	if isExist {
		// Migration already applied.
		return statusExist, nil
	}

	// Parse queries.
	//
	queries, err := s.queryParser.ParseContent(m.Content)
	if err != nil {
		return statusError, err
	}

	// Execute migration queries.
	//
	startTime := time.Now()
	if err := s.repository.ExecuteQueries(queries); err != nil {
		return statusError, err
	}

	// Set execution time.
	//
	m.SetExecutionTime(startTime, time.Now())

	// Create migration record.
	//
	if err := s.repository.CreateMigrationRecord(m.MigrationRecord); err != nil {
		return statusError, err
	}

	return statusOK, nil
}
