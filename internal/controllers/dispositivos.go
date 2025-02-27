package controllers

import (
	// "encoding/base64"
	"fmt"
	// "time"

	m "github.com/ION-Smart/ion.mqtt/internal/models"
	nx "github.com/ION-Smart/ion.mqtt/pkg/nxwitness"
)

type DispositivoCloud struct {
	CodDispositivo int
	NomDispositivo string
	DeviceId       string
	Cloud          *nx.NxCloud
	systemId       string
	cloudBaseUser  string
	cloudBasePass  string
	ip             string
	puerto         int
	server         string
}

func (disp *DispositivoCloud) ObtenerImagen(timestamp int64) (string, error) {
	return nx.GetDeviceThumbnailB64(disp.Cloud, "jpg", disp.DeviceId, timestamp)
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
		return nil, fmt.Errorf("dispositivosDatosCloud: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb DispositivoCloud

		if err := rows.Scan(
			&alb.CodDispositivo,
			&alb.NomDispositivo,
			&alb.DeviceId,
			&alb.systemId,
			&alb.cloudBaseUser,
			&alb.cloudBasePass,
			&alb.ip,
			&alb.puerto,
		); err != nil {
			return nil, fmt.Errorf("dispositivosDatosCloud: %v", err)
		}

		var cerr error
		alb.Cloud, cerr = nx.NewNxCloud(
			alb.systemId,
			alb.cloudBaseUser,
			alb.cloudBasePass,
			alb.ip,
			alb.puerto,
		)
		if cerr != nil {
			return nil, fmt.Errorf("dispositivosDatosCloud: %v", cerr)
		}

		dispositivos = append(dispositivos, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dispositivosDatosCloud: %v", err)
	}
	return dispositivos, nil
}

func ObtenerDispositivoDatosCloud(codDispositivo int, deviceId string) (DispositivoCloud, error) {
	var dispositivo DispositivoCloud

	if codDispositivo <= 0 && deviceId == "" {
		return DispositivoCloud{}, fmt.Errorf("Parámetros insuficientes")
	}

	query :=
		`SELECT 
            d.cod_dispositivo, d.nom_dispositivo, d.deviceId, 
            cl.systemId, cl.user, cl.password, cl.ip, cl.puerto
        FROM dispositivos d
        LEFT JOIN cloud_nx cl ON d.cod_cloud = cl.cod_cloud
        WHERE 1 `
	var values []any

	if deviceId != "" {
		query += "AND deviceId = ? "
		values = append(values, deviceId)
	}

	if codDispositivo != 0 {
		query += "AND cod_dispositivo = ? "
		values = append(values, deviceId)
	}

	rows, err := db.Query(query, values...)
	if err != nil {
		return DispositivoCloud{}, fmt.Errorf("dispositivoDatosCloud: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alb DispositivoCloud
		if err := rows.Scan(
			&alb.CodDispositivo,
			&alb.NomDispositivo,
			&alb.DeviceId,
			&alb.systemId,
			&alb.cloudBaseUser,
			&alb.cloudBasePass,
			&alb.ip,
			&alb.puerto,
		); err != nil {
			return DispositivoCloud{}, fmt.Errorf("dispositivoDatosCloud: %v", err)
		}

		var cerr error
		alb.Cloud, cerr = nx.NewNxCloud(
			alb.systemId,
			alb.cloudBaseUser,
			alb.cloudBasePass,
			alb.ip,
			alb.puerto,
		)
		if cerr != nil {
			return DispositivoCloud{}, fmt.Errorf("dispositivoDatosCloud: %v", cerr)
		}

		dispositivo = alb
	}

	if err := rows.Err(); err != nil {
		return DispositivoCloud{}, fmt.Errorf("dispositivoDatosCloud: %v", err)
	}
	return dispositivo, nil
}

func ObtenerZonasDeteccion(codDispositivo int, crowdest bool) ([]m.ZonaDeteccion, error) {
	var zonas []m.ZonaDeteccion
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
		var alb m.ZonaDeteccion

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
func ObtenerZonaDeteccion(zoneId string) (m.ZonaDeteccion, error) {
	zonaVacia := m.ZonaDeteccion{}
	zona := zonaVacia

	if zoneId == "" {
		return zonaVacia, fmt.Errorf("No se ha recibido id de la zona")
	}

	query :=
		`SELECT 
            zd.zoneId, zd.cod_dispositivo, zd.cod_tipo_area, ta.desc_tipo_area,
            ta.cod_alertagest, ag.nombre_alerta, ta.cod_modulo, zd.solution, zd.cod_infraccion
        FROM analysis_zona_deteccion zd
        LEFT JOIN analysis_tipo_area ta ON zd.cod_tipo_area = ta.cod_tipo_area
        LEFT JOIN alertas_gestion ag ON ta.cod_alertagest = ag.cod_alertagest
        WHERE zd.zoneId = ? `

	rows, err := db.Query(query, zoneId)
	if err != nil {
		return zonaVacia, fmt.Errorf("zona: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb m.ZonaDeteccion

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
			return zonaVacia, fmt.Errorf("zona: %v", err)
		}
		zona = alb
	}

	if err := rows.Err(); err != nil {
		return zonaVacia, fmt.Errorf("zona: %v", err)
	}

	if zona == zonaVacia {
		return zonaVacia, fmt.Errorf("zona no encontrada")
	}

	return zona, nil
}
