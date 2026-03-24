package search

import (
	"fmt"
	"testing"
)

func TestFuzzySearch(t *testing.T) {
	products := []string{
		"Sartén antiadherente 26cm",
		"Olla de presión 6L",
		"Juego de cuchillos x5",
		"Licuadora 3 velocidades",
		"Batería de cocina 9 pzas",
		"Tostador de pan 2 rebanadas",
		"Tabla para picar bambú",
		"Escoba de plástico",
		"Trapeador de microfibra",
		"Cortina para baño",
		"Foco LED 9W luz blanca",
		"Almohada estándar",
		"Cinta métrica 5m",
		"Martillo de uña",
	}

	tests := []struct {
		query     string
		wantFirst string
	}{
		{"sarten", "Sartén antiadherente 26cm"},
		{"olla", "Olla de presión 6L"},
		{"cuchillo", "Juego de cuchillos x5"},
		{"licua", "Licuadora 3 velocidades"},
		{"escoba", "Escoba de plástico"},
		{"almohada", "Almohada estándar"},
		{"martillo", "Martillo de uña"},
		{"tostador", "Tostador de pan 2 rebanadas"},
		{"foco led", "Foco LED 9W luz blanca"},
		{"cinta", "Cinta métrica 5m"},
	}

	for _, tt := range tests {
		matches := Rank(tt.query, products)
		if len(matches) == 0 {
			t.Errorf("query %q: no matches found, want %q", tt.query, tt.wantFirst)
			continue
		}
		got := products[matches[0].Index]
		if got != tt.wantFirst {
			t.Errorf("query %q: got %q, want %q", tt.query, got, tt.wantFirst)
			fmt.Printf("  All matches for %q:\n", tt.query)
			for _, m := range matches {
				fmt.Printf("    [%.2f] %s\n", m.Score, products[m.Index])
			}
		}
	}
}
