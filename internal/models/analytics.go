package models

type Analysis struct {
	CodAi        string
	Type         string
	SolutionCode string
}

type AnalysisOcupacion struct {
	CodLog         string
	FechaHora      DateTime
	Timestamp      int
	Ocupacion      int
	CodDispositivo int
	ZoneId         string
}
