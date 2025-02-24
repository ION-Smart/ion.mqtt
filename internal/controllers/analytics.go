package controllers

import (
	"fmt"
	"log"

	m "github.com/ION-Smart/ion.mqtt/internal/models"
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
	zoneId := zones[0].ZoneId

	ocup := m.AnalysisOcupacion{
		Ocupacion:      datos.Count,
		CodDispositivo: disp.CodDispositivo,
		ZoneId:         zoneId,
		FechaHora:      fechaHora,
	}
	insertarRegistroOcupacion(ocup)
}

func InsertarOcupacionSecurt(
	datos cv.MessageSecuRT,
) {
	fmt.Println("MessageSecuRT")
	fmt.Println(datos)
}

func insertarRegistroOcupacion(ocup m.AnalysisOcupacion) {
	query := `INSERT INTO analysis_ocupacion
	   (fecha_hora, ocupacion, cod_dispositivo, zoneId) VALUES (?, ?, ?, ?)`
	stmt, err := db.Prepare(query)

	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close() // Cerramos el statement

	_, err = stmt.Exec(ocup.FechaHora.Time, ocup.Ocupacion, ocup.CodDispositivo, ocup.ZoneId)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Ocupación insertada en dispositivo %d\n", ocup.CodDispositivo)
}
