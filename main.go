package main

import (
	"fmt"
	"log"
	"os" 
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

// --- CONSTANTES ---
const maxViewHeight = 10 // Altura máxima de la lista antes de hacer scroll

// --- ICONOS (requiere Nerd Font) ---
const (
	appName1		= " tea"
	appName2		= "watch"
	iconApp         = " " // app icon:		      󰈈 󰙅 
	iconFolder      = " " // dir icon
	iconFile        = " " // file icon:	  
	iconCreate      = " " // new event: 	 
	iconWrite       = " " // write event: 	    󰘛 
	iconRemove      = " " // delete event:      
	iconRename      = " " // rename event:	  󰯏 󱈢 
	iconChmod       = " "  // chmod event:  	
	iconAtomic      = " "  // chmod event:     󰛕 
	iconPlaceholder = " " // Espacio para alineación
)

// --- ESTILOS con Lipgloss ---
var (
	// Estilo para el título de la aplicación
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))
			//Bold(true).

	iconStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#88EE6F")). // A bright green
		PaddingLeft(0). // Ensure no extra padding
		PaddingRight(0) // Ensure no extra padding
		//Background(lipgloss.Color("#7D56F4")).

	// Estilo para el pie de página con instrucciones
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Estilo para la fila seleccionada por el cursor
	selectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236"))

	// Estilo para ficheros eliminados (tachado y tenue)
	deletedFileStyle = lipgloss.NewStyle().
				Strikethrough(true).
				Foreground(lipgloss.Color("240"))

	// Estilo para el encabezado de las columnas
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	customBorder = lipgloss.Border{
		Bottom: "-",
	}

	// Separador bajo el encabezado
	separatorStyle = lipgloss.NewStyle().
			BorderStyle(customBorder).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("#7D56F4"))

	appName = titleStyle.Render(appName1) + iconStyle.Render(iconApp) + titleStyle.Render(appName2)

	highlightedRowStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))
)
			//BorderStyle(lipgloss.NormalBorder()).
			//BorderStyle(lipgloss.RoundedBorder()).		

// --- ESTRUCTURAS DE DATOS ---

// fileEventInfo almacena la información agregada para un fichero.
type fileEventInfo struct {
	path      string
	isDir     bool
	deleted   bool
	create    int
	write     int
	remove    int
	rename    int
	chmod     int
	lastEvent time.Time
}

// eventMsg es el mensaje que enviamos a Bubble Tea cuando fsnotify detecta un cambio.
type eventMsg fsnotify.Event

// errorMsg es para enviar errores del watcher a Bubble Tea.
type errorMsg struct{ err error }

// --- MODELO DE BUBBLE TEA ---

// model es el estado principal de nuestra aplicación TUI.
type model struct {
	watcher      *fsnotify.Watcher
	events       map[string]*fileEventInfo // Mapa para agregar eventos por ruta
	sortedPaths  []string                  // Rutas ordenadas para una visualización estable
	width        int                       // Ancho del terminal
	height       int                       // Alto del terminal
	cursor       int                       // Posición del cursor en la lista
	atomicEvents int 					   // eventos de guardado atómico
	scrollOffset int                       // Desplazamiento de la vista para el scroll
	lastError    string
	targetPath   string // Directorio que se está observando
}

