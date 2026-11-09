// Package alimentador aprende el estilo del usuario desde un chat exportado.
//
// El usuario exporta un chat desde WhatsApp (abrir chat > Exportar chat >
// Sin archivos) y pega ese texto en POST /alimentar. De ahí se sacan solo
// SUS mensajes — los de la otra persona solo sirven para emparejar cada
// respuesta con lo que le habían escrito — y se construye su perfil de
// estilo: largo habitual, muletillas, saludos y palabras frecuentes.
// Para entrenar el estilo solo importan las respuestas propias, así que
// los mensajes ajenos no se guardan como ejemplos sueltos.
package alimentador

import (
	"sort"
	"strings"

	"prototipo-whatsapp/internal/memoria"
)

// Resultado es lo que se aprende de un chat exportado.
type Resultado struct {
	Perfil   memoria.PerfilEstilo
	Ejemplos []memoria.Ejemplo
}

// Analizar procesa el texto exportado y devuelve el perfil y los ejemplos.
func Analizar(nombrePropio, texto string) Resultado {
	propios, ejemplos := extraer(nombrePropio, texto)
	return Resultado{Perfil: perfilDe(nombrePropio, propios), Ejemplos: ejemplos}
}

// extraer recorre las líneas y separa los mensajes propios, emparejando
// cada respuesta propia con el último mensaje entrante que la precedió.
func extraer(nombrePropio, texto string) ([]string, []memoria.Ejemplo) {
	var propios []string
	var ejemplos []memoria.Ejemplo
	ultimoEntrante := ""
	for _, linea := range strings.Split(texto, "\n") {
		nombre, mensaje, ok := parsearLinea(linea)
		if !ok {
			continue
		}
		if esPropio(nombre, nombrePropio) {
			propios = append(propios, mensaje)
			if ultimoEntrante != "" {
				ejemplos = append(ejemplos, memoria.Ejemplo{Entrante: ultimoEntrante, Respuesta: mensaje})
			}
			continue
		}
		ultimoEntrante = mensaje
	}
	return propios, ejemplos
}

// parsearLinea lee una línea con formato de exportación de WhatsApp:
// "12/11/2026, 10:15 - Nombre: texto". Devuelve false si no es un mensaje.
func parsearLinea(linea string) (nombre, mensaje string, ok bool) {
	linea = strings.TrimSpace(linea)
	partes := strings.SplitN(linea, " - ", 2)
	if len(partes) != 2 {
		return "", "", false
	}
	nombreYTexto := strings.SplitN(partes[1], ": ", 2)
	if len(nombreYTexto) != 2 {
		return "", "", false
	}
	nombre = strings.TrimSpace(nombreYTexto[0])
	mensaje = strings.TrimSpace(nombreYTexto[1])
	if nombre == "" || mensaje == "" || strings.HasPrefix(mensaje, "<") {
		return "", "", false
	}
	return nombre, mensaje, true
}

// esPropio compara el nombre del mensaje con el nombre del usuario.
func esPropio(nombre, nombrePropio string) bool {
	return strings.EqualFold(strings.TrimSpace(nombre), strings.TrimSpace(nombrePropio))
}

// perfilDe resume los mensajes propios en un perfil de estilo.
func perfilDe(nombrePropio string, propios []string) memoria.PerfilEstilo {
	perfil := memoria.PerfilEstilo{NombrePropio: nombrePropio, MensajesAnalizados: len(propios)}
	if len(propios) == 0 {
		return perfil
	}
	totalRunas, totalEmojis := 0, 0
	saludos := map[string]bool{}
	for _, mensaje := range propios {
		runas := []rune(mensaje)
		totalRunas += len(runas)
		totalEmojis += contarEmojis(mensaje)
		bajo := strings.ToLower(mensaje)
		perfil.UsaYa = perfil.UsaYa || strings.HasPrefix(bajo, "ya") || strings.Contains(bajo, " ya")
		perfil.UsaOye = perfil.UsaOye || strings.HasPrefix(bajo, "oye") || strings.Contains(bajo, " oye")
		if strings.HasPrefix(bajo, "hola") {
			saludos["hola"] = true
		}
		if strings.HasPrefix(bajo, "qué tal") || strings.HasPrefix(bajo, "que tal") {
			saludos["qué tal"] = true
		}
	}
	perfil.LongitudMedia = float64(totalRunas) / float64(len(propios))
	if totalRunas > 0 {
		perfil.FrecuenciaEmoji = float64(totalEmojis) / float64(totalRunas)
	}
	for saludo := range saludos {
		perfil.Saludos = append(perfil.Saludos, saludo)
	}
	sort.Strings(perfil.Saludos)
	perfil.PalabrasFrecuentes = palabrasFrecuentes(propios)
	return perfil
}

// contarEmojis cuenta los caracteres fuera del texto común (emojis y símbolos).
func contarEmojis(texto string) int {
	cantidad := 0
	for _, r := range texto {
		if r >= 0x1F300 || (r >= 0x2600 && r <= 0x27BF) {
			cantidad++
		}
	}
	return cantidad
}

// palabrasFrecuentes devuelve las 5 palabras de más de 3 letras que más se repiten.
func palabrasFrecuentes(mensajes []string) []string {
	conteo := map[string]int{}
	for _, mensaje := range mensajes {
		for _, palabra := range strings.Fields(strings.ToLower(mensaje)) {
			palabra = strings.Trim(palabra, ".,!?¿¡\"'")
			if len([]rune(palabra)) > 3 {
				conteo[palabra]++
			}
		}
	}
	ordenadas := make([]string, 0, len(conteo))
	for palabra := range conteo {
		ordenadas = append(ordenadas, palabra)
	}
	sort.Slice(ordenadas, func(i, j int) bool {
		if conteo[ordenadas[i]] == conteo[ordenadas[j]] {
			return ordenadas[i] < ordenadas[j]
		}
		return conteo[ordenadas[i]] > conteo[ordenadas[j]]
	})
	if len(ordenadas) > 5 {
		ordenadas = ordenadas[:5]
	}
	return ordenadas
}
