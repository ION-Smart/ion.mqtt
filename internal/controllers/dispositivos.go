package controllers

import (
	"fmt"
)

type DispositivoCloud struct {
	CodDispositivo int
	NomDispositivo string
	DeviceId       string
	SystemId       string
	CloudBaseUser  string
	CloudBasePass  string
	Ip             string
	Puerto         int
}

func ObtenerDispositivosDatosCloud(codDispositivo string, deviceId string) ([]DispositivoCloud, error) {
	var dispositivos []DispositivoCloud

	query :=
		`SELECT 
            d.cod_dispositivo, d.nom_dispositivo, d.deviceId, 
            cl.systemId, cl.user, cl.password, cl.ip, cl.puerto
        FROM dispositivos d
        LEFT JOIN cloud_nx cl ON d.cod_cloud = cl.cod_cloud
        WHERE 1 AND deviceId = ? `

	rows, err := db.Query(query, deviceId)
	if err != nil {
		return nil, fmt.Errorf("dispositivos: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb DispositivoCloud

		if err := rows.Scan(
			&alb.CodDispositivo,
			&alb.NomDispositivo,
			&alb.DeviceId,
			&alb.SystemId,
			&alb.CloudBaseUser,
			&alb.CloudBasePass,
			&alb.Ip,
			&alb.Puerto,
		); err != nil {
			return nil, fmt.Errorf("dispositivos: %v", err)
		}
		dispositivos = append(dispositivos, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dispositivos: %v", err)
	}
	return dispositivos, nil
}

type ZonaDeteccion struct {
	ZoneId         string
	CodDispositivo string
	TipoArea       int
	DescTipoArea   string
	CodAlertaGest  any
	NombreAlerta   any
	CodModulo      int
	Solution       string
	CodInfraccion  any
}

func ObtenerZonasDeteccion(codDispositivo int, crowdest bool) ([]ZonaDeteccion, error) {
	var zonas []ZonaDeteccion
	query :=
		`SELECT 
            zd.zoneId, zd.cod_dispositivo, zd.cod_tipo_area, ta.desc_tipo_area,
            ta.cod_alertagest, ag.nombre_alerta, ta.cod_modulo, zd.solution, zd.cod_infraccion
        FROM analysis_zona_deteccion zd
        LEFT JOIN analysis_tipo_area ta ON zd.cod_tipo_area = ta.cod_tipo_area
        LEFT JOIN alertas_gestion ag ON ta.cod_alertagest = ag.cod_alertagest
        `

	where := "WHERE zd.cod_dispositivo = ? "

	if crowdest {
		query +=
			`
            LEFT JOIN analysis_modulos amod 
                ON (amod.cod_tipo_area = ta.cod_tipo_area AND amod.cod_modulo = ta.cod_modulo)
            LEFT JOIN analysis aly
                ON aly.cod_ai = amod.cod_ai
            `
		where += "AND aly.solution_code = 'crowd-estimation' "
	}

	query += where
	rows, err := db.Query(query, codDispositivo)
	if err != nil {
		return nil, fmt.Errorf("zonas: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb ZonaDeteccion

		if err := rows.Scan(
			&alb.ZoneId,
			&alb.CodDispositivo,
			&alb.TipoArea,
			&alb.DescTipoArea,
			&alb.CodAlertaGest,
			&alb.NombreAlerta,
			&alb.CodModulo,
			&alb.Solution,
			&alb.CodInfraccion,
		); err != nil {
			return nil, fmt.Errorf("zonas: %v", err)
		}
		zonas = append(zonas, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("zonas: %v", err)
	}
	return zonas, nil
}
