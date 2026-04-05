package tui

import (
	"pos/internal/store"

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
)

type switchScreenMsg struct {
	screen screen
	data   any
}

type statusMsg string

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
	return a
}

func (a *App) Init() tea.Cmd {
	return tea.SetWindowTitle("POS - Punto de Venta")
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
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
	}
	return screenMargin.Render(v)
}
