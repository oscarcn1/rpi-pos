package tui

import (
	"fmt"
	"os/exec"
	"pos/internal/store"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var screenMargin = lipgloss.NewStyle().PaddingLeft(2).PaddingTop(1)

type screen int

const (
	screenMenu screen = iota
	screenSale
	screenProducts
	screenProductForm
	screenShrinkage
	screenSearch
	screenDayClose
	screenReorder
	screenInventory
	screenMonthlyFinance
	screenReturns
	screenDayReturns
	screenWifi
)

type switchScreenMsg struct {
	screen screen
	data   any
}

type statusMsg string

type tickMsg time.Time
type wifiStatusMsg string

type App struct {
	store       *store.Store
	screen      screen
	width       int
	height      int
	menu        menuModel
	sale        saleModel
	products    productsModel
	productForm productFormModel
	shrinkage   shrinkageModel
	search      searchModel
	dayClose    dayCloseModel
	reorder     reorderModel
	inventory       inventoryModel
	monthlyFinance  monthlyFinanceModel
	returns         returnsModel
	dayReturns      dayReturnsModel
	wifi            wifiModel
	currentTime     time.Time
	wifiSSID        string
}

func NewApp(s *store.Store) *App {
	a := &App{
		store: s,
	}
	a.menu = newMenuModel()
	a.sale = newSaleModel(s)
	a.products = newProductsModel(s)
	a.productForm = newProductFormModel(s)
	a.shrinkage = newShrinkageModel(s)
	a.search = newSearchModel(s)
	a.dayClose = newDayCloseModel(s)
	a.reorder = newReorderModel(s)
	a.inventory = newInventoryModel(s)
	a.monthlyFinance = newMonthlyFinanceModel(s)
	a.returns = newReturnsModel(s)
	a.dayReturns = newDayReturnsModel(s)
	a.wifi = newWifiModel()
	a.currentTime = time.Now()
	return a
}

func tickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func checkWifiCmd() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("nmcli", "-t", "-f", "NAME", "connection", "show", "--active").Output()
		if err != nil {
			return wifiStatusMsg("")
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" && l != "lo" {
				return wifiStatusMsg(l)
			}
		}
		return wifiStatusMsg("")
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen, tickCmd(), checkWifiCmd())
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tickMsg:
		a.currentTime = time.Time(msg)
		return a, tea.Batch(tickCmd(), checkWifiCmd())

	case wifiStatusMsg:
		a.wifiSSID = string(msg)
		return a, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	case switchScreenMsg:
		a.screen = msg.screen
		switch msg.screen {
		case screenMenu:
			a.menu = newMenuModel()
		case screenSale:
			a.sale = newSaleModel(a.store)
		case screenProducts:
			a.products = newProductsModel(a.store)
			return a, a.products.loadProducts()
		case screenProductForm:
			if p, ok := msg.data.(*productFormData); ok {
				a.productForm = newProductFormModelWith(a.store, p)
			} else {
				a.productForm = newProductFormModel(a.store)
			}
		case screenShrinkage:
			a.shrinkage = newShrinkageModel(a.store)
		case screenSearch:
			a.search = newSearchModel(a.store)
		case screenDayClose:
			a.dayClose = newDayCloseModel(a.store)
			return a, a.dayClose.load()
		case screenReorder:
			a.reorder = newReorderModel(a.store)
			return a, a.reorder.load()
		case screenInventory:
			a.inventory = newInventoryModel(a.store)
			return a, a.inventory.load()
		case screenMonthlyFinance:
			a.monthlyFinance = newMonthlyFinanceModel(a.store)
			return a, a.monthlyFinance.load()
		case screenReturns:
			a.returns = newReturnsModel(a.store)
			return a, a.returns.loadSales()
		case screenDayReturns:
			a.dayReturns = newDayReturnsModel(a.store)
			return a, a.dayReturns.load()
		case screenWifi:
			a.wifi = newWifiModel()
			return a, a.wifi.scan()
		}
		return a, nil
	}

	var cmd tea.Cmd
	switch a.screen {
	case screenMenu:
		a.menu, cmd = a.menu.update(msg)
	case screenSale:
		a.sale, cmd = a.sale.update(msg)
	case screenProducts:
		a.products, cmd = a.products.update(msg)
	case screenProductForm:
		a.productForm, cmd = a.productForm.update(msg)
	case screenShrinkage:
		a.shrinkage, cmd = a.shrinkage.update(msg)
	case screenSearch:
		a.search, cmd = a.search.update(msg)
	case screenDayClose:
		a.dayClose, cmd = a.dayClose.update(msg)
	case screenReorder:
		a.reorder, cmd = a.reorder.update(msg)
	case screenInventory:
		a.inventory, cmd = a.inventory.update(msg)
	case screenMonthlyFinance:
		a.monthlyFinance, cmd = a.monthlyFinance.update(msg)
	case screenReturns:
		a.returns, cmd = a.returns.update(msg)
	case screenDayReturns:
		a.dayReturns, cmd = a.dayReturns.update(msg)
	case screenWifi:
		a.wifi, cmd = a.wifi.update(msg)
	}
	return a, cmd
}

func (a *App) View() string {
	var v string
	switch a.screen {
	case screenMenu:
		v = a.menu.view()
	case screenSale:
		v = a.sale.view()
	case screenProducts:
		v = a.products.view()
	case screenProductForm:
		v = a.productForm.view()
	case screenShrinkage:
		v = a.shrinkage.view()
	case screenSearch:
		v = a.search.view()
	case screenDayClose:
		v = a.dayClose.view()
	case screenReorder:
		v = a.reorder.view()
	case screenInventory:
		v = a.inventory.view()
	case screenMonthlyFinance:
		v = a.monthlyFinance.view()
	case screenReturns:
		v = a.returns.view()
	case screenDayReturns:
		v = a.dayReturns.view()
	case screenWifi:
		v = a.wifi.view()
	}

	content := screenMargin.Render(v)

	// Calculate terminal dimensions with fallback
	h := a.height
	w := a.width
	if h == 0 {
		h = 37
	}
	if w == 0 {
		w = 100
	}

	// Build status bar
	bar := a.renderStatusBar(w)

	// Pad content to push status bar to bottom
	contentHeight := lipgloss.Height(content)
	gap := h - contentHeight - 1
	if gap < 0 {
		gap = 0
	}

	return content + strings.Repeat("\n", gap) + bar
}

func (a *App) renderStatusBar(width int) string {
	// Left: app name
	left := statusBarLeft.Render(" POS ")

	// Center: date/time
	t := a.currentTime
	if t.IsZero() {
		t = time.Now()
	}
	dayNames := []string{"Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sáb"}
	monthNames := []string{"Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dic"}
	dateStr := fmt.Sprintf(" %s %d %s %d  %02d:%02d ",
		dayNames[t.Weekday()], t.Day(), monthNames[t.Month()-1], t.Year(),
		t.Hour(), t.Minute())
	center := statusBarCenter.Render(dateStr)

	// Right: WiFi status
	var right string
	if a.wifiSSID != "" {
		right = statusBarRightOk.Render(fmt.Sprintf(" WiFi: %s ", a.wifiSSID))
	} else {
		right = statusBarRightNo.Render(" WiFi: Sin conexión ")
	}

	// Fill the gap between segments
	usedWidth := lipgloss.Width(left) + lipgloss.Width(center) + lipgloss.Width(right)
	fillWidth := width - usedWidth
	if fillWidth < 0 {
		fillWidth = 0
	}
	fill := statusBarFill.Render(strings.Repeat(" ", fillWidth))

	return left + center + fill + right
}
