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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// --- CONSTANTES ---
const maxViewHeight = 10 // Altura máxima de la lista antes de hacer scroll
const uiTotalHeight = 12 // Altura TOTAL de la UI (10 para la lista + 3 para cabecera/pie)
const deleteTimeout = 30 * time.Second 
const scrollAmount = 1 // Número de líneas a mover con cada tick de la rueda

const (
	appName1		= " tea"
	appName2		= "watch"
	iconPlaceholder = " " // Espacio para alineación
)

// --- ICONOS (requiere Nerd Font) ---
const (
	iconAppNF         = " " // app icon:		      󰈈 󰙅 
	iconFolderNF      = " " // dir icon
	iconFileNF        = " " // file icon:	  
	iconCreateNF      = " " // new event: 	 
	iconWriteNF       = " " // write event: 	    󰘛 
	iconRemoveNF      = " " // delete event:      
	iconRenameNF      = " " // rename event:	  󰯏 󱈢 
	iconChmodNF       = " "  // chmod event:  	
	iconAtomicNF      = " "  // chmod event:     󰛕 
	iconTotalEventsNF = " "
)
// --- ICONOS (no NF) ---
const (
	// ASCII Fallback Icons
	iconAppASCII         = "_"
	iconFolderASCII      = " "
	iconFileASCII        = " "
	iconCreateASCII      = "new"
	iconWriteASCII       = "wri"
	iconRemoveASCII      = "del"
	iconRenameASCII      = "mov" // M for Move
	iconChmodASCII       = "chm" // P for Permissions
	iconAtomicASCII      = "Atom"
	iconTotalEventsASCII = "Tot"
)
var (
	iconApp         string
	iconFolder      string
	iconFile        string
	iconCreate      string
	iconWrite       string
	iconRemove      string
	iconRename      string
	iconChmod       string
	iconAtomic      string
	iconTotalEvents string
)

// --- TEXTOS DE LA INTERFAZ (i18n) ---

type uiStrings struct {
	Monitoring    string
	EmptyDir      string
	FilterPrompt  string
	HelpNav       string
	HelpFilter    string
	HelpQuit      string
	AtomicEvents  string
	TotalEvents   string
}

var esStrings = uiStrings{
	Monitoring:    "Monitoreando",
	EmptyDir:      "El directorio está vacío o hubo un error. Pulsa 'q' para salir.",
	FilterPrompt:  "Filtrar: ",
	HelpNav:       "Navega con ↑/↓",
	HelpFilter:    "/ filtrar",
	HelpQuit:      "'q' salir",
	AtomicEvents:  "Eventos Atómicos",
	TotalEvents:   "Eventos",
}

// not used now, they already are at conf.toml if needed
var enStrings = uiStrings{
	Monitoring:    "Monitoring",
	EmptyDir:      "Directory is empty or an error occurred. Press 'q' to quit.",
	FilterPrompt:  "Filter: ",
	HelpNav:       "Navigate with ↑/↓",
	HelpFilter:    "/ filter",
	HelpQuit:      "'q' quit",
	AtomicEvents:  "Atomic Events",
	TotalEvents:   "Events",
}

// Esta variable global contendrá los textos seleccionados
var currentStrings uiStrings

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

	// appName = titleStyle.Render(appName1) + iconStyle.Render(iconApp) + titleStyle.Render(appName2)

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
	totalEvents  int 					   // Eventos totales
	atomicEvents int 					   // eventos de guardado atómico
	scrollOffset int                       // Desplazamiento de la vista para el scroll
	lastError    string
	isFiltering  bool   // Estamos en modo filtro?
	filter       string // El texto del filtro
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
	//return m.watchEvents()
	return tea.Batch(
		m.watchEvents(),
		tea.Tick(time.Second, func(t time.Time) tea.Msg { return t }),
	)
}

