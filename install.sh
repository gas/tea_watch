#!/bin/sh
#
# Script de instalación para tea_watch
#
# Este script descarga el binario apropiado (tar.gz o zip) de la última
# release de GitHub y lo instala en /usr/local/bin.

set -e # Salir inmediatamente si un comando falla

# 1. Detectar la arquitectura y el sistema operativo
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64) ARCH="arm64" ;;
esac

# 2. Encontrar la última versión
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/gas/tea_watch/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_RELEASE" ]; then
  echo "Error: No se pudo encontrar la última release de tea_watch."
  exit 1
fi

# --- LÓGICA DE DETECCIÓN DE FORMATO ---
# 3. Determinar el nombre del fichero, la extensión y el comando de extracción
FILENAME_BASE="tea_watch_${LATEST_RELEASE}-${OS}-${ARCH}"
EXT=""
EXTRACT_CMD=""
BIN_NAME="tea_watch" # Nombre del binario dentro del archivo

if [ "$OS" = "windows" ]; then
    EXT=".zip"
    # Unzip y extrae solo el .exe a la carpeta temporal
    EXTRACT_CMD="unzip -j"
    BIN_NAME="tea_watch.exe"
else
    EXT=".tar.gz"
    # Tar y extrae solo el binario, quitando la estructura de carpetas
    EXTRACT_CMD="tar -xzf"
fi

DOWNLOAD_URL="https://github.com/gas/tea_watch/releases/download/${LATEST_RELEASE}/${FILENAME_BASE}${EXT}"

echo "Descargando tea_watch ${LATEST_RELEASE} para ${OS}/${ARCH}..."

# 4. Descargar y descomprimir
TEMP_DIR=$(mktemp -d)
# Descargamos el archivo con su nombre correcto
curl -fL -o "${TEMP_DIR}/${FILENAME_BASE}${EXT}" "$DOWNLOAD_URL"
echo "Descarga completa."

# Extraemos el contenido dentro del directorio temporal
# La lógica ahora maneja ambos formatos
if [ "$OS" = "windows" ]; then
    ${EXTRACT_CMD} "${TEMP_DIR}/${FILENAME_BASE}${EXT}" "${BIN_NAME}" -d "${TEMP_DIR}"
else
    # Para tar, extraemos solo el binario que nos interesa
    ${EXTRACT_CMD} "${TEMP_DIR}/${FILENAME_BASE}${EXT}" -C "${TEMP_DIR}" --strip-components=1 "${BIN_NAME}"
fi

# 5. Instalar el binario
echo "Instalando ${BIN_NAME} en /usr/local/bin (puede requerir contraseña)..."
sudo mv "${TEMP_DIR}/${BIN_NAME}" "/usr/local/bin/tea_watch" # Siempre se instala como 'tea_watch'
rm -r "$TEMP_DIR" # Limpiar ficheros temporales
echo "¡tea_watch instalado correctamente!"

# 6. Crear fichero de configuración por defecto si no existe
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

# 7. Intentar configurar el atajo de teclado (la parte más compleja)
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