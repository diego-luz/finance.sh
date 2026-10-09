package imports

import (
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

// ToUTF8 returns data as UTF-8. Brazilian banks often export OFX and CSV in
// Windows-1252 (the OFX header even says CHARSET:1252): invalid UTF-8 then
// reached the database (which refuses it) or shifted the OFX parser's offsets.
// Anything that is not valid UTF-8 is taken as Windows-1252, the superset of
// Latin-1 those exports use.
func ToUTF8(data []byte) []byte {
	if utf8.Valid(data) {
		return data
	}
	out, err := charmap.Windows1252.NewDecoder().Bytes(data)
	if err != nil {
		return data
	}
	return out
}

// asciiLower lower-cases ASCII letters only, so every byte keeps its offset:
// strings.ToLower may change the length (invalid bytes become the 3-byte
// U+FFFD, some letters grow), and the OFX parser slices the original text with
// offsets found in the lowered copy.
func asciiLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'A' <= c && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}
