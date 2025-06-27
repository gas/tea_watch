#!/bin/sh
#
# Script de instalación para tea_watch v1.3
#
# Instala tea_watch en el directorio local del usuario (~/.local/bin),
# sin necesidad de contraseña.

set -e # Salir inmediatamente si un comando falla

# 1. Detectar la arquitectura y el sistema operativo
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64) ARCH="arm64" ;;
esac

# 2. Encontrar la última versión (esto no cambia)
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/gas/tea_watch/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_RELEASE" ]; then
  echo "Error: No se pudo encontrar la última release de tea_watch."
  exit 1
fi

# 3. Determinar el nombre del fichero y del binario interno
#    (Como ya no usas versión en el nombre, lo dejamos simple)
FILENAME_BASE="tea_watch-${OS}-${ARCH}"
EXT=".tar.gz"
# ---> ESTA ES LA CORRECCIÓN CLAVE <---
# El nombre del binario DENTRO del .tar.gz es el mismo que el del propio fichero (sin la extensión)
BIN_NAME="${FILENAME_BASE}"
DOWNLOAD_URL="https://github.com/gas/tea_watch/releases/download/${LATEST_RELEASE}/${FILENAME_BASE}${EXT}"

echo "Descargando tea_watch ${LATEST_RELEASE} para ${OS}/${ARCH}..."

# 4. Descargar y descomprimir
TEMP_DIR=$(mktemp -d)
curl -fL -o "${TEMP_DIR}/archive.tar.gz" "$DOWNLOAD_URL"
echo "Descarga completa."
# Extraemos todo el contenido a la carpeta temporal.
tar -xzf "${TEMP_DIR}/archive.tar.gz" -C "${TEMP_DIR}"

# 5. Instalar en el directorio local del usuario (SIN sudo)
# ---> ESTE ES EL SEGUNDO CAMBIO IMPORTANTE <---
INSTALL_DIR="$HOME/.local/bin"

# Crear el directorio de instalación si no existe
if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creando directorio de instalación en ${INSTALL_DIR}..."
    mkdir -p "$INSTALL_DIR"
fi

echo "Instalando tea_watch en ${INSTALL_DIR}..."
# Movemos el binario (que tiene el nombre completo) y lo renombramos a "tea_watch"
mv "${TEMP_DIR}/${BIN_NAME}" "${INSTALL_DIR}/tea_watch"
rm -r "$TEMP_DIR" # Limpiar ficheros temporales
echo "¡tea_watch instalado correctamente!"

# 6. Comprobar si la ruta de instalación está en el $PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    # El PATH ya está configurado, todo bien.
    ;;
  *)
    # El PATH no está configurado, avisamos al usuario.
    echo "----------------------------------------------------------------"
    echo "ADVERTENCIA: Tu directorio ${INSTALL_DIR} no está en tu \$PATH."
    echo "Para poder ejecutar 'tea_watch' desde cualquier lugar, añade la siguiente"
    echo "línea al final de tu fichero ~/.bashrc o ~/.zshrc:"
    echo ""
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "Luego, reinicia tu terminal."
    echo "----------------------------------------------------------------"
    ;;
esac

# 7. Crear fichero de configuración por defecto si no existe
CONFIG_DIR="$HOME/.config/tea_watch"
if [ ! -d "$CONFIG_DIR" ]; then
    echo "Creando directorio de configuración en ${CONFIG_DIR}..."
    mkdir -p "$CONFIG_DIR"
fi

CONFIG_FILE="$CONFIG_DIR/config.toml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creando fichero de configuración de ejemplo en ${CONFIG_FILE}..."
    cat > "$CONFIG_FILE" << EOL
# Configuration file for tea_watch
# -------------------------------------
nerd_fonts = true

# Uncomment (and translate if necessary) 
# the following [strings] to make use of them.

[strings]
# monitoring    = "Monitoring"
# empty_dir      = "Directory is empty or an error occurred. Press 'q' to quit."
# filter_prompt  = "Filter: "
# help_nav       = "Navigate with ↑/↓"
# help_filter    = "/ filter"
# help_quit      = "'q' quit"
# atomic_events  = "Atomic Events"
# total_events   = "Events"
EOL
fi

# 8. Intentar configurar el atajo de teclado (la parte más compleja)
SHELL_CONFIG=""
CURRENT_SHELL=$(basename "$SHELL")

if [ "$CURRENT_SHELL" = "bash" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
elif [ "$CURRENT_SHELL" = "zsh" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
fi

if [ -n "$SHELL_CONFIG" ]; then
    # Lógica para BASH
    if [ "$CURRENT_SHELL" = "bash" ]; then
        if ! grep -q 'bind -x '"'"'"\ew": "tea_watch"'"' "$SHELL_CONFIG"; then
            echo '' >> "$SHELL_CONFIG"
            echo '# Atajo para tea_watch (añadido por el script de instalación)' >> "$SHELL_CONFIG"
            echo 'bind -x '"'"'"\ew": "tea_watch"'"' >> "$SHELL_CONFIG"
            echo "¡Atajo para bash añadido!"
        else
            echo "El atajo de bash ya parece estar configurado."
        fi
    fi

    # Lógica para ZSH
    if [ "$CURRENT_SHELL" = "zsh" ]; then
        if ! grep -q "tea_watch_widget" "$SHELL_CONFIG"; then
            echo '' >> "$SHELL_CONFIG"
            echo '# Atajo para tea_watch (añadido por el script de instalación)' >> "$SHELL_CONFIG"
            echo 'tea_watch_widget() {' >> "$SHELL_CONFIG"
            echo '  tea_watch' >> "$SHELL_CONFIG"
            echo '  zle reset-prompt' >> "$SHELL_CONFIG"
            echo '}' >> "$SHELL_CONFIG"
            echo 'zle -N tea_watch_widget' >> "$SHELL_CONFIG"
            echo 'bindkey '\''\ew'\'' tea_watch_widget' >> "$SHELL_CONFIG"
            echo "¡Atajo para zsh añadido!"
        else
            echo "El atajo de zsh ya parece estar configurado."
        fi
    fi

    echo "Por favor, reinicia tu terminal o ejecuta 'source ${SHELL_CONFIG}' para aplicar los cambios."

else
    echo "No se pudo detectar tu shell (bash/zsh) para configurar el atajo automáticamente."
fi