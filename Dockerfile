# Dockerfile: imagen del prototipo para un servidor privado de un solo dueño.
# No es para un servicio público: hay una sola sesión, una sola base
# embebida y un solo proceso. Ejecutar siempre una única instancia,
# con el volumen /app/datos persistente y PROTOTIPO_TOKEN definido.
FROM golang:1.23-alpine AS construir
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /prototipo ./cmd/servidor

FROM alpine:3.20
WORKDIR /app
COPY --from=construir /prototipo ./prototipo
VOLUME ["/app/datos"]
EXPOSE 8080
ENTRYPOINT ["./prototipo", "-direccion", "0.0.0.0", "-puerto", "8080", "-datos", "datos"]
