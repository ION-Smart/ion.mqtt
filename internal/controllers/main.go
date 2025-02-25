package controllers

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"

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

func ArrayTieneResultados(datos []any) bool {
	if len(datos) > 0 {
		return true
	}
	return false
}

func GuardarImagenBase64(imagenB64 string, rutaImagen string) error {
	dec, err := base64.StdEncoding.DecodeString(imagenB64)
	if err != nil {
		return fmt.Errorf("Error al guardar imagen: %v", err)
	}

	f, err := os.Create(rutaImagen)
	if err != nil {
		return fmt.Errorf("Error al guardar imagen: %v", err)
	}
	defer f.Close()

	if _, err := f.Write(dec); err != nil {
		return fmt.Errorf("Error al guardar imagen: %v", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("Error al guardar imagen: %v", err)
	}

	return nil
}
