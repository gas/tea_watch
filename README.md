![logo_tea_watch](https://github.com/user-attachments/assets/ec88ee38-1b54-40a9-9a38-fa18c29b97a1)


tea_watch es una utilidad de terminal, escrita en Go con lipgloss, para monitorizar cambios en el sistema de ficheros en tiempo real. Muy útil para cualquier proceso que comience a modificar archivos (Gemini-cli, no miro para nadie...).

![Captura de pantalla de tea_watch en acción](https://github.com/user-attachments/assets/fb7c343a-42cd-420c-bd1a-ff27900b8945?raw=true)

## Características

* **Monitorización en Tiempo Real:** Usa `fsnotify` para una detección de eventos eficiente y nativa.
* **Interfaz Clara y Dinámica:** Construida con `Bubble Tea` y `Lipgloss`, la interfaz se adapta al tamaño de tu terminal.
* **Contadores de Eventos:** Visualiza cuántas veces se ha creado, escrito, renombrado o borrado un fichero.
* **Navegación Intuitiva:** Muévete por la lista de ficheros con las flechas del teclado o la rueda del ratón.
* **Filtrado en Tiempo Real:** Pulsa `/` para empezar a escribir y filtrar la lista de ficheros al instante.
* **Resaltado de Eventos:** Los ficheros con cambios recientes se resaltan sutilmente para llamar tu atención.
* **Gestión Inteligente:** Agrupa eventos "atómicos" (de guardado seguro) y oculta los ficheros borrados tras un tiempo para mantener la vista limpia.

## Instalación

Puedes descargar el binario precompilado para tu sistema operativo desde la [página de Releases](https://github.com/gas/tea_watch/releases).

1.  Descarga el archivo correspondiente a tu sistema (ej. `tea_watch-linux-amd64.tar.gz`).
2.  Descomprímelo.
3.  Haz el fichero ejecutable: `chmod +x tea_watch`
4.  (Opcional, recomendado) Mueve el fichero a un directorio en tu `$PATH` para poder ejecutarlo desde cualquier lugar: `sudo mv tea_watch /usr/local/bin/`
5.  (Opcional, recomendado) Crea un binding con un atajo de teclado CTRL+

## Uso

Simplemente ejecuta el comando en tu terminal:

```bash
# Monitorizar el directorio actual
tea_watch

# Monitorizar un directorio específico
tea_watch /ruta/a/tu/directorio
```

### Atajos de Teclado

| Tecla(s)          | Acción                               |
| ----------------- | ------------------------------------ |
| `↑` / `k`         | Mover cursor hacia arriba            |
| `↓` / `j`         | Mover cursor hacia abajo             |
| `Rueda del Ratón` | Desplazarse por la lista             |
| `/`               | Entrar/salir del modo de filtrado    |
| `Esc`             | Salir del modo de filtrado / Salir del programa |
| `q` / `Ctrl+C`    | Salir del programa                   |

