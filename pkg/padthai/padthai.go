// Package padthai implements a base-48 encoding using Thai Unicode characters.
//
// The encoding uses 48 Thai characters (U+0E01–U+0E2F and U+0E3F) to encode
// binary data. Input is processed 2 bytes at a time, converting each 16-bit
// value to base-48 and producing 3 Thai characters.
//
// If the input has an odd number of bytes, the final byte is encoded using
// 2 Buginese characters (U+1A00–U+1A0F), each representing a nibble (4 bits).
package padthai

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	// Thai character range: U+0E01 to U+0E2F (47 chars) + U+0E3F (1 char) = 48 chars
	thaiStart = '\u0e01'
	thaiEnd   = '\u0e2f'
	thaiBaht  = '\u0e3f'

	// Buginese character range: U+1A00 to U+1A0F (16 chars) for padding
	bugineseStart = '\u1a00'
	bugineseEnd   = '\u1a0f'

	// Base is the radix for the main encoding (48 Thai characters).
	Base = 48

	// PadBase is the number of Buginese characters used for padding.
	PadBase = 16
)

// ThaiAlphabet is the ordered set of 48 Thai characters used for encoding.
var ThaiAlphabet [Base]rune

// BugineseAlphabet is the ordered set of 16 Buginese characters used for padding.
var BugineseAlphabet [PadBase]rune

// thaiIndex maps a Thai rune to its index in the alphabet (0–47).
var thaiIndex map[rune]int

func init() {
	idx := 0
	for r := thaiStart; r <= thaiEnd; r++ {
		ThaiAlphabet[idx] = r
		idx++
	}
	ThaiAlphabet[idx] = thaiBaht
	// idx is now 47, total 48

	thaiIndex = make(map[rune]int, Base)
	for i, r := range ThaiAlphabet {
		thaiIndex[r] = i
	}

	for i := 0; i < PadBase; i++ {
		BugineseAlphabet[i] = bugineseStart + rune(i)
	}
}

// isThai returns true if r is one of the 48 Thai encoding characters.
func isThai(r rune) bool {
	_, ok := thaiIndex[r]
	return ok
}

// isBuginese returns true if r is in the Buginese padding range U+1A00–U+1A0F.
func isBuginese(r rune) bool {
	return r >= bugineseStart && r <= bugineseEnd
}

// Encode encodes a byte slice into a padthai string.
//
// Every 2 input bytes are treated as a big-endian 16-bit integer and converted
// to 3 base-48 digits (most-significant first), each mapped to a Thai character.
//
// A trailing single byte is encoded as 2 Buginese characters (high nibble, low nibble).
func Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var sb strings.Builder
	// Pre-allocate: each 2-byte pair -> 3 runes (up to 3 bytes each in UTF-8)
	// worst case ~4.5x expansion, plus possible 2 Buginese chars
	sb.Grow(len(data)*5 + 6)

	i := 0
	for i+1 < len(data) {
		// Take 2 bytes as a big-endian uint16
		val := uint(data[i])<<8 | uint(data[i+1])

		// Convert to 3 base-48 digits, most significant first
		d2 := val % Base
		val /= Base
		d1 := val % Base
		val /= Base
		d0 := val // val < 65536 and 48^3 = 110592, so d0 < 48

		sb.WriteRune(ThaiAlphabet[d0])
		sb.WriteRune(ThaiAlphabet[d1])
		sb.WriteRune(ThaiAlphabet[d2])

		i += 2
	}

	// Handle trailing single byte with Buginese padding
	if i < len(data) {
		b := data[i]
		hi := (b >> 4) & 0x0f
		lo := b & 0x0f
		sb.WriteRune(BugineseAlphabet[hi])
		sb.WriteRune(BugineseAlphabet[lo])
	}

	return sb.String()
}

// Decode decodes a padthai-encoded string back into the original bytes.
//
// Whitespace characters (spaces, tabs, newlines) are silently skipped.
// Returns an error if the input contains invalid characters or has an
// invalid structure.
func Decode(s string) ([]byte, error) {
	// Collect runes, skipping whitespace
	runes := make([]rune, 0, utf8.RuneCountInString(s))
	for _, r := range s {
		switch r {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			runes = append(runes, r)
		}
	}

	if len(runes) == 0 {
		return []byte{}, nil
	}

	// Determine how many trailing Buginese characters we have (0 or 2)
	trailingBuginese := 0
	if len(runes) >= 2 && isBuginese(runes[len(runes)-1]) && isBuginese(runes[len(runes)-2]) {
		trailingBuginese = 2
	}

	thaiRunes := runes[:len(runes)-trailingBuginese]
	bugRunes := runes[len(runes)-trailingBuginese:]

	if len(thaiRunes)%3 != 0 {
		return nil, fmt.Errorf("padthai: invalid encoded length: %d Thai characters is not a multiple of 3", len(thaiRunes))
	}

	// Pre-allocate output: each 3 Thai chars -> 2 bytes, plus maybe 1 byte from Buginese
	out := make([]byte, 0, (len(thaiRunes)/3)*2+trailingBuginese/2)

	// Decode Thai triplets
	for i := 0; i+2 < len(thaiRunes); i += 3 {
		d0, ok0 := thaiIndex[thaiRunes[i]]
		d1, ok1 := thaiIndex[thaiRunes[i+1]]
		d2, ok2 := thaiIndex[thaiRunes[i+2]]
		if !ok0 || !ok1 || !ok2 {
			pos := i
			if !ok0 {
				// pos is i
			} else if !ok1 {
				pos = i + 1
			} else {
				pos = i + 2
			}
			return nil, fmt.Errorf("padthai: invalid character %U at position %d", thaiRunes[pos], pos)
		}

		val := uint(d0)*Base*Base + uint(d1)*Base + uint(d2)
		if val > 0xFFFF {
			return nil, fmt.Errorf("padthai: decoded value %d exceeds 16-bit range at position %d", val, i)
		}

		out = append(out, byte(val>>8), byte(val&0xFF))
	}

	// Decode Buginese padding (single trailing byte)
	if trailingBuginese == 2 {
		hi := byte(bugRunes[0] - bugineseStart)
		lo := byte(bugRunes[1] - bugineseStart)
		if hi > 0x0f || lo > 0x0f {
			return nil, fmt.Errorf("padthai: invalid Buginese padding character")
		}
		out = append(out, (hi<<4)|lo)
	}

	return out, nil
}
