// Package memoria guarda los datos del prototipo en archivos JSON locales.
//
// Es la base de datos embebida del prototipo: no necesita un servidor
// externo, los datos viven en la carpeta de datos y se escriben de forma
// atómica (archivo temporal + renombre) para no dejar archivos a medias.
// Guarda el perfil de estilo, los ejemplos, los mensajes simulados y los
// borradores de respuesta. Si más adelante se cambia a SQLite, solo cambia
// este paquete: el resto del código usa sus métodos, no los archivos.
package memoria

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ErrSinDatos indica que todavía no hay nada guardado en ese archivo.
var ErrSinDatos = errors.New("todavía no hay datos guardados")

// PerfilEstilo resume cómo escribe el usuario, aprendido de sus mensajes.
type PerfilEstilo struct {
	NombrePropio       string    `json:"nombre_propio"`
	MensajesAnalizados int       `json:"mensajes_analizados"`
	LongitudMedia      float64   `json:"longitud_media"`
	UsaYa              bool      `json:"usa_ya"`
	UsaOye             bool      `json:"usa_oye"`
	Saludos            []string  `json:"saludos"`
	PalabrasFrecuentes []string  `json:"palabras_frecuentes"`
	FrecuenciaEmoji    float64   `json:"frecuencia_emoji"`
	ActualizadoEn      time.Time `json:"actualizado_en"`
}

// Ejemplo es un par mensaje entrante -> respuesta del usuario.
type Ejemplo struct {
	Entrante  string `json:"entrante"`
	Respuesta string `json:"respuesta"`
}

// Mensaje es un mensaje entrante simulado, pegado a mano o de ejemplo.
type Mensaje struct {
	ID         string    `json:"id"`
	De         string    `json:"de"`
	Texto      string    `json:"texto"`
	RecibidoEn time.Time `json:"recibido_en"`
	Respondido bool      `json:"respondido"`
}

// Borrador es una respuesta propuesta. Nunca se envía sola: el usuario la aprueba.
type Borrador struct {
	ID                     string    `json:"id"`
	MensajeID              string    `json:"mensaje_id"`
	Texto                  string    `json:"texto"`
	EsperaSugeridaSegundos int       `json:"espera_sugerida_segundos"`
	CreadoEn               time.Time `json:"creado_en"`
}

// Almacen es el acceso a los archivos de datos. Un solo mutex protege
// todas las lecturas y escrituras para mantenerlas coherentes.
type Almacen struct {
	dir string
	mu  sync.Mutex
}

// Nuevo crea el almacén en la carpeta indicada, creándola si hace falta.
func Nuevo(dir string) (*Almacen, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Almacen{dir: dir}, nil
}

// ruta devuelve la ruta completa de un archivo de datos.
func (a *Almacen) ruta(nombre string) string {
	return filepath.Join(a.dir, nombre)
}

// escribirJSON guarda un valor como JSON con escritura atómica.
func escribirJSON[T any](ruta string, valor T) error {
	datos, err := json.MarshalIndent(valor, "", "  ")
	if err != nil {
		return err
	}
	tmp := ruta + ".tmp"
	if err := os.WriteFile(tmp, datos, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, ruta)
}

// leerJSON lee un valor JSON. Si el archivo no existe, devuelve ErrSinDatos.
func leerJSON[T any](ruta string, destino *T) error {
	datos, err := os.ReadFile(ruta)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrSinDatos
		}
		return err
	}
	return json.Unmarshal(datos, destino)
}

// GuardarPerfil guarda el perfil de estilo aprendido.
func (a *Almacen) GuardarPerfil(p PerfilEstilo) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	p.ActualizadoEn = time.Now()
	return escribirJSON(a.ruta("perfil.json"), p)
}

// CargarPerfil lee el perfil guardado, o ErrSinDatos si aún no hay uno.
func (a *Almacen) CargarPerfil() (PerfilEstilo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var p PerfilEstilo
	if err := leerJSON(a.ruta("perfil.json"), &p); err != nil {
		return PerfilEstilo{}, err
	}
	return p, nil
}

// GuardarEjemplos guarda los pares entrante -> respuesta del usuario.
func (a *Almacen) GuardarEjemplos(ejemplos []Ejemplo) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return escribirJSON(a.ruta("ejemplos.json"), ejemplos)
}

// CargarEjemplos lee los ejemplos; si no hay, devuelve una lista vacía.
func (a *Almacen) CargarEjemplos() ([]Ejemplo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var ejemplos []Ejemplo
	if err := leerJSON(a.ruta("ejemplos.json"), &ejemplos); err != nil {
		if errors.Is(err, ErrSinDatos) {
			return []Ejemplo{}, nil
		}
		return nil, err
	}
	if ejemplos == nil {
		ejemplos = []Ejemplo{}
	}
	return ejemplos, nil
}

// listarMensajes lee los mensajes. Solo se usa con el mutex ya tomado.
func (a *Almacen) listarMensajes() ([]Mensaje, error) {
	var mensajes []Mensaje
	if err := leerJSON(a.ruta("mensajes.json"), &mensajes); err != nil {
		if errors.Is(err, ErrSinDatos) {
			return []Mensaje{}, nil
		}
		return nil, err
	}
	if mensajes == nil {
		mensajes = []Mensaje{}
	}
	return mensajes, nil
}

// AgregarMensaje añade un mensaje entrante a la lista guardada.
func (a *Almacen) AgregarMensaje(m Mensaje) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	mensajes, err := a.listarMensajes()
	if err != nil {
		return err
	}
	return escribirJSON(a.ruta("mensajes.json"), append(mensajes, m))
}

// ListarMensajes devuelve todos los mensajes guardados, en orden de llegada.
func (a *Almacen) ListarMensajes() ([]Mensaje, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.listarMensajes()
}

// MarcarRespondido marca un mensaje como ya respondido con un borrador.
func (a *Almacen) MarcarRespondido(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	mensajes, err := a.listarMensajes()
	if err != nil {
		return err
	}
	for i := range mensajes {
		if mensajes[i].ID == id {
			mensajes[i].Respondido = true
		}
	}
	return escribirJSON(a.ruta("mensajes.json"), mensajes)
}

// listarBorradores lee los borradores. Solo se usa con el mutex ya tomado.
func (a *Almacen) listarBorradores() ([]Borrador, error) {
	var borradores []Borrador
	if err := leerJSON(a.ruta("borradores.json"), &borradores); err != nil {
		if errors.Is(err, ErrSinDatos) {
			return []Borrador{}, nil
		}
		return nil, err
	}
	if borradores == nil {
		borradores = []Borrador{}
	}
	return borradores, nil
}

// GuardarBorrador añade un borrador de respuesta a la lista guardada.
func (a *Almacen) GuardarBorrador(b Borrador) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	borradores, err := a.listarBorradores()
	if err != nil {
		return err
	}
	return escribirJSON(a.ruta("borradores.json"), append(borradores, b))
}

// ListarBorradores devuelve todos los borradores guardados.
func (a *Almacen) ListarBorradores() ([]Borrador, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.listarBorradores()
}
