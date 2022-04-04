package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	pkgDomain "dynamodb.data-migration/internal/domain"
	pkgDynamodb "dynamodb.data-migration/internal/dynamodb"
	pkgStorage "dynamodb.data-migration/internal/filestorage"
	pkgMigration "dynamodb.data-migration/internal/migration"
	pkgParser "dynamodb.data-migration/internal/parser"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// AppVersion - application version.
var AppVersion string = "unversioned"

func main() {

	// Define our flags.
	//
	migrationContext := pkgDomain.NewMigrationContext()
	flag.StringVar(&migrationContext.MigrationsDir, "migrations", "migrations", "directory where the migration files are located")
	flag.StringVar(&migrationContext.MigrationsTable, "x-migrations-table", "x_migrations", "name of the migrations table")
	help := flag.Bool("help", false, "Display usage")
	version := flag.Bool("version", false, "Print version & exit")

	flag.Usage = usageFor(os.Args[0] + " [flags]")
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	if *help {
		fmt.Fprintf(os.Stdout, "Usage:\n")
		flag.PrintDefaults()
		return
	}
	if *version {
		appVersion := AppVersion
		if appVersion == "" {
			appVersion = "unversioned"
		}
		fmt.Println(appVersion)
		return
	}

	// Check migration context.
	//
	if err := migrationContext.Validate(); err != nil {
		flag.Usage()
		log.Fatal(err)
	}

	// Setup AWS session.
	//
	awsSession := getAwsSession()

	// Build the layers of the service "onion" from the inside out.
	//
	migrationStorage := pkgStorage.NewMigrationStorage(migrationContext.MigrationsDir)
	migrationRepository := pkgDynamodb.NewMigrationRepository(awsSession, migrationContext.MigrationsTable)
	migrationService := pkgMigration.NewMigrationService(migrationRepository, migrationStorage, pkgParser.NewQueryParser())

	// Run migrations.
	//
	log.Println("Migration started")
	applied, err := migrationService.Migrate()
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("Done", applied)
}

func usageFor(short string) func() {
	return func() {
		_, _ = fmt.Fprintf(os.Stderr, "USAGE\n")
		_, _ = fmt.Fprintf(os.Stderr, "  %s\n", short)
		_, _ = fmt.Fprintf(os.Stderr, "\n")
		_, _ = fmt.Fprintf(os.Stderr, "INFO\n")
		_, _ = fmt.Fprintf(os.Stderr, "  version:  %s\n", AppVersion)
		_, _ = fmt.Fprintf(os.Stderr, "\n")
		_, _ = fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)

		flag.VisitAll(func(f *flag.Flag) {
			_, _ = fmt.Fprintf(w, "\t-%s\t%s\t%s\n", f.Name, f.DefValue, f.Usage)
		})

		_ = w.Flush()
		_, _ = fmt.Fprintf(os.Stderr, "\n")
	}
}

func getAwsSession() *session.Session {
	// Don't use mock server in production otherwise it will override the real s3 endpoint.
	mockServerAddress := os.Getenv("AWS_MOCK_SERVER_ADDRESS")
	if len(mockServerAddress) > 0 {
		return session.Must(session.NewSession(&aws.Config{
			Endpoint:         aws.String(mockServerAddress),
			S3ForcePathStyle: aws.Bool(true), // always must be true for mock servers
		}))
	}
	return session.Must(session.NewSession())
}
