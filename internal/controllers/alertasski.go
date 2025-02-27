package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	m "github.com/ION-Smart/ion.mqtt/internal/models"
)

type AlertaSkiGet struct {
	CodAlerta         int    `json:"cod_alerta"`
	TipoAlerta        int    `json:"tipo_alerta"`
	NombreTipoAlerta  string `json:"nombre_tipo"`
	DescTipoAlerta    string `json:"desc_tipo"`
	CodModulo         int    `json:"cod_modulo"`
	NombreModulo      string `json:"modulo"`
	FechaHora         string `json:"fecha_hora"`
	Imagen            string `json:"imagen"`
	Ocupacion         int    `json:"ocupacion"`
	CodRemontador     int    `json:"cod_remontador"`
	NombreRemontador  string `json:"nombre_remontador"`
	CodTaquilla       int    `json:"cod_taquilla"`
	NombreTaquilla    string `json:"nombre_taquilla"`
	NombrePV          string `json:"nombre_pv"`
	CodRestaurante    int    `json:"cod_restaurante"`
	NombreRestaurante string `json:"nombre_restaurante"`
	CodParking        int    `json:"cod_parking"`
	NombreParking     string `json:"nombre_parking"`
	CodDispositivo    int    `json:"cod_dispositivo"`
	NomDispositivo    string `json:"nom_dispositivo"`
	NombreZona        string `json:"nombre_zona"`
	ZoneId            string `json:"zoneId"`
}

func ObtenerAlertasRemontadoresSkiParam(codAlerta, codRemontador int, limit int) ([]AlertaSkiGet, error) {
	var alertas []AlertaSkiGet

	query :=
		`SELECT DISTINCT
            a.cod_alerta, a.tipo_alerta, at.nombre_tipo, at.desc_tipo, 
            a.cod_modulo, m.nombre_modulo as modulo, a.fecha_hora, a.imagen, a.ocupacion,
            a.cod_remontador, r.nombre_remontador, '0' as cod_taquilla, '' as nombre_taquilla, '' as nombre_pv,
	        '0' as cod_restaurante, '' as nombre_restaurante, '0' as cod_parking, '' as nombre_parking,
            d.cod_dispositivo, d.nom_dispositivo, z.nombre_zona, a.zoneId
		FROM 
			ski_alertas a
		LEFT JOIN
			dispositivos d ON a.cod_dispositivo = d.cod_dispositivo
		LEFT JOIN
			ski_remontadores r 
                ON (
                    a.cod_remontador = r.cod_remontador 
                    OR FIND_IN_SET(d.cod_dispositivo, REPLACE(r.dispositivos, ';', ',')) > 0
                )
		LEFT JOIN
			ski_zonas z ON z.cod_zona = r.cod_zona
		LEFT JOIN
			ski_alertas_tipo at ON a.tipo_alerta = at.cod_tipo_alerta
		LEFT JOIN
			modulos m ON a.cod_modulo = m.cod_modulo
        LEFT JOIN
            dispositivos_modulos dm 
                ON dm.cod_modulo = m.cod_modulo
                AND d.cod_dispositivo = dm.cod_dispositivo
                AND dm.estado_canal != 'caducado' `
	values := []any{}

	where := "WHERE dm.cod_modulo = a.cod_modulo "
	if codRemontador != 0 {
		where += "AND r.cod_remontador = ?"
		values = append(values, codRemontador)
	}

	if codAlerta != 0 {
		where += "AND a.cod_alerta = ?"
		values = append(values, codAlerta)
	}

	query += where

	query += " ORDER BY a.fecha_hora DESC LIMIT ?;"
	values = append(values, limit)

	rows, err := db.Query(query, values...)
	if err != nil {
		return nil, fmt.Errorf("alertas: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alb AlertaSkiGet

		if err := rows.Scan(
			&alb.CodAlerta,
			&alb.TipoAlerta,
			&alb.NombreTipoAlerta,
			&alb.DescTipoAlerta,
			&alb.CodModulo,
			&alb.NombreModulo,
			&alb.FechaHora,
			&alb.Imagen,
			&alb.Ocupacion,
			&alb.CodRemontador,
			&alb.NombreRemontador,
			&alb.CodTaquilla,
			&alb.NombreTaquilla,
			&alb.NombrePV,
			&alb.CodRestaurante,
			&alb.NombreRestaurante,
			&alb.CodParking,
			&alb.NombreParking,
			&alb.CodDispositivo,
			&alb.NomDispositivo,
			&alb.NombreZona,
			&alb.ZoneId,
		); err != nil {
			return nil, fmt.Errorf("alertas: %v", err)
		}

		alertas = append(alertas, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("alertas: %v", err)
	}

	return alertas, nil
}

func InsertarAlertaSki(alrt m.AlertaSki) error {
	var codAlerta int

	query := `INSERT INTO ski_alertas
	   (tipo_alerta, cod_modulo, fecha_hora, imagen, ocupacion, cod_remontador, cod_taquilla, cod_parking, cod_restaurante, cod_dispositivo, zoneId) 
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING cod_alerta;`
	stmt, err := db.Prepare(query)

	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(
		alrt.TipoAlerta,
		alrt.CodModulo,
		alrt.FechaHora,
		alrt.Imagen,
		alrt.Ocupacion,
		alrt.CodRemontador,
		alrt.CodTaquilla,
		alrt.CodParking,
		alrt.CodRestaurante,
		alrt.CodDispositivo,
		alrt.ZoneId,
	).Scan(&codAlerta)

	if err != nil {
		return err
	}

	fmt.Println("cod_alerta: ", codAlerta)

	go enviarAlertaSkiSocket(codAlerta)
	return nil
}

type BodyAlertaSkiSocket struct {
	Alertas []AlertaSkiGet `json:"alertas"`
	Server  string         `json:"server"`
}

type SocketPost struct {
	Message string `json:"message"`
}

func enviarAlertaSkiSocket(codAlerta int) {
	alertas, err := ObtenerAlertasRemontadoresSkiParam(codAlerta, 0, 1)
	if err != nil {
		log.Fatalln(err)
	}

	data := BodyAlertaSkiSocket{
		Server:  "ionsmart.cat",
		Alertas: alertas,
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	postUrl := fmt.Sprintf("%v/skialertas_cambio", SocketUrl)
	res, err := http.Post(
		postUrl,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		log.Fatalln(err)
	}

	post := &SocketPost{}
	derr := json.NewDecoder(res.Body).Decode(post)
	if derr != nil {
		log.Fatalln(err)
	}

	if res.StatusCode != 200 {
		fmt.Println(res.Status)
	}

	fmt.Println(post)
}

func obtenerTiempoUltimaAlertaRemontador(codRemontador int) int {
	segundosUltimaAlerta := -1
	maxInt := int(^uint(0) >> 1)

	alertas, err := ObtenerAlertasRemontadoresSkiParam(0, codRemontador, 1)

	if err != nil || len(alertas) <= 0 {
		return maxInt
	}

	ultimaAlerta := alertas[0]

	fechaHoraAlerta := ultimaAlerta.FechaHora
	timeAlerta, err := time.Parse(time.DateTime, fechaHoraAlerta)
	if err != nil {
		fmt.Printf("Error al parsear el tiempo %s\n", err)
		return maxInt
	}

	timeActual, err := time.Parse(time.DateTime, time.Now().Format(time.DateTime))
	if err != nil {
		fmt.Printf("Error al parsear el tiempo %s\n", err)
		return maxInt
	}

	segundosUltimaAlerta = int(timeActual.Unix()) - int(timeAlerta.Unix())
	return segundosUltimaAlerta
}
