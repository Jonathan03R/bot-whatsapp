# Prototipo WhatsApp — web y API local (simulado)

Prototipo en Go que presenta tu WhatsApp en una web y una API: sesión con
QR, endpoints para leer mensajes y responder, y un flujo aparte para
aprender cómo hablas y contestar desde memoria. Todo en local.

## Qué es y qué NO es

**Sí es:**
- Una web local (`http://localhost:8080`) con tus mensajes simulados y
  borradores de respuesta para aprobar.
- Un flujo de mensajes **interno y automático**: al entrar un mensaje,
  el servidor genera su borrador solo, sin que pidas la respuesta con
  otra llamada. `POST /mensajes` solo hace de *listener* de prueba en
  la simulación, y `GET /mensajes` queda como consulta de administración.
- Un alimentador (`POST /alimentar`) que aprende tu estilo desde un chat
  exportado de WhatsApp y lo guarda en la base embebida local.
- Una sesión que puedes cerrar en cualquier momento
  (`POST /sesion/cerrar`), como cerrar un dispositivo vinculado desde el
  celular.
- Un diseño de **servidor privado y de un solo dueño**: una instancia,
  una sesión y una base embebida. Con `PROTOTIPO_TOKEN` definido, los
  endpoints exigen la cabecera `X-Token` (la web lo pide al primer 401).

**NO es:**
- No se conecta a WhatsApp ni a ningún número real. El QR es simulado
  (texto `SIMULADO-...`), no un QR válido de WhatsApp.
- No envía mensajes solo. `POST /responder` genera un **borrador**; nada
  sale sin tu aprobación.
- No contesta al instante: cada borrador trae una espera sugerida
  (20–69 segundos) para contestar con ritmo humano.
- Cuando falta un dato, el borrador no inventa nada: dice
  "déjame ver y te confirmo".

Si algún día se pasa a la vía oficial (WhatsApp Business / Cloud API), la
estructura se reutiliza: solo cambia el conector de `sesion/`; el
alimentador, la memoria y las respuestas siguen igual. La vía no oficial
por QR a un número personal puede bloquear el número, por eso este
prototipo no la implementa.

## Cómo correrlo

Requisitos: Go 1.23 o superior. Sin dependencias externas: solo la
biblioteca estándar, así que no hace falta descargar nada.

```bash
cd prototipo-whatsapp
go run ./cmd/servidor
# o con otra dirección, puerto y carpeta de datos:
go run ./cmd/servidor -direccion 127.0.0.1 -puerto 8080 -datos datos
```

Abre http://localhost:8080 en el navegador.

### En un servidor privado (Docker)

Pensado para un VPS tuyo, no para un servicio público: cada sesión es
un proceso con estado propio y la base embebida admite un solo proceso,
así que se ejecuta **una sola instancia** y `datos/` debe ser un volumen
persistente. Sin volumen, al redeplegar se pierde el perfil aprendido.

```bash
docker build -t prototipo-whatsapp .
docker run -d --name prototipo -p 127.0.0.1:8080:8080 \
  -e PROTOTIPO_TOKEN=pon-aqui-un-token-largo \
  -v prototipo-datos:/app/datos prototipo-whatsapp
```

El puerto se publica solo en `127.0.0.1` del servidor: el acceso desde
fuera debe ir por tu VPN o tu túnel SSH, nunca abierto directo a internet.

Para compilar y comprobar que todo funciona:

```bash
go vet ./...
go build ./...
```

## Prueba rápida (con el chat de ejemplo)

1. En la web, pulsa **Generar QR simulado** y luego **Simular conexión**.
2. En "Alimentar tu estilo", escribe `William` como nombre y pega el
   contenido de `ejemplos/chat_exportado_ejemplo.txt`.
3. Simula un mensaje entrante, por ejemplo de `Ana`: `hola, ¿mañana puedes?`
4. Sin pulsar nada más, el flujo interno ya generó el borrador: revisa
   el borrador y la espera sugerida en "Borradores".

Lo mismo por API:

```bash
curl -X POST localhost:8080/sesion/qr
curl -X POST localhost:8080/sesion/conectar
curl -X POST localhost:8080/alimentar \
  -H 'Content-Type: application/json' \
  -d '{"nombre_propio":"William","texto":"12/11/2026, 10:15 - Ana: hola\n12/11/2026, 10:16 - William: holaa, sí"}'
curl localhost:8080/estilo
curl -X POST localhost:8080/mensajes \
  -H 'Content-Type: application/json' \
  -d '{"de":"Ana","texto":"hola, ¿mañana puedes?"}'
# la respuesta ya trae mensaje + borrador: el flujo es interno.
# /responder solo queda para regenerar un borrador de un mensaje guardado:
curl -X POST localhost:8080/responder \
  -H 'Content-Type: application/json' \
  -d '{"mensaje_id":"msg-..."}'
curl localhost:8080/borradores
curl -X POST localhost:8080/sesion/cerrar
```

## Endpoints

| Método | Ruta | Qué hace |
|---|---|---|
| GET | `/salud` | Comprueba que el servidor responde |
| GET | `/sesion` | Estado de la sesión simulada |
| POST | `/sesion/qr` | Genera el código QR simulado |
| POST | `/sesion/conectar` | Marca la sesión como conectada (simulada) |
| POST | `/sesion/cerrar` | Cierra la sesión |
| POST | `/alimentar` | Aprende tu estilo desde un chat exportado |
| GET | `/estilo` | Perfil de estilo aprendido |
| GET | `/mensajes` | Mensajes entrantes simulados |
| POST | `/mensajes` | Simula el listener: entra un mensaje y el flujo interno genera su borrador |
| POST | `/responder` | Regenera el borrador de un mensaje guardado |
| GET | `/borradores` | Borradores pendientes de aprobar |
| GET | `/` | La web del prototipo |

## Estructura (por feature, pocos archivos)

```
Dockerfile                  # imagen para un servidor privado (1 instancia)
cmd/servidor/main.go        # arranque: flags, almacén, sesión, servidor
internal/memoria/           # base embebida local (archivos JSON, escritura atómica)
internal/sesion/            # sesión simulada: QR, conectar, cerrar
internal/alimentador/       # aprende tu estilo desde un chat exportado
internal/mensajes/          # mensajes entrantes simulados
internal/respuestas/        # borradores en tu estilo; Generador enchufable a una IA
internal/servidor/          # web + API (rutas y manejadores)
ejemplos/                   # chat exportado de ejemplo para probar
datos/                      # se crea al arrancar; ahí viven perfil, ejemplos,
                            # mensajes y borradores (no se sube a git)
```

Cada archivo lleva comentarios de documentación en español. La
complejidad ciclomática se mantiene baja con funciones pequeñas, retornos
tempranos y ayudantes reutilizados (`escribirJSON`, `leerJSON`,
`contieneAlguna`, `esperaSugerida`).

## Cómo enchufar una IA de verdad (cuando quieras)

`respuestas.Generador` es una interfaz: implementa `Generar(perfil,
entrante)` llamando a la API de tu modelo favorito con el perfil de
estilo + los ejemplos + el historial, y pásala a
`respuestas.Nuevo(almacen, tuGenerador)` en `servidor.Nuevo`. La
inteligencia vive en esa llamada, no embebida en el binario.

## Privacidad

- Código local solamente: no se sube a git (ver `.gitignore`).
- Los chats exportados que pegues quedan solo en tu carpeta `datos/`.
- Para aprender el estilo solo importan tus respuestas; los mensajes de
  la otra persona no se guardan como ejemplos sueltos.
