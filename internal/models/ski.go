package models

type AlertaSki struct {
	CodAlerta      int
	TipoAlerta     int
	CodModulo      int
	FechaHora      string
	Imagen         string
	Ocupacion      int
	ZoneId         string
	CodRemontador  int
	CodTaquilla    int
	CodRestaurante int
	CodParking     int
	CodDispositivo int
}

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

type DispositivoRestaurante struct {
	CodDispositivo int
	NomDispositivo string
	DeviceId       string
	Fabricante     string
	Modelo         string
	Categoria      string
}

type Restaurante struct {
	CodRestaurante         int
	NombreRestaurante      string
	CodZona                int
	NombreZona             string
	Aforo                  int
	TiempoExcedidoSegundos int
	Coordenadas            string
	DispositivosStr        string
	Dispositivos           []DispositivoRestaurante
}
