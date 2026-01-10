package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	"github.com/AbrahamMayowa/ticketmania/internal/data"
	"github.com/AbrahamMayowa/ticketmania/internal/jsonlog"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"sync"
	"fmt"
)

const version = "1.0.0"

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
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	wg     sync.WaitGroup
}

func init() {
	// Load .env file automatically on startup
	err := godotenv.Load()
	if err != nil {
		log.Println(err.Error())
		panic(err)
	}
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	displayVersion := flag.Bool("version", false, "Display version value")
	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	cfg.db = db{
		dsn: os.Getenv("DATABASE_URL"),
	}

	cfg.jwt = struct{ secret string }{
		secret: os.Getenv("HASH_SECRET_KEY"),
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)

	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	err = app.server()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

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
