package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rammyblog/authentication-service/api/data"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

var counts int64

type Config struct {
	Repo   data.Repository
	Client *http.Client
}

func main() {
	log.Println("Starting authentication service")

	conn := connectToDB()

	if conn == nil {
		log.Panic("Can't connect to DB")
	}

	app := Config{
		Client: &http.Client{},
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Panicln("Postgres not ready")
			counts++
		} else {
			log.Println("Connected to Postgres")
			return connection
		}
		if counts > 10 {
			log.Println(err)
			return nil
		}
		log.Println("back off for two secs")

		time.Sleep(2 * time.Second)
		continue
	}
}

func (app *Config) setupRepo(conn *sql.DB) {
	db := data.NewPostgresRepository(conn)

	app.Repo = db
}
