package search

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var normalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func normalize(s string) string {
	result, _, _ := transform.String(normalizer, s)
	return strings.ToLower(result)
}

func levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la, lb := len(ra), len(rb)

	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

type Match struct {
	Index int
	Score float64
}

func Rank(query string, candidates []string) []Match {
	q := normalize(query)
	if q == "" {
		return nil
	}

	qWords := strings.Fields(q)
	var matches []Match

	for i, candidate := range candidates {
		c := normalize(candidate)
		score := scoreCandidate(q, qWords, c)
		if score < math.MaxFloat64 {
			matches = append(matches, Match{Index: i, Score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score < matches[j].Score
	})

	if len(matches) > 10 {
		matches = matches[:10]
	}

	return matches
}

func scoreCandidate(q string, qWords []string, candidate string) float64 {
	// Exact substring match is best
	if strings.Contains(candidate, q) {
		pos := strings.Index(candidate, q)
		return float64(pos) * 0.1
	}

	// Check if all query words appear as substrings
	allFound := true
	subScore := 0.0
	for _, w := range qWords {
		if strings.Contains(candidate, w) {
			subScore += 0.5
		} else {
			allFound = false
		}
	}
	if allFound && len(qWords) > 1 {
		return 1.0 + subScore
	}

	// Word-level Levenshtein: score each query word against candidate words
	cWords := strings.Fields(candidate)
	totalScore := 0.0
	matched := 0

	for _, qw := range qWords {
		bestDist := math.MaxFloat64
		for _, cw := range cWords {
			// Substring of a word
			if strings.Contains(cw, qw) {
				d := float64(len(cw)-len(qw)) * 0.3
				if d < bestDist {
					bestDist = d
				}
				continue
			}

			// Prefix match
			shorter := qw
			longer := cw
			if len(shorter) > len(longer) {
				shorter, longer = longer, shorter
			}
			if strings.HasPrefix(longer, shorter) {
				d := float64(len(longer)-len(shorter)) * 0.5
				if d < bestDist {
					bestDist = d
				}
				continue
			}

			// Levenshtein
			dist := levenshtein(qw, cw)
			maxLen := max(len(qw), len(cw))
			if maxLen == 0 {
				continue
			}
			ratio := float64(dist) / float64(maxLen)
			if ratio <= 0.5 {
				d := float64(dist) * 2.0
				if d < bestDist {
					bestDist = d
				}
			}
		}

		if bestDist < math.MaxFloat64 {
			totalScore += bestDist
			matched++
		}
	}

	if matched == 0 {
		return math.MaxFloat64
	}

	// Penalize if not all query words matched
	penalty := float64(len(qWords)-matched) * 5.0
	return 5.0 + totalScore + penalty
}
