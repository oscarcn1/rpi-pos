package tui

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type wifiState int

const (
	wifiScanning   wifiState = iota
	wifiList
	wifiPassword
	wifiConnecting
	wifiDone
)

type wifiNetwork struct {
	ssid     string
	signal   int
	security string
}

type wifiModel struct {
	state    wifiState
	networks []wifiNetwork
	cursor   int
	input    textinput.Model
	message  string
	isError  bool
}

type wifiScanDoneMsg struct {
	networks []wifiNetwork
	err      error
}

type wifiConnectDoneMsg struct {
	err error
}

func newWifiModel() wifiModel {
	ti := textinput.New()
	ti.Placeholder = "Contraseña de la red"
	ti.CharLimit = 63
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	return wifiModel{state: wifiScanning, input: ti}
}

func (m wifiModel) scan() tea.Cmd {
	return func() tea.Msg {
		// rescan and list
		out, err := exec.Command("nmcli", "-t", "-f", "SSID,SIGNAL,SECURITY", "device", "wifi", "list", "--rescan", "yes").Output()
		if err != nil {
			return wifiScanDoneMsg{err: fmt.Errorf("error al escanear: %v", err)}
		}
		seen := map[string]bool{}
		var nets []wifiNetwork
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			// nmcli -t uses : as separator, but SSID may contain escaped colons
			parts := splitNmcliLine(line)
			if len(parts) < 3 || parts[0] == "" {
				continue
			}
			ssid := parts[0]
			if seen[ssid] {
				continue
			}
			seen[ssid] = true
			signal := 0
			fmt.Sscanf(parts[1], "%d", &signal)
			nets = append(nets, wifiNetwork{
				ssid:     ssid,
				signal:   signal,
				security: parts[2],
			})
		}
		sort.Slice(nets, func(i, j int) bool {
			return nets[i].signal > nets[j].signal
		})
		return wifiScanDoneMsg{networks: nets}
	}
}

func (m wifiModel) connect(ssid, password string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		if password == "" {
			cmd = exec.Command("nmcli", "device", "wifi", "connect", ssid)
		} else {
			cmd = exec.Command("nmcli", "device", "wifi", "connect", ssid, "password", password)
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			// nmcli prints useful error messages
			msg := strings.TrimSpace(string(out))
			if msg == "" {
				msg = err.Error()
			}
			return wifiConnectDoneMsg{err: fmt.Errorf("%s", msg)}
		}
		return wifiConnectDoneMsg{}
	}
}

func (m wifiModel) update(msg tea.Msg) (wifiModel, tea.Cmd) {
	switch msg := msg.(type) {
	case wifiScanDoneMsg:
		if msg.err != nil {
			m.state = wifiDone
			m.message = msg.err.Error()
			m.isError = true
			return m, nil
		}
		if len(msg.networks) == 0 {
			m.state = wifiDone
			m.message = "No se encontraron redes WiFi disponibles"
			m.isError = true
			return m, nil
		}
		m.networks = msg.networks
		m.state = wifiList
		m.cursor = 0
		return m, nil

	case wifiConnectDoneMsg:
		m.state = wifiDone
		if msg.err != nil {
			m.message = "Error al conectar: " + msg.err.Error()
			m.isError = true
		} else {
			selected := m.networks[m.cursor]
			m.message = fmt.Sprintf("Conectado a %s", selected.ssid)
			m.isError = false
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case wifiList:
			return m.updateList(msg)
		case wifiPassword:
			return m.updatePassword(msg)
		case wifiDone:
			return m.updateDone(msg)
		}
	}
	return m, nil
}

func (m wifiModel) updateList(msg tea.KeyMsg) (wifiModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.networks)-1 {
			m.cursor++
		}
	case "r":
		m.state = wifiScanning
		m.message = ""
		return m, m.scan()
	case "enter":
		selected := m.networks[m.cursor]
		if selected.security == "" || selected.security == "--" {
			// Open network
			m.state = wifiConnecting
			m.message = fmt.Sprintf("Conectando a %s...", selected.ssid)
			m.isError = false
			return m, m.connect(selected.ssid, "")
		}
		m.state = wifiPassword
		m.input.SetValue("")
		m.input.Focus()
		m.message = ""
		return m, nil
	}
	return m, nil
}

