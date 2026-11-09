// Package servidor expone la web y la API del prototipo.
//
// Endpoints (todos locales, sin conexión a WhatsApp real):
//
//	GET  /salud           - comprueba que el servidor responde
//	GET  /sesion          - estado de la sesión simulada
//	POST /sesion/qr       - genera el código QR simulado
//	POST /sesion/conectar - marca la sesión como conectada (simulada)
//	POST /sesion/cerrar   - cierra la sesión, como desde el celular
//	POST /alimentar       - aprende el estilo desde un chat exportado
//	GET  /estilo          - perfil de estilo aprendido
//	GET  /mensajes        - mensajes entrantes simulados
//	POST /mensajes        - simula el listener interno: al entrar un
//	                        mensaje, el flujo interno genera su borrador
//	                        automáticamente, sin llamadas extra
//	POST /responder       - regenera el borrador para un mensaje
//	GET  /borradores      - borradores generados, pendientes de aprobar
//	GET  /                - la web del prototipo
//
// El flujo de mensajes es interno y automático en modo simulado: el
// único papel de POST /mensajes es hacer de listener de prueba (en un
// diseño con conector, ese papel lo haría el evento entrante, no una
// llamada del dueño). El borrador generado nunca se envía solo.
package servidor

import (
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"prototipo-whatsapp/internal/alimentador"
	"prototipo-whatsapp/internal/memoria"
	"prototipo-whatsapp/internal/mensajes"
	"prototipo-whatsapp/internal/respuestas"
	"prototipo-whatsapp/internal/sesion"
)

//go:embed web/index.html
var paginaWeb []byte

// Servidor reúne los servicios que usan los endpoints.
type Servidor struct {
	almacen    *memoria.Almacen
	sesion     *sesion.Sesion
	mensajes   *mensajes.Servicio
	respuestas *respuestas.Servicio
	token      string
}

// Nuevo cablea los servicios del prototipo sobre el almacén y la sesión dados.
func Nuevo(almacen *memoria.Almacen, ses *sesion.Sesion) *Servidor {
	return &Servidor{
		almacen:    almacen,
		sesion:     ses,
		mensajes:   mensajes.Nuevo(almacen),
		respuestas: respuestas.Nuevo(almacen, respuestas.Plantilla{}),
	}
}

// ConToken activa la protección de los endpoints con el token dado.
// Con el token vacío no se protege nada (modo local de un solo equipo).
// Pensado para un servidor privado de un único dueño, nunca público.
func (s *Servidor) ConToken(token string) *Servidor {
	s.token = strings.TrimSpace(token)
	return s
}

// proteger exige el token en X-Token o en Authorization: Bearer, salvo
// en /salud (sondeo) y / (la web, que pide el token al primer 401).
func (s *Servidor) proteger(siguiente http.Handler) http.Handler {
	if s.token == "" {
		return siguiente
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/salud" || r.URL.Path == "/" {
			siguiente.ServeHTTP(w, r)
			return
		}
		recibido := r.Header.Get("X-Token")
		if recibido == "" {
			recibido = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		}
		if subtle.ConstantTimeCompare([]byte(recibido), []byte(s.token)) != 1 {
			errorTexto(w, http.StatusUnauthorized, "token de acceso inválido o ausente")
			return
		}
		siguiente.ServeHTTP(w, r)
	})
}

// Rutas registra todos los endpoints y devuelve el manejador HTTP.
func (s *Servidor) Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /salud", s.salud)
	mux.HandleFunc("GET /sesion", s.verSesion)
	mux.HandleFunc("POST /sesion/qr", s.generarQR)
	mux.HandleFunc("POST /sesion/conectar", s.conectarSesion)
	mux.HandleFunc("POST /sesion/cerrar", s.cerrarSesion)
	mux.HandleFunc("POST /alimentar", s.alimentar)
	mux.HandleFunc("GET /estilo", s.verEstilo)
	mux.HandleFunc("GET /mensajes", s.listarMensajes)
	mux.HandleFunc("POST /mensajes", s.recibirMensaje)
	mux.HandleFunc("POST /responder", s.responder)
	mux.HandleFunc("GET /borradores", s.listarBorradores)
	mux.HandleFunc("GET /", s.servirWeb)
	return s.proteger(mux)
}

// escribirJSON responde con un valor en JSON y el código indicado.
func escribirJSON(w http.ResponseWriter, codigo int, valor any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(codigo)
	_ = json.NewEncoder(w).Encode(valor)
}

// errorTexto responde con un mensaje de error en JSON.
func errorTexto(w http.ResponseWriter, codigo int, mensaje string) {
	escribirJSON(w, codigo, map[string]string{"error": mensaje})
}

// leerCuerpo decodifica el JSON del cuerpo de la petición.
func leerCuerpo[T any](r *http.Request, destino *T) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(destino); err != nil {
		return fmt.Errorf("cuerpo JSON inválido: %w", err)
	}
	return nil
}

// salud responde "ok" para comprobar que el servidor está vivo.
func (s *Servidor) salud(w http.ResponseWriter, _ *http.Request) {
	escribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
}

// verSesion devuelve el estado actual de la sesión simulada.
func (s *Servidor) verSesion(w http.ResponseWriter, _ *http.Request) {
	estado, codigoQR, cerradaEn := s.sesion.Ver()
	escribirJSON(w, http.StatusOK, map[string]any{
		"estado": estado, "codigo_qr": codigoQR, "cerrada_en": cerradaEn,
		"simulada": true,
	})
}

