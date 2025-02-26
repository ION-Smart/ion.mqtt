package controllers

import (
	"fmt"
	"log"
	"math"
	"strconv"

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
	fmt.Println("MessageCrowd")

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
	datos cv.MessageSecuRT,
) {
	fmt.Println("MessageSecuRT")
	fmt.Println(datos)
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

	remontadores, err := ObtenerRemontadorDispositivo(ocup.CodDispositivo)
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

	if float64(ocup.Ocupacion) >= mitad_aforo && ComprobarEnvioAlertaOcupacionRemontador(remontador) {
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
		EnviarPlazasOcupadasRemontadorSocket(remontador.CodRemontador)
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

func EnviarTiempoEsperaRemontadorSocket(codRemontador int)   {}
func EnviarPlazasOcupadasRemontadorSocket(codRemontador int) {}
func EnviarPersonasTransportadasSocket(codRemontador int)    {}
func EnviarOcupacionRemontadorSocket(codRemontador int)      {}
