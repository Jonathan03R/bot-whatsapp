// Package sesion representa el estado de la sesión del prototipo.
//
// IMPORTANTE: es una sesión SIMULADA. No se conecta a WhatsApp, no lee
// ningún chat real y el código QR que genera es solo texto de demostración,
// no un QR válido de WhatsApp. Sirve para probar el flujo completo —
// generar QR, conectar, cerrar — con las mismas variables y endpoints que
// tendría la versión real, sin arriesgar el número personal.
//
// El control que pedía el usuario está aquí: la sesión se puede cerrar en
// cualquier momento con POST /sesion/cerrar, igual que se cierra un
// dispositivo vinculado desde el celular. Cuando se enchufa el conector
// oficial de WhatsApp Business, este paquete es el que cambia: el resto
// (alimentador, memoria, respuestas) sigue igual.
package sesion

import (
	"fmt"
	"sync"
	"time"
)

// Estado describe en qué punto está la sesión simulada.
type Estado string

// Estados posibles de la sesión.
const (
	EstadoDesconectada Estado = "desconectada"
	EstadoEsperandoQR  Estado = "esperando_qr"
	EstadoConectada    Estado = "conectada"
	EstadoCerrada      Estado = "cerrada"
)

// Sesion guarda el estado actual de la sesión simulada.
type Sesion struct {
	mu        sync.Mutex
	estado    Estado
	id        string
	codigoQR  string
	creadaEn  time.Time
	cerradaEn *time.Time
}

// Nueva crea una sesión en estado desconectada.
func Nueva() *Sesion {
	return &Sesion{estado: EstadoDesconectada}
}

// GenerarQR pone la sesión en espera y devuelve un código simulado.
// No es un QR real: es solo un texto para probar la pantalla de vinculación.
func (s *Sesion) GenerarQR() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.estado = EstadoEsperandoQR
	s.codigoQR = fmt.Sprintf("SIMULADO-%d", time.Now().UnixNano())
	s.creadaEn = time.Now()
	s.cerradaEn = nil
	return s.codigoQR
}

// Conectar marca la sesión como conectada, solo si estaba esperando el QR.
func (s *Sesion) Conectar() Estado {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.estado != EstadoEsperandoQR {
		return s.estado
	}
	s.estado = EstadoConectada
	s.id = fmt.Sprintf("sesion-%d", time.Now().UnixNano())
	return s.estado
}

// Cerrar corta la sesión en cualquier momento, como cerrar el dispositivo
// vinculado desde el celular. Después de cerrar, no se recibe nada más.
func (s *Sesion) Cerrar() Estado {
	s.mu.Lock()
	defer s.mu.Unlock()
	ahora := time.Now()
	s.estado = EstadoCerrada
	s.cerradaEn = &ahora
	s.codigoQR = ""
	return s.estado
}

// Ver devuelve el estado actual, el código QR si lo hay y la hora de cierre.
func (s *Sesion) Ver() (Estado, string, *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.estado, s.codigoQR, s.cerradaEn
}
