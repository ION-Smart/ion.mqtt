package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	m "github.com/ION-Smart/ion.mqtt/internal/models"
)

type SocketPost struct {
	Message string `json:"message"`
}

type BodyAlertaSkiSocket struct {
	Alertas []AlertaSkiGet `json:"alertas"`
	Server  string         `json:"server"`
}

type BodyPlazasOcupadasRemontadorSocket struct {
	Ocupacion     OcupacionSend `json:"ocupacion"`
	CodRemontador string        `json:"cod_remontador"`
	NombreZona    string        `json:"nombre_zona"`
	Server        string        `json:"server"`
}

type BodyPersonasTransportadasSocket struct {
	Ocupacion     PersonasTransportadas `json:"ocupacion"`
	CodRemontador string                `json:"cod_remontador"`
	Server        string                `json:"server"`
}

type BodyOcupacionRestauranteTiempoRealSocket struct {
	Ocupacion OcupacionRestaurante `json:"ocupacion"`
	Cod       string               `json:"cod"`
	Server    string               `json:"server"`
}

func llamadaSocket(route string, body []byte) error {
	postUrl := fmt.Sprintf("%v/%v", SocketUrl, route)
	res, err := http.Post(
		postUrl,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	post := &SocketPost{}
	derr := json.NewDecoder(res.Body).Decode(post)
	if derr != nil {
		return derr
	}

	return nil
}

func enviarAlertaSkiSocket(codAlerta int) {
	alertas, err := ObtenerAlertasRemontadoresSkiParam(codAlerta, 0, 1)
	if err != nil {
		log.Fatalln(err)
	}

	data := BodyAlertaSkiSocket{
		Server:  "ionsmart.cat",
		Alertas: alertas,
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	err = llamadaSocket("skialertas_cambio", body)
	if err != nil {
		log.Printf("Alertas ski: %s\n", err)
	}
}

func EnviarPersonasTransportadasSocket(remontador m.Remontador) {
	ocup, err := ObtenerPersonasTransportadasHoyRemontador(remontador.CodRemontador)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := BodyPersonasTransportadasSocket{
		Server:        "ionsmart.cat",
		Ocupacion:     ocup,
		CodRemontador: fmt.Sprintf("%010d", remontador.CodRemontador),
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	err = llamadaSocket("nueva_persona_transportada", body)
	if err != nil {
		log.Printf("PersonasTransportadas: %s\n", err)
	}
}
func EnviarPlazasOcupadasRemontadorSocket(remontador m.Remontador) {
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

	err = llamadaSocket("plazas_ocupadas_remontadores", body)
	if err != nil {
		log.Printf("PlazasOcupadas: %s\n", err)
	}
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

	err = llamadaSocket("nuevo_tiempo_espera_remontadores", body)
	if err != nil {
		log.Printf("TiempoEsperaRemontador: %s\n", err)
	}
}

type BodyOcupacionTiempoRealSocket struct {
	Ocupacion OcupacionRemontadorTiempoReal `json:"ocupacion"`
	Cod       string                        `json:"cod"`
	Server    string                        `json:"server"`
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

	err = llamadaSocket("nueva_ocupacion", body)
	if err != nil {
		log.Printf("OcupacionTiempoReal: %s\n", err)
	}
}

func EnviarOcupacionRestauranteSocket(codRestaurante int) {
	ocupaciones, err := ObtenerOcupacionRestaurante(codRestaurante)
	if err != nil {
		fmt.Println(err)
		return
	} else if len(ocupaciones) <= 0 {
		fmt.Println("No hay datos a enviar")
		return
	}
	ocup := ocupaciones[0]

	data := BodyOcupacionRestauranteTiempoRealSocket{
		Server:    "ionsmart.cat",
		Cod:       fmt.Sprintf("%05d", codRestaurante),
		Ocupacion: ocup,
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	err = llamadaSocket("nueva_ocupacion_restaurante", body)
	if err != nil {
		log.Printf("ocupacionRestaurante: %s\n", err)
	}
}
