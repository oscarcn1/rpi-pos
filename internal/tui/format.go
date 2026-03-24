package tui

import (
	"fmt"
	"strings"
)

// fmtP formats a price with commas: "$1,234.56"
func fmtP(n float64) string {
	return "$" + fmtD(n)
}

// fmtD formats a float with 2 decimals and thousand separators.
func fmtD(n float64) string {
	neg := ""
	if n < 0 {
		neg = "-"
		n = -n
	}
	s := fmt.Sprintf("%.2f", n)
	parts := strings.Split(s, ".")
	return neg + addCommas(parts[0]) + "." + parts[1]
}

// fmtQ formats a quantity: whole numbers without decimals, otherwise 2 decimals.
func fmtQ(n float64) string {
	if n == float64(int64(n)) {
		return addCommas(fmt.Sprintf("%.0f", n))
	}
	s := fmt.Sprintf("%.2f", n)
	parts := strings.Split(s, ".")
	return addCommas(parts[0]) + "." + parts[1]
}

// fmtI formats an integer with commas.
func fmtI(n int) string {
	return addCommas(fmt.Sprintf("%d", n))
}

func addCommas(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// stripCommas removes commas from user input before parsing.
func stripCommas(s string) string {
	return strings.ReplaceAll(s, ",", "")
}
