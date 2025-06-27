![logo_tea_watch](https://github.com/user-attachments/assets/ec88ee38-1b54-40a9-9a38-fa18c29b97a1)

![Go Build & Test](https://github.com/gas/tea_watch/actions/workflows/go.yml/badge.svg) ![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/gas/tea_watch)

**tea_watch** es una utilidad de terminal, escrita en Go con lipgloss, para monitorizar cambios en el sistema de ficheros en tiempo real. Muy útil para cualquier proceso que comience a modificar muchos archivos (Gemini-cli, no miro para nadie...).

![Captura de pantalla de tea_watch en acción](https://github.com/user-attachments/assets/fb7c343a-42cd-420c-bd1a-ff27900b8945?raw=true)

## Características

* **Monitorización en Tiempo Real:** Usa `fsnotify` para una detección de eventos eficiente y nativa.
* **Interfaz Clara y Dinámica:** Construida con `Bubble Tea` y `Lipgloss`, la interfaz se adapta al tamaño de tu terminal.
* **Contadores de Eventos:** Visualiza cuántas veces se ha creado, escrito, renombrado o borrado un fichero.
* **Navegación Intuitiva:** Muévete por la lista de ficheros con las flechas del teclado o la rueda del ratón.
* **Filtrado en Tiempo Real:** Pulsa `/` para empezar a escribir y filtrar la lista de ficheros al instante.
* **Resaltado de Eventos:** Los ficheros con cambios recientes se resaltan sutilmente para llamar tu atención.
* **Gestión Inteligente:** Agrupa eventos "atómicos" (de guardado seguro) y oculta los ficheros borrados tras un tiempo para mantener la vista limpia.

## Recomendaciones

* Una [Nerd Font](https://www.nerdfonts.com/) instalada y configurada en tu terminal para visualizar correctamente los iconos.
* Si no usas Nerd Font, tea_watch puede funcionar en modo ASCII (flag *--no-nerd-font*).

## Instalación

### Método 1: Script de Instalación (Linux y macOS)

Esta es la forma más fácil y rápida. Simplemente copia y pega esta línea en tu terminal. El script detectará tu sistema operativo, descargará la última versión, la instalará en `/usr/local/bin` y te pedirá la contraseña si es necesario.

```bash  
curl -sSL https://raw.githubusercontent.com/gas/tea_watch/main/install.sh | bash
```

El script añadirá un atajo de teclado ALT+w para ejecutarse en el directorio actual (bash y zsh)

### Método 2: Con go install (para Desarrolladores y Vibe Coders)

Si tienes el entorno de Go instalado en tu máquina, puedes instalar tea_watch con un solo comando. El binario se instalará en tu directorio $GOPATH/bin.

```Bash

go install github.com/gas/tea_watch@latest  
```


### Método 3: Instalación Manual (Todas las Plataformas)

Puedes descargar el binario precompilado para tu sistema operativo desde la [página de Releases](https://github.com/gas/tea_watch/releases).

1.  Descarga el archivo correspondiente a tu sistema (ej. `tea_watch-linux-amd64.tar.gz`).
2.  Descomprímelo.
3.  Haz el fichero ejecutable: `chmod +x tea_watch`
4.  (Opcional, recomendado) Mueve el fichero a un directorio en tu `$PATH` para poder ejecutarlo desde cualquier lugar: `sudo mv tea_watch /usr/local/bin/`
5.  (Opcional, recomendado) Crea un binding con un atajo de teclado CTRL+

## Uso

Simplemente ejecuta el comando en tu terminal:

```bash
# Monitorizar el directorio actual con iconos NF
tea_watch

# Forzar el modo sin iconos (modo ASCII)
tea_watch --no-nerd-fonts

# Monitorizar un directorio específico
tea_watch /ruta/a/tu/directorio

# Usar el flag y un directorio a la vez
tea_watch --no-nerd-fonts /ruta/a/tu/directorio
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

