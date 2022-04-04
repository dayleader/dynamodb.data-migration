package dynamodb

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"dynamodb.data-migration/internal/domain"

	aws "github.com/aws/aws-sdk-go/aws"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testMigrationRepository domain.MigrationRepository
	testAwsSession          *awsSession.Session
)

func TestMain(m *testing.M) {

	// Define if we will use reaper to clean up resources.
	//
	skipReaper := os.Getenv("TESTCONTAINERS_RYUK_DISABLED") != ""
	log.Printf("flag SkipReaper = %v", skipReaper)

	// Setup Dynamodb database.
	//
	ctx := context.Background()
	exposedPort := "4566"
	req := testcontainers.ContainerRequest{
		Image:        "localstack/localstack:latest",
		ExposedPorts: []string{exposedPort},
		WaitingFor:   wait.ForListeningPort(nat.Port(exposedPort)),
		Env: map[string]string{
			"DEBUG":    "1",
			"SERVICES": "dynamodb",
		},
		SkipReaper: skipReaper, // sometimes we need to skip reaper.
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start container %v", err)
	}

	defer func() { _ = container.Terminate(ctx) }()

	// Setup test dependencies.
	//
	ip, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to start container host %v", err)
	}
	port, err := container.MappedPort(ctx, nat.Port(exposedPort))
	if err != nil {
		log.Fatalf("Failed to start container port %v", err)
	}
	mockServerAddress := fmt.Sprintf("http://%s:%s", ip, port.Port())
	fmt.Printf("Mock server address: %s\n", mockServerAddress)

	// Init AWS Session.
	//
	testAwsSession, err = awsSession.NewSession(&aws.Config{
		Endpoint:         aws.String(mockServerAddress),
		S3ForcePathStyle: aws.Bool(true), // always must be true for mock servers
	})
	if err != nil {
		log.Fatalf("Failed to start elasticsearch client %v", err)
	}

	// Init repositories.
	//
	testMigrationRepository = NewMigrationRepository(testAwsSession, "testMigrations")

	exitVal := m.Run()
	os.Exit(exitVal)
}