// initialModel crea el modelo inicial de la aplicación.
func initialModel(path string) model {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error al crear el watcher de fsnotify:", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("No se pudo obtener la ruta absoluta para '%s': %v", path, err)
	}

	// Creamos el modelo base
	m := model{
		watcher:     watcher,
		events:      make(map[string]*fileEventInfo),
		sortedPaths: []string{},
		targetPath:  absPath,
	}

	// --- ESCANEO INICIAL Y AÑADIR A WATCHER ---
	// Usamos filepath.Walk para recorrer el directorio de forma recursiva.
	// Hacemos dos cosas a la vez:
	// 1. Añadir cada subdirectorio al watcher.
	// 2. Poblar nuestro mapa 'events' con los ficheros y directorios existentes.
	walkErr := filepath.Walk(absPath, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignorar directorios que generan mucho ruido
		name := info.Name()

		// --- AÑADE ESTA CONDICIÓN ---
		// Ignorar los ficheros temporales de guardado de GTK/GNOME
		if strings.HasPrefix(name, ".goutputstream-") {
			return nil // Ignora este fichero y sigue con el siguiente
		}

		if info.IsDir() && (name == ".git" || name == "node_modules" || name == ".idea") {
			return filepath.SkipDir // No procesar este directorio ni sus hijos
		}

		// Añadir el directorio al watcher para vigilar cambios futuros
		if info.IsDir() {
			if err := m.watcher.Add(walkPath); err != nil {
				// En lugar de fallar, podemos registrar el error y continuar
				log.Printf("Error añadiendo '%s' al watcher: %v", walkPath, err)
			}
		}

		// Añadir la entrada al mapa de eventos para la vista inicial
		m.events[walkPath] = &fileEventInfo{
			path:  walkPath,
			isDir: info.IsDir(),
		}

		return nil
	})

	if walkErr != nil {
		log.Fatalf("Error durante el escaneo inicial del directorio '%s': %v", absPath, walkErr)
	}

	delete(m.events, absPath) // ocultar directorio .

	// Una vez que el mapa de eventos está poblado, actualizamos las rutas ordenadas
	m.updateSortedPaths()

	return m
}

// watchEvents es una goroutine que escucha eventos de fsnotify.
func (m *model) watchEvents() tea.Cmd {
	return func() tea.Msg {
		for {
			select {
			case event, ok := <-m.watcher.Events:
				if !ok {
					return nil // Channel cerrado
				}
				return eventMsg(event)
			case err, ok := <-m.watcher.Errors:
				if !ok {
					return nil
				}
				return errorMsg{err}
			}
		}
	}
}

// addPathRecursively añade un directorio y sus subdirectorios al watcher.
func addPathRecursively(watcher *fsnotify.Watcher, path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Ignorar directorios que generan mucho ruido
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == ".idea" {
				return filepath.SkipDir
			}
			if err := watcher.Add(walkPath); err != nil {
				return fmt.Errorf("error añadiendo '%s' al watcher: %w", walkPath, err)
			}
		}
		return nil
	})
}

// Init se ejecuta una sola vez cuando la aplicación se inicia.
func (m model) Init() tea.Cmd {
	// El watcher ya fue configurado en initialModel.
	// Solo necesitamos empezar a escuchar eventos en una goroutine.
	return m.watchEvents()
}

// Update maneja los mensajes y actualiza el modelo.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.watcher.Close()
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Ajustar scroll si el cursor sale por arriba de la vista
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.sortedPaths)-1 {
				m.cursor++
				// Ajustar scroll si el cursor sale por abajo de la vista
				if m.cursor >= m.scrollOffset+maxViewHeight {
					m.scrollOffset = m.cursor - maxViewHeight + 1
				}
			}
		}

	case errorMsg:
		m.lastError = msg.err.Error()

	case eventMsg:

		// Usamos filepath.Base para obtener solo el nombre del fichero.
		if strings.HasPrefix(filepath.Base(msg.Name), ".goutputstream-") {
			m.atomicEvents++
			// Importante: Re-armamos la escucha y terminamos el procesamiento para este evento.
			return m, m.watchEvents()
		}

		// Normalizar la ruta del evento
		path := msg.Name

	// Si es un evento de creación para un nuevo directorio, hay que añadirlo al watcher.
		if msg.Op&fsnotify.Create == fsnotify.Create {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				// No necesitamos llamar a addPathRecursively, solo añadir la nueva carpeta.
				if err := m.watcher.Add(path); err != nil {
					m.lastError = fmt.Sprintf("No se pudo vigilar el nuevo dir: %v", err)
				}
			}
		}

		// Obtener o crear la entrada para este fichero.
		// Esto es importante para ficheros creados después del inicio.
		if _, ok := m.events[path]; !ok {
			isDir := false
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				isDir = true
			}
			m.events[path] = &fileEventInfo{path: path, isDir: isDir}
			m.updateSortedPaths() // Reordenar si hay un nuevo elemento
		}
		
		info := m.events[path]
		info.lastEvent = time.Now()

		// Manejar los contadores de eventos
		if msg.Op&fsnotify.Create == fsnotify.Create {
			info.create++
			info.deleted = false // Un fichero recreado ya no está borrado
		}
		if msg.Op&fsnotify.Write == fsnotify.Write {
			info.write++
		}
		if msg.Op&fsnotify.Chmod == fsnotify.Chmod {
			info.chmod++
		}

		// --- LÓGICA DE BORRADO Y RENOMBRADO MEJORADA ---
		// Tanto REMOVE como RENAME implican que la ruta original ya no existe.
		isRemoveOrRename := msg.Op&fsnotify.Remove == fsnotify.Remove || msg.Op&fsnotify.Rename == fsnotify.Rename
		if isRemoveOrRename {
			if msg.Op&fsnotify.Remove == fsnotify.Remove {
				info.remove++
			}
			if msg.Op&fsnotify.Rename == fsnotify.Rename {
				info.rename++
			}
			
			info.deleted = true

			// Si la entrada que se ha ido era un directorio, marcamos todo su contenido como borrado.
			if info.isDir {
				// La ruta base debe terminar con el separador de sistema para evitar falsos positivos
				// (p.ej. no confundir "/home/user/app" con "/home/user/application").
				dirPathPrefix := info.path + string(os.PathSeparator)
				
				// Iteramos sobre una copia de las claves para no modificar el mapa mientras lo recorremos
				for eventPath, eventInfo := range m.events {
					// Comprobamos si la ruta de un evento empieza por la ruta del directorio borrado.
					// También nos aseguramos de no procesar el directorio en sí mismo otra vez.
					if eventPath != info.path && strings.HasPrefix(eventPath, dirPathPrefix) {
						eventInfo.deleted = true
					}
				}
			}
		}

		// **CORRECCIÓN CRÍTICA:** Volvemos a armar la escucha del watcher.
		// Esto es necesario porque el comando `watchEvents` termina después de enviar un mensaje.
		return m, m.watchEvents()
	}

	return m, nil
}

