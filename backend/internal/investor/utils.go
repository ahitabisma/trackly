package investor

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func NormalizeInvestorName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))

	// Hapus aksen/diakritik (opsional, berguna untuk nama asing)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	s = result

	// Hapus tanda baca: . , - _
	rePunct := regexp.MustCompile(`[.,\-_]`)
	s = rePunct.ReplaceAllString(s, "")

	// Normalisasi "ABC PT" atau "ABC, PT" → "pt abc"
	// Pattern: nama diikuti suffix entity di akhir
	reSuffix := regexp.MustCompile(`(?i)^(.+?)\s+(pt|cv|tbk|persero|llc|ltd|inc|corp)$`)
	if m := reSuffix.FindStringSubmatch(s); len(m) == 3 {
		s = m[2] + " " + m[1]
	}

	// Collapse whitespace
	reSpace := regexp.MustCompile(`\s+`)
	s = strings.TrimSpace(reSpace.ReplaceAllString(s, " "))

	return s
}
