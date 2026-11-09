// Command servidor arranca la web y la API del prototipo de WhatsApp.
//
// Uso:
//
//	go run ./cmd/servidor
//	go run ./cmd/servidor -puerto 8080 -datos datos
//
// Después se abre http://localhost:8080 en el navegador. Todo es local:
// no se conecta a WhatsApp, no se envía ningún mensaje real y nada se
// sube a git. Los datos quedan en la carpeta indicada con -datos.
//
// Pensado para un único dueño en un servidor privado: por defecto solo
// escucha en 127.0.0.1. Si se define la variable de entorno
// PROTOTIPO_TOKEN, los endpoints (salvo /salud y la web) exigen ese
// token en la cabecera X-Token.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"prototipo-whatsapp/internal/memoria"
	"prototipo-whatsapp/internal/servidor"
	"prototipo-whatsapp/internal/sesion"
)

func main() {
	direccion := flag.String("direccion", "127.0.0.1", "dirección en la que escucha el servidor")
	puerto := flag.String("puerto", "8080", "puerto en el que escucha el servidor")
	datos := flag.String("datos", "datos", "carpeta local donde se guardan los datos")
	flag.Parse()

	almacen, err := memoria.Nuevo(*datos)
	if err != nil {
		log.Fatalf("no se pudo crear el almacén de datos: %v", err)
	}

	srv := servidor.Nuevo(almacen, sesion.Nueva()).ConToken(os.Getenv("PROTOTIPO_TOKEN"))
	if os.Getenv("PROTOTIPO_TOKEN") != "" {
		log.Print("token de acceso activado para los endpoints")
	}
	escucha := *direccion + ":" + *puerto
	log.Printf("prototipo listo en http://%s (sesión simulada, sin WhatsApp real)", escucha)
	log.Fatal(http.ListenAndServe(escucha, srv.Rutas()))
}
