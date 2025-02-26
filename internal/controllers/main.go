package controllers

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/ION-Smart/ion.mqtt/internal/config"
	m "github.com/ION-Smart/ion.mqtt/internal/models"
	mysql "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB
var ModulosMap = make(map[string]m.Modulo)

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

	modulosDb, err := obtenerModulos()
	fmt.Println("Obtengo módulos")
	if err != nil {
		log.Fatal("Error al obtener módulos")
	}

	if len(modulosDb) > 0 {
		for _, v := range modulosDb {
			ModulosMap[v.NombreModulo] = v
		}
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

func obtenerModulos() ([]m.Modulo, error) {
	var modulos []m.Modulo

	query :=
		`SELECT 
            m.*
        FROM modulos m`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("modulos: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb m.Modulo

		if err := rows.Scan(
			&alb.CodModulo,
			&alb.Abreviacion,
			&alb.NombreModulo,
			&alb.CodSector,
		); err != nil {
			return nil, fmt.Errorf("modulos: %v", err)
		}

		modulos = append(modulos, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("modulos: %v", err)
	}
	return modulos, nil
}

func ObtenerModulo(nombreModulo string) (m.Modulo, error) {
	if val, ok := ModulosMap[nombreModulo]; ok {
		return val, nil
	}

	return m.Modulo{}, fmt.Errorf("No hay modulos que coincidan")
}
