#!/bin/sh
#
# Script de instalación para tea_watch
#
# Este script descarga el binario apropiado de la última release de GitHub
# y lo instala en /usr/local/bin. También intenta añadir un atajo de teclado.

set -e # Salir inmediatamente si un comando falla

# 1. Detectar la arquitectura y el sistema operativo
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64) ARCH="arm64" ;;
esac

# 2. Encontrar la última versión y construir la URL de descarga
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/gas/tea_watch/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_RELEASE" ]; then
  echo "Error: No se pudo encontrar la última release de tea_watch."
  exit 1
fi
DOWNLOAD_URL="https://github.com/gas/tea_watch/releases/download/${LATEST_RELEASE}/tea_watch-${OS}-${ARCH}.tar.gz"

echo "Descargando tea_watch ${LATEST_RELEASE} para ${OS}/${ARCH}..."

# 3. Descargar y descomprimir
TEMP_DIR=$(mktemp -d)
curl -L -o "${TEMP_DIR}/tea_watch.tar.gz" "$DOWNLOAD_URL"
tar -xzf "${TEMP_DIR}/tea_watch.tar.gz" -C "${TEMP_DIR}"
echo "Descarga completa."

# 4. Instalar el binario
echo "Instalando tea_watch en /usr/local/bin (puede requerir contraseña)..."
sudo mv "${TEMP_DIR}/tea_watch" /usr/local/bin/tea_watch
rm -r "$TEMP_DIR" # Limpiar ficheros temporales
echo "¡tea_watch instalado correctamente!"

# 5. Intentar configurar el atajo de teclado (la parte más compleja)
SHELL_CONFIG=""
CURRENT_SHELL=$(basename "$SHELL")

if [ "$CURRENT_SHELL" = "bash" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
elif [ "$CURRENT_SHELL" = "zsh" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
fi

if [ -n "$SHELL_CONFIG" ]; then
    echo "Añadiendo el atajo 'Alt+w' a tu fichero ${SHELL_CONFIG}..."
    # Añadimos una entrada para que no se duplique si se ejecuta de nuevo
    if ! grep -q 'bind -x '"'"'"\ew": "tea_watch"'"'" /etc/inputrc "$SHELL_CONFIG"; then
        echo '' >> "$SHELL_CONFIG"
        echo '# Atajo para tea_watch (añadido por el script de instalación)' >> "$SHELL_CONFIG"
        echo 'bind -x '"'"'"\ew": "tea_watch"'"' >> "$SHELL_CONFIG"
        echo "¡Atajo añadido! Por favor, reinicia tu terminal o ejecuta 'source ${SHELL_CONFIG}'"
    else
        echo "El atajo ya parece estar configurado."
    fi
else
    echo "No se pudo detectar tu shell (bash/zsh) para configurar el atajo automáticamente."
fi