// Update maneja los mensajes y actualiza el modelo.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case time.Time:
		// Con cada tick, un fichero borrado puede desaparecer, así que recalculamos y sujetamos el cursor.
		visibleCount := len(m.getVisiblePaths())
		if m.cursor >= visibleCount {
			if visibleCount > 0 {
				m.cursor = visibleCount - 1
			} else {
				m.cursor = 0
			}
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			// Mover hacia arriba
			m.cursor -= scrollAmount
			// Asegurarse de que no nos pasamos del límite superior (0)
			if m.cursor < 0 {
				m.cursor = 0
			}
			// Ajustar el scroll si es necesario
			if m.cursor < m.scrollOffset {
				m.scrollOffset = m.cursor
			}
		case tea.MouseWheelDown:
			// Mover hacia abajo
			m.cursor += scrollAmount
			// Asegurarse de que no nos pasamos del límite inferior
			maxCursor := len(m.getVisiblePaths()) - 1
			if m.cursor > maxCursor {
				m.cursor = maxCursor
			}
			// Ajustar el scroll si es necesario
			if m.cursor >= m.scrollOffset+maxViewHeight {
				m.scrollOffset = m.cursor - maxViewHeight + 1
			}
		}

	case tea.KeyMsg:
		// Si estamos en modo filtro, la entrada de teclado tiene prioridad
		if m.isFiltering {
			switch msg.String() {
			case "esc":
				m.isFiltering = false
				m.filter = ""
			case "backspace":
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
				}
			default:
				// Añadimos el caracter tecleado al filtro
				m.filter += msg.String()
			}
			visibleCount := len(m.getVisiblePaths())
			if m.cursor >= visibleCount {
				// Si hay resultados, lo ponemos en el último. Si no, en 0.
				if visibleCount > 0 {
					m.cursor = visibleCount - 1
				} else {
					m.cursor = 0
				}
			}			
			return m, nil // Salimos para no procesar otras teclas
		}

		switch msg.String() {
		case "/":
			m.isFiltering = true
			m.filter = ""
			m.cursor = 0 // Reiniciamos el cursor al empezar a filtrar
			m.scrollOffset = 0
			return m, nil
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
			if m.cursor < len(m.getVisiblePaths())-1 {
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
		// --- PASO 1: Ignorar eventos que no nos interesan ---
		if strings.HasPrefix(filepath.Base(msg.Name), ".goutputstream-") {
			m.atomicEvents++
			return m, m.watchEvents()
		}

		m.totalEvents++
		path := msg.Name

		// --- PASO 2: Asegurarnos de que el fichero/directorio existe en nuestro mapa ---
		// Si es un evento para una ruta que no conocemos (un fichero totalmente nuevo),
		// la creamos en nuestro mapa de eventos ahora.
		if _, ok := m.events[path]; !ok {
			isDir := false
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				isDir = true
			}
			m.events[path] = &fileEventInfo{path: path, isDir: isDir}
			m.updateSortedPaths() // Reordenar la lista porque hay un nuevo elemento
		}
		info := m.events[path]

		// --- PASO 3: Lógica principal de eventos (unificada) ---

		// Si es un evento de CREACIÓN...
		if msg.Op&fsnotify.Create == fsnotify.Create {
			info.create++
			info.deleted = false
			info.lastEvent = time.Now()

			// Y si lo que se ha creado es un DIRECTORIO...
			if info.isDir {
				// 1. Lo añadimos al watcher para vigilar su interior.
				m.watcher.Add(path) // Ignoramos errores aquí por simplicidad

				// 2. Escaneamos su contenido. ESTO SOLUCIONA EL PROBLEMA DE DIRECTORIOS RENOMBRADOS.
				// Al "descubrir" su contenido, los ficheros que se "movieron" con él aparecerán en la lista.
				filepath.Walk(path, func(walkPath string, fileInfo os.FileInfo, err error) error {
					if err != nil || walkPath == path { // No procesar el directorio padre de nuevo
						return nil
					}
					if _, exists := m.events[walkPath]; !exists {
						m.events[walkPath] = &fileEventInfo{path: walkPath, isDir: fileInfo.IsDir(), lastEvent: time.Now()}
					}
					return nil
				})
				m.updateSortedPaths() // Reordenar de nuevo por si hemos añadido más cosas
			}
		}

		// Si es un evento de ESCRITURA...
		if msg.Op&fsnotify.Write == fsnotify.Write {
			info.write++
			info.lastEvent = time.Now()
		}

		// Si es un evento de PERMISOS...
		if msg.Op&fsnotify.Chmod == fsnotify.Chmod {
			info.chmod++
			info.lastEvent = time.Now()
		}

		// Si es un evento de BORRADO o RENOMBRADO...
		isRemoveOrRename := msg.Op&fsnotify.Remove == fsnotify.Remove || msg.Op&fsnotify.Rename == fsnotify.Rename
		if isRemoveOrRename {
			if msg.Op&fsnotify.Remove == fsnotify.Remove {
				info.remove++
			}
			if msg.Op&fsnotify.Rename == fsnotify.Rename {
				info.rename++
			}
			info.deleted = true
			info.lastEvent = time.Now()

			// Y si lo que se ha ido era un DIRECTORIO, marcamos sus hijos como borrados.
			if info.isDir {
				dirPathPrefix := info.path + string(os.PathSeparator)
				for eventPath, eventInfo := range m.events {
					if eventPath != info.path && strings.HasPrefix(eventPath, dirPathPrefix) {
						eventInfo.deleted = true
						eventInfo.lastEvent = time.Now() // Actualizamos también para que se oculten con el padre
					}
				}
			}
		}

		// --- PASO 4: Devolvemos el control a Bubble Tea ---
		// Al final de todo el procesamiento, re-armamos la escucha.
		return m, m.watchEvents()

	}

	return m, nil
}

