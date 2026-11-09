// Package mensajes guarda los mensajes entrantes del prototipo.
//
// El procesamiento es interno: al recibir un mensaje, el servidor
// genera su borrador automáticamente, sin que el dueño tenga que
// pedirlo con otra llamada. En esta versión simulada los mensajes no
// llegan desde WhatsApp: POST /mensajes hace de listener de prueba
// para ejercitar ese flujo sin conectar ningún número real. Cada
// mensaje queda guardado y marcado como respondido al generar su borrador.
package mensajes

import (
	"fmt"
	"strings"
	"time"

	"prototipo-whatsapp/internal/memoria"
)

// Servicio recibe y lista mensajes entrantes.
type Servicio struct {
	almacen *memoria.Almacen
}

// Nuevo crea el servicio de mensajes sobre el almacén dado.
func Nuevo(almacen *memoria.Almacen) *Servicio {
	return &Servicio{almacen: almacen}
}

// Recibir guarda un mensaje entrante simulado. Rechaza textos vacíos.
func (s *Servicio) Recibir(de, texto string) (memoria.Mensaje, error) {
	de = strings.TrimSpace(de)
	texto = strings.TrimSpace(texto)
	if de == "" || texto == "" {
		return memoria.Mensaje{}, fmt.Errorf("el remitente y el texto no pueden estar vacíos")
	}
	mensaje := memoria.Mensaje{
		ID:         fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		De:         de,
		Texto:      texto,
		RecibidoEn: time.Now(),
	}
	if err := s.almacen.AgregarMensaje(mensaje); err != nil {
		return memoria.Mensaje{}, err
	}
	return mensaje, nil
}

// Listar devuelve todos los mensajes guardados, en orden de llegada.
func (s *Servicio) Listar() ([]memoria.Mensaje, error) {
	return s.almacen.ListarMensajes()
}

// MarcarRespondido marca un mensaje como ya respondido con un borrador.
func (s *Servicio) MarcarRespondido(id string) error {
	return s.almacen.MarcarRespondido(id)
}