func (m wifiModel) updatePassword(msg tea.KeyMsg) (wifiModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = wifiList
		m.message = ""
		return m, nil
	case "enter":
		pw := strings.TrimSpace(m.input.Value())
		if pw == "" {
			m.message = "Ingresa la contraseña"
			m.isError = true
			return m, nil
		}
		selected := m.networks[m.cursor]
		m.state = wifiConnecting
		m.message = fmt.Sprintf("Conectando a %s...", selected.ssid)
		m.isError = false
		return m, m.connect(selected.ssid, pw)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m wifiModel) updateDone(msg tea.KeyMsg) (wifiModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
	case "enter":
		if m.isError {
			// Retry: go back to list or rescan
			m.state = wifiScanning
			m.message = ""
			return m, m.scan()
		}
		return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
	}
	return m, nil
}

func (m wifiModel) view() string {
	var b strings.Builder

	b.WriteString(screenTitleStyle.Render("Conectarse a WiFi"))
	b.WriteString("\n\n")

	switch m.state {
	case wifiScanning:
		b.WriteString("  Escaneando redes WiFi disponibles...\n")

	case wifiList:
		if m.message != "" {
			if m.isError {
				b.WriteString("  " + errorStyle.Render(m.message) + "\n")
			} else {
				b.WriteString("  " + successStyle.Render(m.message) + "\n")
			}
		}
		// Header
		header := fmt.Sprintf("  %-30s %8s  %-15s", "Red", "Señal", "Seguridad")
		b.WriteString(tableHeaderRow.Render(header))
		b.WriteString("\n")

		maxVisible := 18
		total := len(m.networks)
		start := m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end := start + maxVisible
		if end > total {
			end = total
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}

		var rows strings.Builder
		for i := start; i < end; i++ {
			n := m.networks[i]
			bars := signalBars(n.signal)
			sec := n.security
			if sec == "" || sec == "--" {
				sec = "Abierta"
			}
			row := fmt.Sprintf("  %-30s %s %3d%%  %-15s", truncate(n.ssid, 30), bars, n.signal, sec)
			if i == m.cursor {
				rows.WriteString(tableRowSelected.Render(padRight(row, 64)))
			} else {
				rows.WriteString(tableRowNormal.Render(row))
			}
			rows.WriteString("\n")
		}
		b.WriteString(tableBox("42").Render(rows.String()))

		if total > maxVisible {
			b.WriteString(fmt.Sprintf("\n  "+dimStyle.Render("Mostrando %d-%d de %d redes"), start+1, end, total))
		}

		b.WriteString("\n\n")
		b.WriteString("  " + hKey(hkNav, "↑↓", "navegar") + hSep() + hKey(hkOk, "enter", "conectar") + hSep() + hKey(hkAct, "r", "re-escanear") + hSep() + hKey(hkDel, "esc", "menú"))

	case wifiPassword:
		selected := m.networks[m.cursor]
		b.WriteString("  Red: " + successStyle.Render(selected.ssid) + "\n\n")
		if m.message != "" && m.isError {
			b.WriteString("  " + errorStyle.Render(m.message) + "\n")
		}
		b.WriteString("  " + m.input.View() + "\n\n")
		b.WriteString("  " + hKey(hkOk, "enter", "conectar") + hSep() + hKey(hkDel, "esc", "volver"))

	case wifiConnecting:
		b.WriteString("  " + warnStyle.Render(m.message) + "\n")

	case wifiDone:
		if m.isError {
			b.WriteString("  " + errorStyle.Render(m.message) + "\n\n")
			b.WriteString("  " + hKey(hkOk, "enter", "reintentar") + hSep() + hKey(hkDel, "esc", "menú"))
		} else {
			b.WriteString("  " + successStyle.Render(m.message) + "\n\n")
			b.WriteString("  " + hKey(hkOk, "enter", "menú") + hSep() + hKey(hkNav, "esc", "menú"))
		}
	}

	b.WriteString("\n")
	return b.String()
}

func signalBars(signal int) string {
	switch {
	case signal >= 75:
		return "▂▄▆█"
	case signal >= 50:
		return "▂▄▆ "
	case signal >= 25:
		return "▂▄  "
	default:
		return "▂   "
	}
}

// splitNmcliLine splits a nmcli -t output line by `:` respecting `\:` escapes.
func splitNmcliLine(line string) []string {
	var parts []string
	var cur strings.Builder
	escaped := false
	for _, ch := range line {
		if escaped {
			cur.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == ':' {
			parts = append(parts, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteRune(ch)
	}
	parts = append(parts, cur.String())
	return parts
}
