package database

import (
	"bugforge-backend/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(cfg *config.Config) *pgxpool.Pool {
	fmt.Println("Configuration values: ", cfg)

	// Lets use pgxpool to connect to DB, First lets create a pgconnection string
	// DSN (Data Source Name) : postgress://DBUSER:DBPASS@DBHOST:DBPORT/DBNAME
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.DB_USER, cfg.DB_PASS, cfg.DB_HOST, cfg.DB_PORT, cfg.DB_NAME);
	log.Println("Connecting to dsn ", dsn)

	// Lets create a new pool connection for our app to use
	pool, err := pgxpool.New(context.TODO(), dsn) // In here, right now we are passing nil but in future for production ready application lets pass on context.Background() value
	if(err != nil) {
		log.Fatal("Database not connected", err)
	}

	log.Println("Connected to DB");
	return pool;
}