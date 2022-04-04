package filestorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"dynamodb.data-migration/internal/domain"
)

type storage struct {
	migrationsDir string
	pattern       *regexp.Regexp
}

// NewMigrationStorage creates a service with necessary dependencies.
func NewMigrationStorage(migrationsDir string) domain.MigrationStorage {
	return &storage{
		migrationsDir: migrationsDir,
		pattern:       regexp.MustCompile(domain.MigrationFilePattern),
	}
}

func (s *storage) GetExecutableMigrations() ([]*domain.Migration, error) {
	var migrations []*domain.Migration
	err := filepath.Walk(s.migrationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Error walking filepath: %v", err)
		}
		if info != nil && info.IsDir() {
			return nil
		}
		match := s.pattern.FindStringSubmatch(path)
		if match == nil {
			return fmt.Errorf("File is ignored, naming pattern is wrong: %s", path)
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		name, err := getTitle(match, 2, path)
		if err != nil {
			return err
		}
		major, err := getVersion(match, 3, path)
		if err != nil {
			return err
		}
		minor, err := getVersion(match, 4, path)
		if err != nil {
			return err
		}
		patch, err := getVersion(match, 5, path)
		if err != nil {
			return err
		}
		migration := &domain.Migration{
			MigrationRecord: domain.MigrationRecord{
				Version: domain.Version{
					Major: major,
					Minor: minor,
					Patch: patch,
				},
				Name: name,
			},
			Content: content,
		}
		migrations = append(migrations, migration)
		return nil
	})
	return migrations, err
}

func getTitle(match []string, i int, filename string) (string, error) {
	if len(match) > i {
		return match[i], nil
	}
	return "", fmt.Errorf("Incorrect file naming pattern: %s", filename)
}

func getVersion(match []string, i int, filename string) (int, error) {
	if len(match) > i {
		return strconv.Atoi(match[i])
	}
	return 0, fmt.Errorf("Incorrect file versioning pattern: %s", filename)
}
