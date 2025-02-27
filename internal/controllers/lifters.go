package controllers

import (
	"fmt"
	"log"
	"strings"
)

type DispositivoRemontador struct {
	CodDispositivo int
	NomDispositivo string
	DeviceId       string
	Fabricante     string
	Modelo         string
	Categoria      string
}

type Remontador struct {
	CodRemontador          int
	NombreRemontador       string
	Aforo                  int
	TiempoExcedidoSegundos int
	SegundosTrayecto       int
	Plazas                 int
	Remontes               int
	Coordenadas            string
	CodZona                int
	NombreZona             string
	DispositivosStr        string
	Dispositivos           []DispositivoRemontador
}

func ObtenerRemontadores(codRemontador, codDispositivo int) ([]Remontador, error) {
	var remontadores []Remontador
	query :=
		`SELECT 
            r.cod_remontador, r.nombre_remontador, r.aforo, r.tiempo_excedido_segundos,
            r.segundos_duracion_trayecto, r.num_plazas, r.num_remontes, r.coordenadas, r.cod_zona,
            z.nombre_zona, r.dispositivos as dispositivosStr
        FROM ski_remontadores r
        LEFT JOIN ski_zonas z 
            ON z.cod_zona = r.cod_zona
        WHERE 1 `

	var values []any
	if codRemontador != 0 {
		query += "AND r.cod_remontador = ? "
		values = append(values, codRemontador)
	}

	if codDispositivo != 0 {
		query += `AND FIND_IN_SET(?, REPLACE(r.dispositivos, ";", ",")) > 0 `
		values = append(values, fmt.Sprintf("%06d", codDispositivo))
	}

	rows, err := db.Query(query, values...)
	if err != nil {
		return nil, fmt.Errorf("remontadores: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb Remontador

		if err := rows.Scan(
			&alb.CodRemontador,
			&alb.NombreRemontador,
			&alb.Aforo,
			&alb.TiempoExcedidoSegundos,
			&alb.SegundosTrayecto,
			&alb.Plazas,
			&alb.Remontes,
			&alb.Coordenadas,
			&alb.CodZona,
			&alb.NombreZona,
			&alb.DispositivosStr,
		); err != nil {
			return nil, fmt.Errorf("remontadores: %v", err)
		}

		alb.Dispositivos, err = ObtenerDispositivosRemontador(
			alb.CodRemontador,
			alb.DispositivosStr,
		)

		if err != nil {
			log.Println(err)
		}
		remontadores = append(remontadores, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("remontadores: %v", err)
	}
	return remontadores, nil
}

func ObtenerDispositivosRemontador(codRemontador int, dispositivosStr string) ([]DispositivoRemontador, error) {
	var dispositivos []DispositivoRemontador
	query :=
		fmt.Sprintf(`
        SELECT DISTINCT 
            d.cod_dispositivo, d.nom_dispositivo, d.deviceId, f.nombre_fabricante, m.nombre_modelo, c.nombre_categoria
        FROM ski_remontadores r
        INNER JOIN dispositivos d ON FIND_IN_SET(d.cod_dispositivo, REPLACE(r.dispositivos, ';', ',')) > 0 
        LEFT JOIN cloud_nx cl ON d.cod_cloud = cl.cod_cloud
        LEFT JOIN fabricantes f ON d.cod_fabricante = f.cod_fabricante
        LEFT JOIN fabricantes_modelo m ON d.cod_modelo = m.cod_modelo
        LEFT JOIN fabricantes_categoria c ON d.cod_categoria = c.cod_categoria
        LEFT JOIN dispositivos_modulos dm ON dm.cod_dispositivo = d.cod_dispositivo
        LEFT JOIN modulos ON modulos.cod_modulo = dm.cod_modulo
        LEFT JOIN sectores_verticales sv ON sv.cod_sector = modulos.cod_sector
        WHERE dm.cod_modulo = 101 AND dm.estado_canal != 'caducado'
        AND r.cod_remontador = ?
        ORDER BY FIELD(d.cod_dispositivo, '%v')`, strings.Join(strings.Split(dispositivosStr, ";"), "', '"))

	rows, err := db.Query(query, codRemontador)
	if err != nil {
		return nil, fmt.Errorf("dispositivos: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alb DispositivoRemontador

		if err := rows.Scan(
			&alb.CodDispositivo,
			&alb.NomDispositivo,
			&alb.DeviceId,
			&alb.Fabricante,
			&alb.Modelo,
			&alb.Categoria,
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

func ComprobarEnvioAlertaOcupacionRemontador(rem Remontador) bool {
	tiempoDesdeUltimaAlerta := obtenerTiempoUltimaAlertaRemontador(rem.CodRemontador)
	tiempoTimeout := rem.TiempoExcedidoSegundos

	fmt.Println(tiempoDesdeUltimaAlerta)
	insertarAlerta := false

	if tiempoDesdeUltimaAlerta > tiempoTimeout || tiempoDesdeUltimaAlerta == -1 {
		insertarAlerta = true
	}
	return insertarAlerta
}

type OcupacionSubida struct {
	CodLogOcupacion int
	Ocupacion       int
	FechaHora       string
}

func ObtenerPersonasTransportadasRemontadoresTodosDatos(
	codRemontador int,
	fechaHoraIni string,
	limit int,
) ([]OcupacionSubida, error) {
	var ocupacion []OcupacionSubida
	query := `
        SELECT 
			o.cod_log_ocupacion, o.ocupacion, o.fecha_hora
		FROM 
			analysis_ocupacion o
		LEFT JOIN
			dispositivos d ON o.cod_dispositivo = d.cod_dispositivo
		LEFT JOIN
			dispositivos_modulos dm ON d.cod_dispositivo = dm.cod_dispositivo
		LEFT JOIN 
			ski_remontadores r ON FIND_IN_SET(d.cod_dispositivo, REPLACE(r.dispositivos, ';', ',')) > 0 
		LEFT JOIN
			ski_zonas z ON r.cod_zona = z.cod_zona
		LEFT JOIN
			analysis_zona_deteccion zd ON (o.zoneId = zd.zoneId AND d.cod_dispositivo = zd.cod_dispositivo)
		LEFT JOIN
			analysis_tipo_area ta ON zd.cod_tipo_area = ta.cod_tipo_area
		WHERE r.cod_remontador IS NOT NULL 
		AND (ta.cod_tipo_area = 6 OR (ta.desc_tipo_area = 'Subida remontador' AND ta.cod_modulo = 101))
        AND dm.cod_modulo = ?
        AND dm.estado_canal != 'caducado' `
	modulo, _ := ObtenerModulo("lifters")
	var values []any
	values = append(values, modulo.CodModulo)

	if codRemontador != 0 {
		query += "AND r.cod_remontador = ? "
		values = append(values, codRemontador)
	}

	if fechaHoraIni != "" {
		query += "AND o.fecha_hora >= ? "
		values = append(values, fechaHoraIni)
	}

	query += "GROUP BY cod_remontador, cod_log_ocupacion ORDER BY o.fecha_hora DESC LIMIT ?;"
	values = append(values, limit)

	rows, err := db.Query(query, values...)
	if err != nil {
		return nil, fmt.Errorf("obtenerPersonasTransportadas query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alb OcupacionSubida

		if err := rows.Scan(
			&alb.CodLogOcupacion,
			&alb.Ocupacion,
			&alb.FechaHora,
		); err != nil {
			return nil, fmt.Errorf("obtenerPersonasTransportadas asignar valores: %v", err)
		}

		ocupacion = append(ocupacion, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("obtenerPersonasTransportadas: %v", err)
	}
	return ocupacion, nil
}
