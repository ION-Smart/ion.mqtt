package controllers

import (
	"fmt"

	m "github.com/ION-Smart/ion.mqtt/internal/models"
)

type AlertaSki struct {
	CodAlerta        int
	TipoAlerta       int
	NombreTipoAlerta string
	DescTipoAlerta   string
	CodModulo        int
	NombreModulo     string
	FechaHora        m.DateTime
	Imagen           string
	Ocupacion        int
	CodRemontador    int
	NombreRemontador string
	CodDispositivo   int
	NomDispositivo   string
	NombreZona       string
	ZoneId           string
}

func ObtenerAlertasRemontadoresSkiParam(codRemontador int, limit int) ([]AlertaSki, error) {
	var alertas []AlertaSki

	query :=
		`SELECT DISTINCT
            a.cod_alerta, a.tipo_alerta, at.nombre_tipo, at.desc_tipo, 
            a.cod_modulo, m.nombre_modulo, a.fecha_hora, a.imagen, a.ocupacion,
            a.cod_remontador, r.nombre_remontador, d.cod_dispositivo, d.nom_dispositivo, 
            z.nombre_zona, a.zoneId
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
			ski_zonas z ON (
                z.cod_zona = pv.cod_zona 
                OR z.cod_zona = r.cod_zona
            )
		LEFT JOIN
			ski_alertas_tipo at ON a.tipo_alerta = at.cod_tipo_alerta
		LEFT JOIN
			modulos m ON a.cod_modulo = m.cod_modulo
        LEFT JOIN
            dispositivos_modulos dm 
                ON dm.cod_modulo = m.cod_modulo
                AND d.cod_dispositivo = dm.cod_dispositivo
                AND dm.estado_canal != 'caducado' 
        `

	where := "WHERE dm.cod_modulo = a.cod_modulo "
	where += "AND r.cod_remontador = ? ORDER BY a.fecha_hora DESC LIMIT ?;"

	query += where

	rows, err := db.Query(query, codRemontador, limit)
	if err != nil {
		return nil, fmt.Errorf("alertas: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb AlertaSki

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
