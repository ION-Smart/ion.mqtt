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

func ObtenerRemontadorDispositivo(codDispositivo int) ([]Remontador, error) {
	var remontadores []Remontador
	query :=
		`SELECT 
            r.cod_remontador, r.nombre_remontador, r.aforo, r.tiempo_excedido_segundos,
            r.segundos_duracion_trayecto, r.num_plazas, r.num_remontes, r.coordenadas, r.cod_zona,
            z.nombre_zona, r.dispositivos as dispositivosStr
        FROM ski_remontadores r
        LEFT JOIN ski_zonas z 
            ON z.cod_zona = r.cod_zona
        WHERE FIND_IN_SET(?, REPLACE(r.dispositivos, ";", ",")) > 0`

	rows, err := db.Query(query, fmt.Sprintf("%06d", codDispositivo))
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
			"DISTINCT d.cod_dispositivo, d.nom_dispositivo, d.deviceId, f.nombre_fabricante, m.nombre_modelo, c.nombre_categoria",
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

func ObtenerDispositivosRemontador(codRemontador int, selection string, dispositivosStr string) ([]DispositivoRemontador, error) {
	var dispositivos []DispositivoRemontador
	query :=
		fmt.Sprintf(`
        SELECT %v
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
        ORDER BY FIELD(d.cod_dispositivo, '%v')`,
			selection, strings.Join(strings.Split(dispositivosStr, ";"), "', '"))

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

	insertarAlerta := false

	if tiempoDesdeUltimaAlerta > tiempoTimeout || tiempoDesdeUltimaAlerta == -1 {
		insertarAlerta = true
	}
	return insertarAlerta
}

func obtenerTiempoUltimaAlertaRemontador(codRemontador int) int {
	segundosUltimaAlerta := -1
	// var alertas []AlertaSki
	//
	// ultima_alerta = ObtenerAlertasRemontadoresSkiParam(
	//     cod_remontador,
	//     1,
	// )
	//
	// if (ArrayTieneResultados(ultima_alerta)) {
	//     // fecha_hora_alerta = ultima_alerta[0].fecha_hora
	//     //
	//     // fecha_hora_alerta = new DateTime(fecha_hora_alerta)
	//     // fecha_hora_alerta.setTimezone(TIME_ZONE)
	//     //
	//     // fecha_hora_actual = new Datetime("now")
	//     // fecha_hora_actual.setTimezone(TIME_ZONE)
	//     //
	//     // diff = fecha_hora_alerta.diff(fecha_hora_actual)
	//     //
	//     // daysInSecs = diff.format("%r%a") * 24 * 60 * 60
	//     // hoursInSecs = diff.h * 60 * 60
	//     // minsInSecs = diff.i * 60
	//     //
	//     // segundosUltimaAlerta = daysInSecs + hoursInSecs + minsInSecs + diff.s
	// }

	return segundosUltimaAlerta
}
