package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/AbrahamMayowa/ticketmania/internal/data"
	"github.com/AbrahamMayowa/ticketmania/internal/jsonlog"
	"github.com/AbrahamMayowa/ticketmania/internal/mailer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)


type db struct {
	dsn string
}

type config struct {
	port int
	env  string
	db   db
	jwt  struct {
		secret string
	}
	mailerConfig mailer.Config
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	wg     sync.WaitGroup
	mailer mailer.Mailer
}

func init() {
	// Load .env file automatically on startup
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Println(err.Error())
			panic(err)
		}
	}

}

func main() {

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("Configuration error:", err)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(*cfg)

	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: *cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: *mailer.New(cfg.mailerConfig),
	}

	err = app.server()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

}

func loadConfig() (*config, error) {
	cfg := &config{
		port: getEnvAsInt("PORT", 4000),
		env:  getEnv("APP_ENV", "development"),
	}

	// Database configuration
	cfg.db.dsn = os.Getenv("DATABASE_URL")
	if cfg.db.dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// JWT configuration
	cfg.jwt.secret = os.Getenv("HASH_SECRET_KEY")
	if cfg.jwt.secret == "" {
		return nil, fmt.Errorf("HASH_SECRET_KEY is required")
	}

	// Mailer configuration
	mailerPort, err := strconv.Atoi(os.Getenv("MAILTRAP_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid MAILTRAP_PORT: %w", err)
	}

	cfg.mailerConfig = mailer.Config{
		Host:     getEnv("MAILTRAP_HOST", ""),
		Port:     mailerPort,
		Username: os.Getenv("MAILTRAP_USERNAME"),
		Sender:   getEnv("MAILTRAP_SENDER", ""),
		Password: os.Getenv("MAILTRAP_PASSWORD"),
	}

	// Validate mailer config
	if cfg.mailerConfig.Host == "" {
		return nil, fmt.Errorf("MAILTRAP_HOST is required")
	}
	if cfg.mailerConfig.Username == "" {
		return nil, fmt.Errorf("MAILTRAP_USERNAME is required")
	}
	if cfg.mailerConfig.Password == "" {
		return nil, fmt.Errorf("MAILTRAP_PASSWORD is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid integer value for %s, using default %d", key, defaultValue)
		return defaultValue
	}
	
	return value
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	// Use PingContext() to establish a new connection to the database
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil

}