// updateSortedPaths recalcula y ordena la lista de rutas.
func (m *model) updateSortedPaths() {
	paths := make([]string, 0, len(m.events))
	for path := range m.events {
		paths = append(paths, path)
	}
	sort.Strings(paths) // Orden alfabético para estructura de árbol
	m.sortedPaths = paths
}

// View renderiza la interfaz de usuario.
func (m model) View() string {
	// Vista inicial si aún no hay eventos (aunque ahora debería estar poblada desde el inicio)
	if len(m.sortedPaths) == 0 {
		return fmt.Sprintf("%s Monitoreando %s...\n%s",
			titleStyle.Render(appName),
			m.targetPath,
			helpStyle.Render("El directorio está vacío o hubo un error. Pulsa 'q' para salir."),
		)
	}

	var b strings.Builder

	// --- 1. CABECERA PERSONALIZADA (TÍTULO A LA IZQ, ICONOS A LA DER) ---

	// Título a la izquierda
	title := titleStyle.Render(appName)

	// Cabeceras de iconos a la derecha
	iconHeaders := lipgloss.JoinHorizontal(lipgloss.Left,
		headerStyle.Width(5).Render(iconCreate),
		headerStyle.Width(5).Render(iconWrite),
		headerStyle.Width(5).Render(iconRemove),
		headerStyle.Width(5).Render(iconRename),
		headerStyle.Width(5).Render(iconChmod),
	)

	// Calculamos el espacio libre para empujar los iconos a la derecha
	spaceWidth := m.width - lipgloss.Width(title) - lipgloss.Width(iconHeaders)
	if spaceWidth < 0 {
		spaceWidth = 0
	}
	spacer := lipgloss.NewStyle().Width(spaceWidth).Render("")

	// Unimos todo para formar la cabecera
	header := lipgloss.JoinHorizontal(lipgloss.Bottom, title, spacer, iconHeaders)
	b.WriteString(header)
	b.WriteString(separatorStyle.Width(m.width).Render("")) // Separador visual
	b.WriteString("\n")

	//	b.WriteString("\n")

	// --- 2. RENDERIZAR LISTA DE FICHEROS ---
	viewHeight := maxViewHeight
	if len(m.sortedPaths) < maxViewHeight {
		viewHeight = len(m.sortedPaths)
	}
	
	// Ancho flexible para la columna de fichero, dejando espacio para los contadores
	fileColWidth := m.width - 25 // 5*5 = 25 para los contadores

	for i := 0; i < viewHeight; i++ {
		idx := m.scrollOffset + i
		if idx >= len(m.sortedPaths) {
			break
		}

		pathKey := m.sortedPaths[idx]
		event := m.events[pathKey]

		// Construir nombre de fichero con indentación y icono
		relPath, _ := filepath.Rel(m.targetPath, event.path)
		depth := strings.Count(relPath, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)
		icon := iconFile
		if event.isDir {
			icon = iconFolder
		}
		
		fileNameStr := fmt.Sprintf("%s%s %s", indent, icon, filepath.Base(relPath))
		if relPath == "." {
			fileNameStr = fmt.Sprintf("%s %s", icon, relPath)
		}

		fileStyle := lipgloss.NewStyle()
		if event.deleted {
			fileStyle = deletedFileStyle
		}
		
		// Unir columnas de contadores (ajustamos el ancho a 5 para que coincida con la cabecera)
		counts := lipgloss.JoinHorizontal(
			lipgloss.Left,
			formatCounter(event.create, 5),
			formatCounter(event.write, 5),
			formatCounter(event.remove, 5),
			formatCounter(event.rename, 5),
			formatCounter(event.chmod, 5),
		)

		rowStr := lipgloss.JoinHorizontal(
			lipgloss.Bottom,
			fileStyle.Width(fileColWidth).Render(fileNameStr),
			counts,
		)

		if m.cursor == idx {
			b.WriteString(selectedRowStyle.Width(m.width).Render(rowStr))
		} else {
			b.WriteString(rowStr)
		}
		b.WriteString("\n")
	}

	// --- 3. RENDERIZAR AYUDA Y ESTADO ---
	// Texto de ayuda a la izquierda
	helpLeft := helpStyle.Render(fmt.Sprintf("Eventos: %d | Navega con ↑/↓ | 'q' para salir", len(m.sortedPaths)))
	
	// --- LÓGICA DE PAGINACIÓN ---
	// Calculamos el rango de elementos que se están mostrando
	totalFiles := len(m.sortedPaths)
	startIndex := m.scrollOffset + 1
	// El índice final es el menor entre el final del scroll o el total de ficheros
	endIndex := m.scrollOffset + viewHeight
	if endIndex > totalFiles {
		endIndex = totalFiles
	}
    // Formateamos el string de paginación
	paginationStr := fmt.Sprintf("%d-%d/%d", startIndex, endIndex, totalFiles)

	// --- EVENTOS ATÓMICOS ---
	// Contador de eventos atómicos a la derecha (solo si es mayor que cero)
	atomicEventsStr := ""
	if m.atomicEvents > 0 {
		atomicEventsStr = helpStyle.Render(fmt.Sprintf("%s: %d", iconAtomic, m.atomicEvents))
	}else{
		atomicEventsStr = helpStyle.Render(fmt.Sprintf("%s: 0", iconAtomic))		
	}

    // Unimos los elementos de la derecha con un separador
    helpRight := lipgloss.JoinHorizontal(lipgloss.Center,
        atomicEventsStr,
        helpStyle.Render(" | "), // Pequeño separador
        helpStyle.Render(paginationStr),
    )

	// Calculamos el espacio para empujar el contador a la derecha
	spaceWidth = m.width - lipgloss.Width(helpLeft) - lipgloss.Width(helpRight)
	if spaceWidth < 0 {
		spaceWidth = 0
	}
	spacer = strings.Repeat(" ", spaceWidth)

	// Unimos todo para formar el pie de página
	footer := lipgloss.JoinHorizontal(lipgloss.Bottom, helpLeft, spacer, helpRight)
	b.WriteString(footer)


	if m.lastError != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: "+m.lastError))
	}

	return b.String()
}

// formatCounter ahora usa un ancho fijo de 5 y no necesita icono
func formatCounter(n int, width int) string {
	var s string
	if n > 0 {
		s = fmt.Sprintf("%d", n)
	}
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(s)
}

func main() {
	// Usar el directorio actual si no se especifica otro
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	p := tea.NewProgram(initialModel(dir))

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error al ejecutar el programa: %v", err)
	}
}
