package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	m "github.com/ION-Smart/ion.mqtt/internal/models"
	"github.com/ION-Smart/ion.mqtt/internal/projectpath"
	cv "github.com/ION-Smart/ion.mqtt/pkg/cvevents"
)

func GetAnalysis() ([]m.Analysis, error) {
	var analysisTypes []m.Analysis

	rows, err := db.Query("SELECT * FROM analysis;")
	if err != nil {
		return nil, fmt.Errorf("analysisTypes: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb m.Analysis
		if err := rows.Scan(&alb.CodAi, &alb.Type, &alb.SolutionCode); err != nil {
			return nil, fmt.Errorf("analysisTypes: %v", err)
		}
		analysisTypes = append(analysisTypes, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("analysisTypes: %v", err)
	}
	return analysisTypes, nil
}

func InsertarOcupacionCrowdest(
	datos cv.MessageCrowd,
) {
	var fechaHora m.DateTime
	fechaHora.GetDateTimeFromStringMilli(datos.SystemTimestamp)

	dispositivoCloud, err := ObtenerDispositivosDatosCloud("", datos.InstanceId)
	if err != nil {
		log.Println(err)
		return
	}

	if len(dispositivoCloud) == 0 {
		log.Println("Dispositivos vacíos")
		return
	}
	disp := dispositivoCloud[0]
	zones, err := ObtenerZonasDeteccion(disp.CodDispositivo, true)
	if err != nil {
		log.Println(err)
		return
	}

	if len(zones) == 0 {
		log.Println("No hay zonas para el dispositivo")
		return
	}
	zona := zones[0]
	timestamp, err := strconv.Atoi(datos.SystemTimestamp)
	if err != nil {
		log.Println(err)
		return
	}

	ocup := m.AnalysisOcupacion{
		Ocupacion:      datos.Count,
		CodDispositivo: disp.CodDispositivo,
		Zona:           zona,
		FechaHora:      fechaHora,
		Timestamp:      timestamp,
	}
	insertarRegistroOcupacion(ocup, disp)
}

func InsertarOcupacionSecurt(
	datos cv.MessageSecuRTest,
) {
	fmt.Println("MessageSecuRT")

	var fechaHora m.DateTime
	fechaHora.GetDateTimeFromStringMilli(datos.SystemTimestamp)
	disp, err := ObtenerDispositivoDatosCloud(0, datos.Event.InstanceId)

	if err != nil {
		log.Println(err)
		return
	}
	zoneId := datos.Event.ZoneId
	if strings.Contains(datos.Event.Type, "tripwire") {
		zoneId = datos.Event.TripwireId
	}

	zona, err := ObtenerZonaDeteccion(zoneId)
	if err != nil {
		log.Println(err)
		return
	}

	timestamp, err := strconv.Atoi(datos.SystemTimestamp)
	if err != nil {
		log.Println(err)
		return
	}

	ocup := m.AnalysisOcupacion{
		Ocupacion:      datos.Event.Extra.CurrentEntries,
		CodDispositivo: disp.CodDispositivo,
		Zona:           zona,
		FechaHora:      fechaHora,
		Timestamp:      timestamp,
	}

	insertarRegistroOcupacion(ocup, disp)
}

func insertarRegistroOcupacion(ocup m.AnalysisOcupacion, disp DispositivoCloud) {
	query := `INSERT INTO analysis_ocupacion
	   (fecha_hora, ocupacion, cod_dispositivo, zoneId) VALUES (?, ?, ?, ?)`
	stmt, err := db.Prepare(query)

	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close() // Cerramos el statement

	_, err = stmt.Exec(ocup.FechaHora.Time, ocup.Ocupacion, ocup.CodDispositivo, ocup.Zona.ZoneId)
	if err != nil {
		fmt.Println(err)
	}

	// Comprobaciones aforo para saber si se inserta alerta
	comprobacionAlertaRemontador(ocup, disp)
	comprobacionAlertaTaquillas(ocup, disp)
	comprobacionAlertaRestaurante(ocup, disp)
	comprobacionAlertaParking(ocup, disp)
}

func comprobacionAlertaRemontador(ocup m.AnalysisOcupacion, disp DispositivoCloud) {
	modulo, err := ObtenerModulo("lifters")
	if err != nil {
		log.Fatal("Módulo no encontrado: ", err)
	}

	remontadores, err := ObtenerRemontadores(0, ocup.CodDispositivo)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(remontadores) < 1 {
		fmt.Println("No hay remontadores que coincidan")
		return
	}

	remontador := remontadores[0]

	mitad_aforo := math.Round(float64(remontador.Aforo) / 2.0)
	aforoElevado := math.Round(float64(remontador.Aforo) * 0.9)

	if (ocup.Zona.TipoArea == 1 || ocup.Zona.TipoArea == 2) && float64(ocup.Ocupacion) >= mitad_aforo && ComprobarEnvioAlertaOcupacionRemontador(remontador) {
		timestampMs := ocup.Timestamp

		imagenB64, err := disp.ObtenerImagen(int64(timestampMs))
		if err != nil {
			fmt.Println("Error al obtener la imagen: ", err)
			return
		}

		rutaImagenInsert := ""
		nombreImagen := fmt.Sprintf("%d_%06d.jpg", timestampMs, disp.CodDispositivo)
		rutaImagenInsert = fmt.Sprintf("fotos/%v", nombreImagen)

		rutaImagen := fmt.Sprintf("%v/%v", projectpath.Root, rutaImagenInsert)
		_ = rutaImagenInsert

		err = GuardarImagenBase64(imagenB64, rutaImagen)
		if err != nil {
			log.Fatal(err)
		}

		var tipoAlerta int
		if float64(ocup.Ocupacion) >= aforoElevado {
			tipoAlerta = 5
		} else if float64(ocup.Ocupacion) >= mitad_aforo {
			tipoAlerta = 2
		} else {
			return
		}

		alrt := m.AlertaSki{
			TipoAlerta:     tipoAlerta,
			CodModulo:      modulo.CodModulo,
			Ocupacion:      ocup.Ocupacion,
			Imagen:         rutaImagenInsert,
			CodRemontador:  remontador.CodRemontador,
			CodTaquilla:    0,
			CodRestaurante: 0,
			CodParking:     0,
			CodDispositivo: disp.CodDispositivo,
			FechaHora:      ocup.FechaHora.Time,
			ZoneId:         ocup.Zona.ZoneId,
		}
		fmt.Println(alrt)
		fmt.Printf("Alerta a insertar: mod --> %v, tipo --> %d en %v", alrt.CodModulo, alrt.TipoAlerta, alrt.ZoneId)

		err = InsertarAlertaSki(alrt)
		if err != nil {
			log.Fatalf("Error al insertar alerta: %v", err)
		}
	}

	switch ocup.Zona.TipoArea {
	case 1, 2, 6:
		EnviarTiempoEsperaRemontadorSocket(remontador.CodRemontador)
		EnviarPlazasOcupadasRemontadorSocket(remontador)
	case 3:
		// Personas transportadas
		EnviarPersonasTransportadasSocket(remontador.CodRemontador)
	case 4, 5, 7:
		// Tiempo espera taquillas
	default:
		EnviarOcupacionRemontadorSocket(remontador.CodRemontador)
	}
}

func comprobacionAlertaTaquillas(ocup m.AnalysisOcupacion, disp DispositivoCloud)   {}
func comprobacionAlertaRestaurante(ocup m.AnalysisOcupacion, disp DispositivoCloud) {}
func comprobacionAlertaParking(ocup m.AnalysisOcupacion, disp DispositivoCloud)     {}
func EnviarPersonasTransportadasSocket(codRemontador int)                           {}

type OcupacionSend struct {
	Remontes            int         `json:"num_remontes"`
	Plazas              int         `json:"num_plazas"`
	OcupacionRemontador map[int]int `json:"ocupacion_remontador"`
	PorcentajeOcupacion int         `json:"porcentaje_ocupacion_remontador"`
}

type BodyPlazasOcupadasRemontadorSocket struct {
	Ocupacion     OcupacionSend `json:"ocupacion"`
	CodRemontador string        `json:"cod_remontador"`
	NombreZona    string        `json:"nombre_zona"`
	Server        string        `json:"server"`
}

func ObtenerPlazasOcupadasRemontador(rem Remontador) (OcupacionSend, error) {
	template := map[int]int{
		0:   0,
		25:  0,
		50:  0,
		75:  0,
		100: 0,
	}
	ocupacion := OcupacionSend{
		Remontes:            rem.Remontes,
		Plazas:              rem.Plazas,
		OcupacionRemontador: template,
		PorcentajeOcupacion: 0,
	}
	ocupacion.OcupacionRemontador[0] = ocupacion.Remontes

	encontrarValueCercano := func(val int, template map[int]int) int {
		if _, ok := template[val]; ok {
			return val
		} else if val >= 100 {
			return 100
		}

		valueReturn := 100
		lastInd := 0
		for i := range template {
			if i == 0 {
				continue
			} else if valueReturn != 100 {
				continue
			} else if i > val+15 {
				valueReturn = lastInd
				continue
			}
			lastInd = i
		}
		return valueReturn
	}

	fechaHoraActual := time.Now()
	if rem.SegundosTrayecto != 0 {
		fechaHoraActual = fechaHoraActual.Add(time.Duration(-math.Abs(float64(rem.SegundosTrayecto))) * time.Second)

		fechaHoraIni := fechaHoraActual.Format(time.DateTime)
		personasTransportadas, err := ObtenerPersonasTransportadasRemontadoresTodosDatos(
			rem.CodRemontador,
			fechaHoraIni,
			rem.Remontes,
		)
		if err != nil {
			return ocupacion, err
		}

		for _, silla := range personasTransportadas {
			porcentaje := math.Round(float64(silla.Ocupacion) * 100 / float64(rem.Plazas))
			porcentajeIns := encontrarValueCercano(int(porcentaje), template)

			if porcentajeIns > 0 {
				ocupacion.OcupacionRemontador[porcentajeIns]++
				ocupacion.OcupacionRemontador[0]--
			}
		}
	}

	if ocupacion.OcupacionRemontador[0] < ocupacion.Remontes {
		ocupacion.PorcentajeOcupacion = int(
			math.Round((float64(ocupacion.OcupacionRemontador[25])*0.25 +
				float64(ocupacion.OcupacionRemontador[50])*0.50 +
				float64(ocupacion.OcupacionRemontador[75])*0.75 +
				float64(ocupacion.OcupacionRemontador[100])) /
				(float64(ocupacion.OcupacionRemontador[0]) +
					float64(ocupacion.OcupacionRemontador[25]) +
					float64(ocupacion.OcupacionRemontador[50]) +
					float64(ocupacion.OcupacionRemontador[75]) +
					float64(ocupacion.OcupacionRemontador[100])) * 100),
		)
	}

	return ocupacion, nil
}

func EnviarPlazasOcupadasRemontadorSocket(remontador Remontador) {
	ocup, err := ObtenerPlazasOcupadasRemontador(remontador)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := BodyPlazasOcupadasRemontadorSocket{
		Server:        "ionsmart.cat",
		Ocupacion:     ocup,
		CodRemontador: fmt.Sprintf("%010d", remontador.CodRemontador),
		NombreZona:    remontador.NombreZona,
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(bytes.NewBuffer(body))

	postUrl := fmt.Sprintf("%v/plazas_ocupadas_remontadores", SocketUrl)
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

type BodyTiempoEsperaRemontadorSocket struct {
	Ocupacion     TiempoEsperaRemontador `json:"ocupacion"`
	CodRemontador string                 `json:"cod_remontador"`
	NombreZona    string                 `json:"nombre_zona"`
	Server        string                 `json:"server"`
}

func EnviarTiempoEsperaRemontadorSocket(codRemontador int) {
	ocupaciones, err := ObtenerTiempoEsperaRemontador(codRemontador)
	if err != nil {
		fmt.Println(err)
		return
	} else if len(ocupaciones) <= 0 {
		fmt.Println("No hay datos a enviar")
		return
	}
	ocup := ocupaciones[0]

	data := BodyTiempoEsperaRemontadorSocket{
		Server:        "ionsmart.cat",
		Ocupacion:     ocup,
		CodRemontador: fmt.Sprintf("%010d", codRemontador),
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	postUrl := fmt.Sprintf("%v/nuevo_tiempo_espera_remontadores", SocketUrl)
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
	// fmt.Println(post)
}

type BodyOcupacionTiempoRealSocket struct {
	Ocupacion OcupacionTiempoReal `json:"ocupacion"`
	Cod       string              `json:"cod"`
	Server    string              `json:"server"`
}

func EnviarOcupacionRemontadorSocket(codRemontador int) {
	ocupaciones, err := ObtenerOcupacionRemontadorTiempoReal(codRemontador)
	if err != nil {
		fmt.Println(err)
		return
	} else if len(ocupaciones) <= 0 {
		fmt.Println("No hay datos a enviar")
		return
	}
	ocup := ocupaciones[0]

	data := BodyOcupacionTiempoRealSocket{
		Server:    "ionsmart.cat",
		Cod:       fmt.Sprintf("%010d", codRemontador),
		Ocupacion: ocup,
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	postUrl := fmt.Sprintf("%v/nueva_ocupacion", SocketUrl)
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
}

type OcupacionTiempoReal struct {
	Ocupacion        int    `json:"ocupacion"`
	Aforo            int    `json:"aforo"`
	Fecha            string `json:"fecha"`
	Hora             string `json:"hora"`
	CodRemontador    int    `json:"cod_remontador"`
	NombreRemontador string `json:"nombre_remontador"`
	CodZona          int    `json:"cod_zona"`
	NombreZona       string `json:"nombre_zona"`
	Color            string `json:"color"`
}

func ObtenerOcupacionRemontadorTiempoReal(codRemontador int) ([]OcupacionTiempoReal, error) {
	query :=
		`SELECT 
            o.ocupacion, r.aforo, CAST(o.fecha_hora AS DATE) as fecha, 
            CAST(o.fecha_hora AS TIME) as hora, r.cod_remontador, r.nombre_remontador,
            r.cod_zona, z.nombre_zona, z.color
        FROM 
            analysis_ocupacion o
        LEFT JOIN
            dispositivos d ON o.cod_dispositivo = d.cod_dispositivo
        LEFT JOIN 
            dispositivos_modulos dm ON dm.cod_dispositivo = d.cod_dispositivo
        LEFT JOIN 
            ski_remontadores r ON FIND_IN_SET(d.cod_dispositivo, REPLACE(r.dispositivos, ';', ',')) > 0 
        LEFT JOIN
            ski_zonas z ON r.cod_zona = z.cod_zona
        WHERE r.cod_remontador = ?
        AND o.fecha_hora >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
        GROUP BY CAST(o.fecha_hora AS DATE), CAST(o.fecha_hora AS TIME), r.cod_remontador, r.cod_zona
        ORDER BY o.fecha_hora DESC LIMIT 1;`
	var ocupaciones []OcupacionTiempoReal

	rows, err := db.Query(query, codRemontador)
	if err != nil {
		return nil, fmt.Errorf("ocupacionTiempoReal: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alb OcupacionTiempoReal

		if err := rows.Scan(
			&alb.Ocupacion,
			&alb.Aforo,
			&alb.Fecha,
			&alb.Hora,
			&alb.CodRemontador,
			&alb.NombreRemontador,
			&alb.CodZona,
			&alb.NombreZona,
			&alb.Color,
		); err != nil {
			return nil, fmt.Errorf("ocupacionTiempoReal: %v", err)
		}

		ocupaciones = append(ocupaciones, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ocupacionTiempoReal: %v", err)
	}
	return ocupaciones, nil
}

type TiempoEsperaStruct struct {
	Segundos     int    `json:"segundos"`
	Minutos      int    `json:"minutos"`
	CalculadoCon string `json:"calculado_con"`
}

type TiempoEsperaRemontador struct {
	CodRemontador    int                `json:"cod_remontador"`
	NombreRemontador string             `json:"nombre_remontador"`
	NombreZona       string             `json:"nombre_zona"`
	Remontes         int                `json:"num_remontes"`
	Plazas           int                `json:"num_plazas"`
	TiempoEspera     TiempoEsperaStruct `json:"tiempo_espera"`
	OcupacionEntrada int                `json:"ocupacion_entrada"`
	OcupacionEspera  int                `json:"ocupacion_espera"`
	ColaTotal        int                `json:"cola_total"`
}

type ColaRemontador struct {
	Ocupacion        int
	Aforo            int
	CodRemontador    int
	NombreRemontador string
}

func ObtenerColaRemontadoresActual(codRemontador int, tipo string) (ColaRemontador, error) {
	var cola ColaRemontador
	if codRemontador <= 0 {
		return cola, fmt.Errorf("Remontador no recibido")
	} else if tipo != "espera" && tipo != "entrada" {
		return cola, fmt.Errorf("Tipo de cola inválido")
	}
	modulo, _ := ObtenerModulo("lifters")

	// Obtenemos la fecha y hora actual - un minuto
	horaMenosMin := time.Now().Add(time.Duration(-1) * time.Minute).Format(time.DateTime)

	query := `
        SELECT 
			IFNULL(ROUND(AVG(o.ocupacion), 0), 0) as ocupacion, IFNULL(r.aforo, 0) as aforo, 
            IFNULL(r.cod_remontador, ?) as cod_remontador, IFNULL(r.nombre_remontador, '') as nombre_remontador
		FROM 
            ski_remontadores r 
		LEFT JOIN
			dispositivos d ON FIND_IN_SET(d.cod_dispositivo, REPLACE(r.dispositivos, ';', ',')) > 0 
        LEFT JOIN 
            dispositivos_modulos dm ON d.cod_dispositivo = dm.cod_dispositivo
		LEFT JOIN 
            analysis_ocupacion o ON o.cod_dispositivo = d.cod_dispositivo
		LEFT JOIN
			ski_zonas z ON r.cod_zona = z.cod_zona
		LEFT JOIN
			analysis_zona_deteccion zd ON (o.zoneId = zd.zoneId AND d.cod_dispositivo = zd.cod_dispositivo)
		LEFT JOIN
			analysis_tipo_area ta ON zd.cod_tipo_area = ta.cod_tipo_area
		WHERE r.cod_remontador IS NOT NULL 
        `

	var values []any
	values = append(values, codRemontador)

	query += fmt.Sprintf("AND ta.desc_tipo_area = 'Cola de %v' ", tipo)

	query += "AND o.fecha_hora > ? "
	values = append(values, horaMenosMin)

	query += fmt.Sprintf("AND dm.estado_canal != 'caducado' AND dm.cod_modulo = '%v' ", modulo.CodModulo)

	if codRemontador != 0 {
		query += "AND r.cod_remontador = ? "
		values = append(values, codRemontador)
	}

	rows, err := db.Query(query, values...)
	if err != nil {
		return ColaRemontador{}, fmt.Errorf("Query error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&cola.Ocupacion,
			&cola.Aforo,
			&cola.CodRemontador,
			&cola.NombreRemontador,
		); err != nil {
			return ColaRemontador{}, fmt.Errorf("%s", err)
		}
	}

	return cola, nil
}

func ObtenerTiempoEsperaRemontador(codRemontador int) ([]TiempoEsperaRemontador, error) {
	tiempos := []TiempoEsperaRemontador{}

	remontadores, err := ObtenerRemontadores(codRemontador, 0)
	if err != nil || len(remontadores) <= 0 {
		return nil, fmt.Errorf("Remontador no encontrado: %s", err)
	}

	for _, rem := range remontadores {
		t := TiempoEsperaRemontador{
			CodRemontador:    rem.CodRemontador,
			NombreRemontador: rem.NombreRemontador,
			NombreZona:       rem.NombreZona,
			Remontes:         rem.Remontes,
			Plazas:           rem.Plazas,
		}

		colaEspera, err := ObtenerColaRemontadoresActual(rem.CodRemontador, "espera")
		if err != nil {
			return nil, fmt.Errorf("Error al obtener la cola: %s", err)
		}
		colaEntrada, err := ObtenerColaRemontadoresActual(rem.CodRemontador, "entrada")
		if err != nil {
			return nil, fmt.Errorf("Error al obtener la cola: %s", err)
		}

		t.OcupacionEntrada = colaEspera.Ocupacion
		t.OcupacionEntrada = colaEntrada.Ocupacion
		t.ColaTotal = t.OcupacionEntrada + t.OcupacionEspera

		tiempos = append(tiempos, t)
	}

	return tiempos, nil
}
