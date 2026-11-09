# Alimentar mi estilo — siguiente paso del prototipo (simulado)

Guía corta para que el prototipo aprenda cómo contestas TÚ, con tus
ejemplos reales, sin conectar tu WhatsApp real ni enviar nada solo.

## Lo que falta

El prototipo que ya está verificado solo aprendió 3 respuestas de
ejemplo (`ejemplos/chat_exportado_ejemplo.txt`). Con tan poco, los
borradores salen genéricos. Para que suenen a ti hacen falta tus
ejemplos reales: entre 10 y 20 pares de mensaje entrante → tu respuesta.

## Opción A — exportar un chat (la más rápida)

1. En WhatsApp, abre un chat donde contestes mucho.
2. Exportar chat > Sin archivos. Sale un `.txt`.
3. En la web del prototipo (`http://localhost:8080`), en
   "Alimentar tu estilo", escribe `William` como nombre y pega el
   contenido del `.txt`. O por API: `POST /alimentar` con
   `{"nombre_propio":"William","texto":"..."}`.
4. Revisa el perfil en `GET /estilo`.

Solo importan TUS respuestas; los mensajes de la otra persona solo
sirven para emparejar cada respuesta con lo que te habían escrito y
no se guardan como ejemplos sueltos.

## Opción B — pegar 10 pares aquí, en el chat con Muse

Si no quieres exportar un chat entero, pega pares así, tal como los
escribiste tú (con tus faltas, tu "ya", tus emojis o sin ellos):

```
Te escriben: hola, ¿mañana puedes?
Tú: holaa, sí, mañana en la tarde te confirmo la hora

Te escriben: ¿me pasas el precio?
Tú: ya, déjame ver y te confirmo
```

Con eso se arma tu primer perfil de estilo y se prueba en el
prototipo: se simula un mensaje entrante y se revisa el borrador
que genera, para corregir lo que no suene a ti.

## Límites que conviene confirmar antes de probar

Estos temas NUNCA deberían contestarse solos, ni siquiera en
borrador automático: dinero / precios finales, compromisos o citas,
temas familiares o personales, y emergencias. En esos casos el
borrador debe decir "déjame ver y te confirmo" y dejarte a ti la
respuesta.

## Qué NO cambia esta guía

- Sigue siendo un prototipo simulado: no se conecta a tu WhatsApp
  real, el QR es simulado y nada se envía sin tu aprobación.
- La demora humana (20–69 segundos sugeridos) y el control de
  cerrar la sesión ya están en el prototipo, pero solo para la
  sesión simulada.
