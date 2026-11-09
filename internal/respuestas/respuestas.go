// Package respuestas genera borradores de respuesta en el estilo del usuario.
//
// Regla principal: cuando falta un dato, el borrador NO inventa nada.
// Responde como contestaría una persona que no sabe: "déjame ver y te
// confirmo". Los borradores nunca se envían solos: quedan guardados para
// que el usuario los apruebe o corrija.
//
// El generador es una interfaz (Generador) para que mañana se pueda enchufar
// un modelo de IA por API sin tocar el resto: la versión actual, Plantilla,
// genera con reglas simples a partir del perfil de estilo.
package respuestas

import (
	"fmt"
	"strings"
	"time"

	"prototipo-whatsapp/internal/memoria"
)

// Generador produce el texto de una respuesta y la espera sugerida, en
// segundos, para contestar con ritmo humano y no al instante.
type Generador interface {
	Generar(perfil memoria.PerfilEstilo, entrante string) (texto string, esperaSegundos int)
}

// Plantilla es el generador por reglas del prototipo. No llama a ninguna IA.
type Plantilla struct{}

// Generar elige una respuesta corta según el tipo de mensaje entrante.
func (Plantilla) Generar(perfil memoria.PerfilEstilo, entrante string) (string, int) {
	saludo := "hola"
	if len(perfil.Saludos) > 0 {
		saludo = perfil.Saludos[0]
	}
	switch clasificar(entrante) {
	case "saludo":
		return saludo + ", ¿qué tal?", esperaSugerida(entrante)
	case "agradecimiento":
		return "de nada, cualquier cosa me avisas", esperaSugerida(entrante)
	case "pregunta":
		// Sin datos en memoria no se inventa una respuesta: se pide tiempo.
		return "déjame ver y te confirmo", esperaSugerida(entrante)
	default:
		if perfil.UsaYa {
			return "ya, te entiendo, déjame ver y te aviso", esperaSugerida(entrante)
		}
		return "te entiendo, déjame ver y te aviso", esperaSugerida(entrante)
	}
}

// clasificar distingue saludos, agradecimientos, preguntas y lo demás.
func clasificar(texto string) string {
	bajo := strings.ToLower(strings.TrimSpace(texto))
	if contieneAlguna(bajo, []string{"hola", "qué tal", "que tal", "buenas"}) {
		return "saludo"
	}
	if strings.Contains(bajo, "gracias") {
		return "agradecimiento"
	}
	if contieneAlguna(bajo, []string{"?", "puedes", "podrías", "podrias", "mañana", "hoy", "cuándo", "cuando", "a qué hora"}) {
		return "pregunta"
	}
	return "otro"
}

// contieneAlguna dice si el texto contiene alguna de las palabras dadas.
func contieneAlguna(texto string, palabras []string) bool {
	for _, palabra := range palabras {
		if strings.Contains(texto, palabra) {
			return true
		}
	}
	return false
}

// esperaSugerida propone una demora humana (entre 20 y 69 segundos)
// derivada del largo del mensaje, para no contestar siempre al instante.
func esperaSugerida(texto string) int {
	return 20 + len([]rune(texto))%50
}

// Servicio genera borradores para mensajes ya guardados.
type Servicio struct {
	almacen   *memoria.Almacen
	generador Generador
}

// Nuevo crea el servicio de respuestas con el generador dado.
func Nuevo(almacen *memoria.Almacen, generador Generador) *Servicio {
	return &Servicio{almacen: almacen, generador: generador}
}

// Responder genera y guarda el borrador para un mensaje, y lo marca respondido.
func (s *Servicio) Responder(mensajeID string) (memoria.Borrador, error) {
	mensajes, err := s.almacen.ListarMensajes()
	if err != nil {
		return memoria.Borrador{}, err
	}
	var encontrado *memoria.Mensaje
	for i := range mensajes {
		if mensajes[i].ID == mensajeID {
			encontrado = &mensajes[i]
		}
	}
	if encontrado == nil {
		return memoria.Borrador{}, fmt.Errorf("no hay ningún mensaje con id %s", mensajeID)
	}
	perfil, err := s.almacen.CargarPerfil()
	if err != nil {
		// Sin perfil aprendido se responde con el estilo neutro por defecto.
		perfil = memoria.PerfilEstilo{}
	}
	texto, espera := s.generador.Generar(perfil, encontrado.Texto)
	borrador := memoria.Borrador{
		ID:                     fmt.Sprintf("borrador-%d", time.Now().UnixNano()),
		MensajeID:              mensajeID,
		Texto:                  texto,
		EsperaSugeridaSegundos: espera,
		CreadoEn:               time.Now(),
	}
	if err := s.almacen.GuardarBorrador(borrador); err != nil {
		return memoria.Borrador{}, err
	}
	if err := s.almacen.MarcarRespondido(mensajeID); err != nil {
		return memoria.Borrador{}, err
	}
	return borrador, nil
}