// getVisiblePaths devuelve la lista de rutas que deben ser visibles en la UI.
// Contiene la lógica de filtrado y de ocultar borrados antiguos.
func (m *model) getVisiblePaths() []string {
	var visiblePaths []string

	// Lógica de filtrado
	if m.isFiltering {
		for _, path := range m.sortedPaths {
			if strings.Contains(strings.ToLower(filepath.Base(path)), strings.ToLower(m.filter)) {
				visiblePaths = append(visiblePaths, path)
			}
		}
	} else {
		// Lógica para ocultar borrados después de un tiempo
		for _, path := range m.sortedPaths {
			event := m.events[path]
			if event.deleted && time.Since(event.lastEvent) > deleteTimeout {
				continue // No añadir a la lista visible
			}
			visiblePaths = append(visiblePaths, path)
		}
	}
	return visiblePaths
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
	var b strings.Builder

	// --- 1. CABECERA PERSONALIZADA (TÍTULO A LA IZQ, ICONOS A LA DER) ---
	title := titleStyle.Render(appName1) + iconStyle.Render(iconApp) + titleStyle.Render(appName2)

	// Vista inicial si aún no hay eventos (aunque ahora debería estar poblada desde el inicio)
	if len(m.getVisiblePaths()) == 0 {
		return fmt.Sprintf("%s %s %s...\n%s",
			title,
			currentStrings.Monitoring,
			m.targetPath,
			helpStyle.Render(currentStrings.EmptyDir),
		)
	}



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

	//	b.WriteString("\n") // ya no es necesario

	// --- 2. RENDERIZAR LISTA DE FICHEROS ---
	visiblePaths := m.getVisiblePaths()
	viewHeight := maxViewHeight
	if len(visiblePaths) < maxViewHeight {
		viewHeight = len(visiblePaths)
	}
	
	fileColWidth := m.width - 25

	for i := 0; i < viewHeight; i++ {
		idx := m.scrollOffset + i
		if idx >= len(visiblePaths) {
			break
		}
		pathKey := visiblePaths[idx]
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
		
		// evitamos saltos de línea
		fileColumnStyle := fileStyle.
					MaxWidth(fileColWidth). 
					Inline(true)           

		// Renderizamos la columna del fichero con el nuevo estilo de truncado
		fileColumnStr := fileColumnStyle.Render(fileNameStr)


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
			lipgloss.NewStyle().Width(fileColWidth).Render(fileColumnStr),			counts,
		)
//			fileStyle.Width(fileColWidth).Render(fileNameStr),

		if m.cursor == idx {
			//b.WriteString(selectedRowStyle.Width(m.width).Render(rowStr))
			b.WriteString(selectedRowStyle.Width(m.width).Render(rowStr))
		} else if time.Since(event.lastEvent) < 3*time.Second {
			// Si no está el cursor, pero el evento es reciente, lo resaltamos
			b.WriteString(highlightedRowStyle.Width(m.width).Render(rowStr))
		} else {
			b.WriteString(rowStr)
		}
		b.WriteString("\n")
	}

	// --- 3. RENDERIZAR AYUDA Y ESTADO ---
	var helpLeft string
	if m.isFiltering {
		// Mostramos el campo del filtro cuando está activo
		filterPrompt := currentStrings.FilterPrompt + m.filter
		// Añadimos un "cursor" parpadeante (simple)
		if time.Now().Second()%2 == 0 {
			filterPrompt += "_"
		}
		// Texto de ayuda a la izquierda
		helpLeft = helpStyle.Render(filterPrompt)
	} else {
		// Texto de ayuda a la izquierda
		helpLeft = helpStyle.Render(fmt.Sprintf("%s: %d | %s | %s | %s", iconTotalEvents, m.totalEvents, currentStrings.HelpNav, currentStrings.HelpFilter, currentStrings.HelpQuit))
	}

	
	// --- LÓGICA DE PAGINACIÓN ---
	// Calculamos el rango de elementos que se están mostrando
	totalFiles := len(visiblePaths)
	startIndex := m.scrollOffset + 1
	// El índice final es el menor entre el final del scroll o el total de ficheros
	endIndex := m.scrollOffset + viewHeight
	if endIndex > totalFiles {
		endIndex = totalFiles
	}
    // Formateamos el string de paginación
	paginationStr := fmt.Sprintf("%d-%d/%d", startIndex, endIndex, totalFiles)

	// --- EVENTOS ATÓMICOS ---
	// Contador de eventos atómicos a la derecha
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
	// --- CONFIGURACIÓN CON VIPER ---

