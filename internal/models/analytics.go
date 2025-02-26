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
	Zona           ZonaDeteccion
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