// generarQR genera el código QR simulado y lo devuelve.
func (s *Servidor) generarQR(w http.ResponseWriter, _ *http.Request) {
	codigo := s.sesion.GenerarQR()
	escribirJSON(w, http.StatusOK, map[string]any{
		"codigo_qr": codigo, "estado": sesion.EstadoEsperandoQR, "simulada": true,
	})
}

// conectarSesion marca la sesión simulada como conectada.
func (s *Servidor) conectarSesion(w http.ResponseWriter, _ *http.Request) {
	escribirJSON(w, http.StatusOK, map[string]any{"estado": s.sesion.Conectar()})
}

// cerrarSesion cierra la sesión simulada, como desde el celular.
func (s *Servidor) cerrarSesion(w http.ResponseWriter, _ *http.Request) {
	escribirJSON(w, http.StatusOK, map[string]any{"estado": s.sesion.Cerrar()})
}

// peticionAlimentar es el cuerpo de POST /alimentar.
type peticionAlimentar struct {
	NombrePropio string `json:"nombre_propio"`
	Texto        string `json:"texto"`
}

// alimentar aprende el estilo del usuario desde un chat exportado.
func (s *Servidor) alimentar(w http.ResponseWriter, r *http.Request) {
	var peticion peticionAlimentar
	if err := leerCuerpo(r, &peticion); err != nil {
		errorTexto(w, http.StatusBadRequest, err.Error())
		return
	}
	if peticion.NombrePropio == "" || peticion.Texto == "" {
		errorTexto(w, http.StatusBadRequest, "nombre_propio y texto son obligatorios")
		return
	}
	resultado := alimentador.Analizar(peticion.NombrePropio, peticion.Texto)
	if err := s.almacen.GuardarPerfil(resultado.Perfil); err != nil {
		errorTexto(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.almacen.GuardarEjemplos(resultado.Ejemplos); err != nil {
		errorTexto(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Se relee el perfil para devolverlo con su fecha de actualización.
	perfil, err := s.almacen.CargarPerfil()
	if err != nil {
		errorTexto(w, http.StatusInternalServerError, err.Error())
		return
	}
	escribirJSON(w, http.StatusOK, perfil)
}

// verEstilo devuelve el perfil de estilo aprendido, si ya existe.
func (s *Servidor) verEstilo(w http.ResponseWriter, _ *http.Request) {
	perfil, err := s.almacen.CargarPerfil()
	if err != nil {
		if errors.Is(err, memoria.ErrSinDatos) {
			errorTexto(w, http.StatusNotFound, "aún no hay perfil: usa POST /alimentar")
			return
		}
		errorTexto(w, http.StatusInternalServerError, err.Error())
		return
	}
	escribirJSON(w, http.StatusOK, perfil)
}

// listarMensajes devuelve los mensajes entrantes simulados.
func (s *Servidor) listarMensajes(w http.ResponseWriter, _ *http.Request) {
	lista, err := s.mensajes.Listar()
	if err != nil {
		errorTexto(w, http.StatusInternalServerError, err.Error())
		return
	}
	escribirJSON(w, http.StatusOK, lista)
}

// peticionMensaje es el cuerpo de POST /mensajes.
type peticionMensaje struct {
	De    string `json:"de"`
	Texto string `json:"texto"`
}

// recibirMensaje simula el listener interno: guarda el mensaje entrante
// y, sin ninguna llamada extra, el flujo interno genera su borrador.
// El borrador queda pendiente de aprobación; no se envía nada.
func (s *Servidor) recibirMensaje(w http.ResponseWriter, r *http.Request) {
	var peticion peticionMensaje
	if err := leerCuerpo(r, &peticion); err != nil {
		errorTexto(w, http.StatusBadRequest, err.Error())
		return
	}
	mensaje, err := s.mensajes.Recibir(peticion.De, peticion.Texto)
	if err != nil {
		errorTexto(w, http.StatusBadRequest, err.Error())
		return
	}
	borrador, err := s.respuestas.Responder(mensaje.ID)
	if err != nil {
		escribirJSON(w, http.StatusCreated, map[string]any{
			"mensaje": mensaje, "flujo": "interno_automatico_simulado",
			"aviso": "mensaje guardado, pero sin borrador: " + err.Error(),
		})
		return
	}
	escribirJSON(w, http.StatusCreated, map[string]any{
		"mensaje": mensaje, "borrador": borrador,
		"flujo": "interno_automatico_simulado", "envio": "no_enviado_borrador",
	})
}

// peticionResponder es el cuerpo de POST /responder.
type peticionResponder struct {
	MensajeID string `json:"mensaje_id"`
}

// responder regenera el borrador para un mensaje ya guardado.
// El flujo normal no lo necesita: recibirMensaje ya lo genera dentro.
// Se conserva para reintentar o corregir un borrador. No envía nada.
func (s *Servidor) responder(w http.ResponseWriter, r *http.Request) {
	var peticion peticionResponder
	if err := leerCuerpo(r, &peticion); err != nil {
		errorTexto(w, http.StatusBadRequest, err.Error())
		return
	}
	borrador, err := s.respuestas.Responder(peticion.MensajeID)
	if err != nil {
		errorTexto(w, http.StatusBadRequest, err.Error())
		return
	}
	escribirJSON(w, http.StatusCreated, borrador)
}

// listarBorradores devuelve los borradores pendientes de aprobar.
func (s *Servidor) listarBorradores(w http.ResponseWriter, _ *http.Request) {
	lista, err := s.almacen.ListarBorradores()
	if err != nil {
		errorTexto(w, http.StatusInternalServerError, err.Error())
		return
	}
	escribirJSON(w, http.StatusOK, lista)
}

// servirWeb entrega la página del prototipo.
func (s *Servidor) servirWeb(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(paginaWeb)
}