// 1. Le decimos a Viper dónde buscar el fichero de configuración.
	viper.AddConfigPath("$HOME/.config/tea_watch") // Ruta
	viper.SetConfigName("config")                  // Nombre del fichero
	viper.SetConfigType("toml")                    // Tipo de fichero

	// 2. Definimos los valores por defecto (español).
	// Viper los usará si el fichero de config no existe o si una clave falta.
	viper.SetDefault("nerd_fonts", true)
	viper.SetDefault("strings.monitoring", esStrings.Monitoring)
	viper.SetDefault("strings.empty_dir", esStrings.EmptyDir)
	viper.SetDefault("strings.filter_prompt", esStrings.FilterPrompt)
	viper.SetDefault("strings.help_nav", esStrings.HelpNav)
	viper.SetDefault("strings.help_filter", esStrings.HelpFilter)
	viper.SetDefault("strings.help_quit", esStrings.HelpQuit)
	viper.SetDefault("strings.atomic_events", esStrings.AtomicEvents)
	viper.SetDefault("strings.total_events", esStrings.TotalEvents)

	// 3. Leemos el fichero de configuración. No da error si no lo encuentra.
	viper.ReadInConfig()

	// 4. Definimos y bindeamos los flags. Los flags siempre tendrán prioridad.
	pflag.Bool("nerd-fonts", viper.GetBool("nerd_fonts"), "Usa iconos Nerd Font.")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// --- LÓGICA DE LA APLICACIÓN ---

	// 5. Rellenamos la struct de textos directamente desde Viper.
	// Viper elige el valor correcto: flag > fichero de config > valor por defecto.
	currentStrings = uiStrings{
		Monitoring:    viper.GetString("strings.monitoring"),
		EmptyDir:      viper.GetString("strings.empty_dir"),
		FilterPrompt:  viper.GetString("strings.filter_prompt"),
		HelpNav:       viper.GetString("strings.help_nav"),
		HelpFilter:    viper.GetString("strings.help_filter"),
		HelpQuit:      viper.GetString("strings.help_quit"),
		AtomicEvents:  viper.GetString("strings.atomic_events"),
		TotalEvents:   viper.GetString("strings.total_events"),
	}

	// Selección de iconos
	if viper.GetBool("nerd-fonts") {
		// iconos NF
		iconApp = iconAppNF
		iconFolder = iconFolderNF
		iconFile = iconFileNF
		iconCreate = iconCreateNF
		iconWrite = iconWriteNF
		iconRemove = iconRemoveNF
		iconRename = iconRenameNF
		iconChmod = iconChmodNF
		iconAtomic = iconAtomicNF
		iconTotalEvents = iconTotalEventsNF
	} else {
		// iconos-caracteres ASCII
		iconApp = iconAppASCII
		iconFolder = iconFolderASCII
		iconFile = iconFileASCII
		iconCreate = iconCreateASCII
		iconWrite = iconWriteASCII
		iconRemove = iconRemoveASCII
		iconRename = iconRenameASCII
		iconChmod = iconChmodASCII
		iconAtomic = iconAtomicASCII
		iconTotalEvents = iconTotalEventsASCII
	}

	// Usar el directorio actual si no se especifica otro
	dir := "."
	if len(pflag.Args()) > 0 {
		dir = pflag.Args()[0]
	}

	//p := tea.NewProgram(initialModel(dir))
	m := initialModel(dir)
	p := tea.NewProgram(&m,
		tea.WithMouseAllMotion(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error al ejecutar el programa: %v", err)
	}

	// --- CÓDIGO DE LIMPIEZA MANUAL ---
	// Esto se ejecuta JUSTO DESPUÉS de que el programa termine (al pulsar 'q')
	// \033[%dA : Mueve el cursor N líneas hacia ARRIBA
	// \033[J   : Borra todo desde el cursor hasta el final de la pantalla
	fmt.Printf("\033[%dA\033[J", uiTotalHeight)	
}
