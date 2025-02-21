package controllers

import (
	"fmt"

	m "github.com/ION-Smart/ion.mqtt/internal/models"
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

func InsertarOcupacion(
	system_timestamp string,
	ocupacion int,
	cod_dispositivo string,
	zoneId string,
	movement string,
) {
}
