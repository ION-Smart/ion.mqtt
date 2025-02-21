package controllers

import (
	"database/sql"
	"log"

	"github.com/ION-Smart/ion.mqtt/internal/config"
	mysql "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func init() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}
}

func InitDB() error {
	var err error

	conf := config.New()
	db, err = GetConnection(conf)

	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func GetConnection(conf *config.Config) (*sql.DB, error) {
	cfg := mysql.Config{
		User:   conf.Database.User,
		Passwd: conf.Database.Pass,
		Net:    "tcp",
		Addr:   conf.Database.Host,
		DBName: conf.Database.Name,

		AllowNativePasswords: true,
	}

	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
		return nil, pingErr
	}
	return db, nil
}
