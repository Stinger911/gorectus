package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	migrationsPath = "file://migrations"
)

func main() {
	// Configure logrus
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// Parse command line flags
	var (
		up    = flag.Bool("up", false, "Run migrations up")
		down  = flag.Bool("down", false, "Run migrations down")
		force = flag.Int("force", -1, "Force migration to specific version")
		steps = flag.Int("steps", -1, "Number of migration steps to run")
		drop  = flag.Bool("drop", false, "Drop everything in database")
		reset = flag.Bool("reset", false, "Drop and recreate database")
		hash  = flag.String("hash", "", "Generate bcrypt hash for password")
	)
	flag.Parse()

	// If hash flag is provided, generate password hash and exit
	if *hash != "" {
		generatePasswordHash(*hash)
		return
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using system environment variables")
	}

	// Connect to database
	db, err := connectDB()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Create migration instance
	m, err := createMigrator(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create migrator")
	}
	defer m.Close()

	// Execute migration command based on flags
	switch {
	case *hash != "":
		generatePasswordHash(*hash)
	case *up:
		if *steps > 0 {
			runMigrationSteps(m, *steps)
		} else {
			runMigrationUp(m)
		}
	case *down:
		if *steps > 0 {
			runMigrationDown(m, *steps)
		} else {
			logrus.Error("Use -steps flag with -down to specify how many migrations to rollback")
			os.Exit(1)
		}
	case *force >= 0:
		forceMigrationVersion(m, *force)
	case *drop:
		dropDatabase(m)
	case *reset:
		resetDatabase(m)
	default:
		showMigrationStatus(m)
	}
}

func connectDB() (*sql.DB, error) {
	// Build connection string from environment variables
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD")
	dbname := getEnvOrDefault("DB_NAME", "gorectus")
	sslmode := getEnvOrDefault("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	logrus.WithFields(logrus.Fields{
		"host":   host,
		"port":   port,
		"dbname": dbname,
		"user":   user,
	}).Info("Connecting to database")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Info("Successfully connected to database")
	return db, nil
}

func createMigrator(db *sql.DB) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m, nil
}

func runMigrationUp(m *migrate.Migrate) {
	logrus.Info("Running migrations up...")

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logrus.Info("No migrations to run")
		} else {
			logrus.WithError(err).Fatal("Failed to run migrations up")
		}
	} else {
		logrus.Info("Migrations completed successfully")
	}

	showMigrationStatus(m)
}

func runMigrationDown(m *migrate.Migrate, steps int) {
	logrus.WithField("steps", steps).Info("Running migrations down...")

	if err := m.Steps(-steps); err != nil {
		if err == migrate.ErrNoChange {
			logrus.Info("No migrations to rollback")
		} else {
			logrus.WithError(err).Fatal("Failed to run migrations down")
		}
	} else {
		logrus.WithField("steps", steps).Info("Migrations rolled back successfully")
	}

	showMigrationStatus(m)
}

func runMigrationSteps(m *migrate.Migrate, steps int) {
	logrus.WithField("steps", steps).Info("Running migration steps...")

	if err := m.Steps(steps); err != nil {
		if err == migrate.ErrNoChange {
			logrus.Info("No migrations to run")
		} else {
			logrus.WithError(err).Fatal("Failed to run migration steps")
		}
	} else {
		logrus.WithField("steps", steps).Info("Migration steps completed successfully")
	}

	showMigrationStatus(m)
}

func forceMigrationVersion(m *migrate.Migrate, version int) {
	logrus.WithField("version", version).Info("Forcing migration version...")

	if err := m.Force(version); err != nil {
		logrus.WithError(err).Fatal("Failed to force migration version")
	} else {
		logrus.WithField("version", version).Info("Migration version forced successfully")
	}

	showMigrationStatus(m)
}

func dropDatabase(m *migrate.Migrate) {
	logrus.Warn("Dropping all tables in database...")

	if err := m.Drop(); err != nil {
		logrus.WithError(err).Fatal("Failed to drop database")
	} else {
		logrus.Info("Database dropped successfully")
	}
}

func resetDatabase(m *migrate.Migrate) {
	logrus.Warn("Resetting database (drop and recreate)...")

	// First drop everything
	if err := m.Drop(); err != nil {
		logrus.WithError(err).Fatal("Failed to drop database")
	}

	// Then run migrations up
	if err := m.Up(); err != nil {
		logrus.WithError(err).Fatal("Failed to run migrations after reset")
	}

	logrus.Info("Database reset completed successfully")
	showMigrationStatus(m)
}

func showMigrationStatus(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			logrus.Info("Migration status: No migrations have been run")
		} else {
			logrus.WithError(err).Warn("Failed to get migration version")
		}
		return
	}

	status := "clean"
	if dirty {
		status = "dirty"
	}

	logrus.WithFields(logrus.Fields{
		"version": version,
		"status":  status,
	}).Info("Current migration status")
}

func generatePasswordHash(password string) {
	if strings.TrimSpace(password) == "" {
		logrus.Error("Password cannot be empty")
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to generate password hash")
	}

	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Bcrypt hash: %s\n", string(hash))

	// Verify the hash works
	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		logrus.WithError(err).Fatal("Hash verification failed")
	}

	logrus.Info("Hash generated and verified successfully")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func printUsage() {
	fmt.Println("GoRectus Migration Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/migrate/main.go [flags]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -up              Run all pending migrations")
	fmt.Println("  -down -steps N   Rollback N migrations")
	fmt.Println("  -steps N         Run N migrations (used with -up)")
	fmt.Println("  -force V         Force migration to version V")
	fmt.Println("  -drop            Drop all tables")
	fmt.Println("  -reset           Drop and recreate database")
	fmt.Println("  -hash PASSWORD   Generate bcrypt hash for password")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate/main.go -up")
	fmt.Println("  go run cmd/migrate/main.go -down -steps 1")
	fmt.Println("  go run cmd/migrate/main.go -hash 'mypassword'")
	fmt.Println("  go run cmd/migrate/main.go -reset")
	fmt.Println("")
	fmt.Println("Without flags, shows current migration status.")
}
