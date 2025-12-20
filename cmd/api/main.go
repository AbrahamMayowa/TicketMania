package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/AbrahamMayowa/ticketmania/internal/data"
)

const version = "1.0.0"

type db struct {
	dsn string
}
 

type config struct {
	port int
	env string
	db db
	jwt struct {
	secret string
}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
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
	flag.Parse()

	cfg.db = db{
		dsn: os.Getenv("DATABASE_URL"),
	}

	cfg.jwt = struct{secret string}{
		secret: os.Getenv("HASH_SECRET_KEY"),
	}

	logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

	db, err := openDB(cfg)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Printf("Database has connected!")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.port),
		Handler: app.router(),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout:30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr) 
	err = srv.ListenAndServe()
	logger.Fatal(err)

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



