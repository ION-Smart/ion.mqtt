package main

import (
	"fmt"
	"log"

	c "github.com/ION-Smart/ion.mqtt/internal/controllers"
	"github.com/ION-Smart/ion.mqtt/internal/projectpath"
)

func main() {
	err := c.InitDB()
	if err != nil {
		log.Fatalln(err)
	}

	dispositivos, err := c.ObtenerDispositivosDatosCloud("", "e5acec58-578f-8890-d21e-fc35b8bb4683")
	if err != nil {
		log.Fatalln(err)
	}

	disp := dispositivos[0]

	timestampMs := int64(1740487959000)
	if err != nil {
		log.Fatalln("Error al obtener el timestamp: ", err)
	}

	imagenB64, err := disp.ObtenerImagen(timestampMs)
	if err != nil {
		log.Fatalln("Error al obtener la imagen: ", err)
	}

	rutaImagenInsert := ""
	nombreImagen := fmt.Sprintf("%d_%06d.jpg", timestampMs, disp.CodDispositivo)
	rutaImagenInsert = fmt.Sprintf("ski/alertas/fotos/%v", nombreImagen)
	rutaImagenInsert = fmt.Sprintf("fotos/%v", nombreImagen)

	rutaImagen := fmt.Sprintf("%v/%v", projectpath.Root, rutaImagenInsert)

	// fmt.Println(rutaImagen, rutaImagenInsert, imagenB64)

	err = c.GuardarImagenBase64(imagenB64, rutaImagen)
	if err != nil {
		log.Fatal(err)
	}
}
