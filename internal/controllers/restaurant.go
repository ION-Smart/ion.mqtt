package controllers

import (
	"fmt"
	"log"
	"strings"

	m "github.com/ION-Smart/ion.mqtt/internal/models"
)

func ObtenerRestaurantes(codRestaurante, codDispositivo int) ([]m.Restaurante, error) {
	var restaurantes []m.Restaurante
	query :=
		`SELECT 
            r.cod_restaurante, r.nombre_restaurante, r.cod_zona, z.nombre_zona, r.aforo, 
            r.tiempo_excedido_segundos, r.coordenadas, r.dispositivos as dispositivosStr
        FROM restaurante r
        LEFT JOIN ski_zonas z 
            ON z.cod_zona = r.cod_zona
        WHERE 1 `

	var values []any
	if codRestaurante != 0 {
		query += "AND r.cod_restaurante = ? "
		values = append(values, codRestaurante)
	}

	if codDispositivo != 0 {
		query += `AND FIND_IN_SET(?, REPLACE(r.dispositivos, ";", ",")) > 0 `
		values = append(values, fmt.Sprintf("%06d", codDispositivo))
	}

	rows, err := db.Query(query, values...)
	if err != nil {
		return nil, fmt.Errorf("restaurantes: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb m.Restaurante

		if err := rows.Scan(
			&alb.CodRestaurante,
			&alb.NombreRestaurante,
			&alb.CodZona,
			&alb.NombreZona,
			&alb.Aforo,
			&alb.TiempoExcedidoSegundos,
			&alb.Coordenadas,
			&alb.DispositivosStr,
		); err != nil {
			return nil, fmt.Errorf("restaurantes: %v", err)
		}

		alb.Dispositivos, err = ObtenerDispositivosRestaurante(
			alb.CodRestaurante,
			alb.DispositivosStr,
		)

		if err != nil {
			log.Println(err)
		}
		restaurantes = append(restaurantes, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("restaurantes: %v", err)
	}
	return restaurantes, nil
}

func ObtenerDispositivosRestaurante(codRestaurante int, dispositivosStr string) ([]m.DispositivoRestaurante, error) {
	var dispositivos []m.DispositivoRestaurante
	query :=
		fmt.Sprintf(`
        SELECT DISTINCT 
            d.cod_dispositivo, d.nom_dispositivo, d.deviceId, f.nombre_fabricante, m.nombre_modelo, c.nombre_categoria
        FROM restaurante r
        INNER JOIN dispositivos d ON FIND_IN_SET(d.cod_dispositivo, REPLACE(r.dispositivos, ';', ',')) > 0 
        LEFT JOIN cloud_nx cl ON d.cod_cloud = cl.cod_cloud
        LEFT JOIN fabricantes f ON d.cod_fabricante = f.cod_fabricante
        LEFT JOIN fabricantes_modelo m ON d.cod_modelo = m.cod_modelo
        LEFT JOIN fabricantes_categoria c ON d.cod_categoria = c.cod_categoria
        LEFT JOIN dispositivos_modulos dm ON dm.cod_dispositivo = d.cod_dispositivo
        LEFT JOIN modulos ON modulos.cod_modulo = dm.cod_modulo
        LEFT JOIN sectores_verticales sv ON sv.cod_sector = modulos.cod_sector
        WHERE dm.cod_modulo = 105 AND dm.estado_canal != 'caducado'
        AND r.cod_restaurante = ?
        ORDER BY FIELD(d.cod_dispositivo, '%v')
        `, strings.Join(strings.Split(dispositivosStr, ";"), "', '"))

	rows, err := db.Query(query, codRestaurante)
	if err != nil {
		return nil, fmt.Errorf("dispositivos: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alb m.DispositivoRestaurante

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

func ComprobarEnvioAlertaOcupacionRestaurante(rest m.Restaurante) bool {
	tiempoDesdeUltimaAlerta := obtenerTiempoUltimaAlertaRestaurante(rest.CodRestaurante)
	tiempoTimeout := rest.TiempoExcedidoSegundos

	fmt.Println(tiempoDesdeUltimaAlerta)
	insertarAlerta := false

	if tiempoDesdeUltimaAlerta > tiempoTimeout || tiempoDesdeUltimaAlerta == -1 {
		insertarAlerta = true
	}
	return insertarAlerta
}
