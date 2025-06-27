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