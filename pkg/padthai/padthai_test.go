package padthai

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestAlphabetSize(t *testing.T) {
	if len(ThaiAlphabet) != Base {
		t.Errorf("expected Thai alphabet size %d, got %d", Base, len(ThaiAlphabet))
	}
	if len(BugineseAlphabet) != PadBase {
		t.Errorf("expected Buginese alphabet size %d, got %d", PadBase, len(BugineseAlphabet))
	}
}

func TestAlphabetUniqueness(t *testing.T) {
	seen := make(map[rune]bool)
	for _, r := range ThaiAlphabet {
		if seen[r] {
			t.Errorf("duplicate Thai character: %U", r)
		}
		seen[r] = true
	}
	for _, r := range BugineseAlphabet {
		if seen[r] {
			t.Errorf("duplicate or overlapping Buginese character: %U", r)
		}
		seen[r] = true
	}
}

func TestEncodeEmpty(t *testing.T) {
	encoded := Encode([]byte{})
	if encoded != "" {
		t.Errorf("expected empty string, got %q", encoded)
	}
}

func TestDecodeEmpty(t *testing.T) {
	decoded, err := Decode("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decoded) != 0 {
		t.Errorf("expected empty output, got %x", decoded)
	}
}

func TestRoundTripSingleByte(t *testing.T) {
	for b := 0; b < 256; b++ {
		input := []byte{byte(b)}
		encoded := Encode(input)

		// Single byte should produce 2 Buginese characters
		runes := []rune(encoded)
		if len(runes) != 2 {
			t.Errorf("byte 0x%02x: expected 2 runes, got %d: %q", b, len(runes), encoded)
			continue
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("decode byte 0x%02x: %v", b, err)
		}
		if !bytes.Equal(decoded, input) {
			t.Errorf("byte 0x%02x: roundtrip mismatch, got 0x%x", b, decoded)
		}
	}
}

func TestRoundTripTwoBytes(t *testing.T) {
	testCases := [][]byte{
		{0x00, 0x00},
		{0x00, 0x01},
		{0x01, 0x00},
		{0xFF, 0xFF},
		{0xDE, 0xAD},
		{0xBE, 0xEF},
		{0x12, 0x34},
	}

	for _, input := range testCases {
		encoded := Encode(input)

		// Two bytes should produce 3 Thai characters
		runes := []rune(encoded)
		if len(runes) != 3 {
			t.Errorf("%x: expected 3 runes, got %d: %q", input, len(runes), encoded)
			continue
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("decode %x: %v", input, err)
		}
		if !bytes.Equal(decoded, input) {
			t.Errorf("%x: roundtrip mismatch, got %x", input, decoded)
		}
	}
}

func TestRoundTripEvenLength(t *testing.T) {
	input := []byte("Hello, World! This is base-padthai encoding.")
	// Make it even length
	if len(input)%2 != 0 {
		input = append(input, '!')
	}

	encoded := Encode(input)

	// Even length input: all Thai chars, no Buginese padding
	runes := []rune(encoded)
	expectedRunes := (len(input) / 2) * 3
	if len(runes) != expectedRunes {
		t.Errorf("expected %d runes, got %d", expectedRunes, len(runes))
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Equal(decoded, input) {
		t.Errorf("roundtrip mismatch:\n  input:   %q\n  decoded: %q", input, decoded)
	}
}

func TestRoundTripOddLength(t *testing.T) {
	input := []byte("Hello, World! This is base-padthai!")
	// Make it odd length
	if len(input)%2 == 0 {
		input = input[:len(input)-1]
	}

	encoded := Encode(input)

	// Odd length: (n/2)*3 Thai chars + 2 Buginese chars
	runes := []rune(encoded)
	expectedRunes := (len(input)/2)*3 + 2
	if len(runes) != expectedRunes {
		t.Errorf("expected %d runes, got %d", expectedRunes, len(runes))
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Equal(decoded, input) {
		t.Errorf("roundtrip mismatch:\n  input:   %q\n  decoded: %q", input, decoded)
	}
}

func TestRoundTripRandomData(t *testing.T) {
	for _, size := range []int{0, 1, 2, 3, 4, 5, 10, 100, 255, 256, 1000, 1023, 1024} {
		input := make([]byte, size)
		if size > 0 {
			_, err := io.ReadFull(rand.Reader, input)
			if err != nil {
				t.Fatalf("rand read: %v", err)
			}
		}

		encoded := Encode(input)

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("decode (size %d): %v", size, err)
		}

		if !bytes.Equal(decoded, input) {
			t.Errorf("size %d: roundtrip mismatch", size)
		}
	}
}

func TestEncodedOutputIsValidUTF8(t *testing.T) {
	input := make([]byte, 137)
	_, _ = io.ReadFull(rand.Reader, input)

	encoded := Encode(input)

	for i, r := range encoded {
		if r == '\uFFFD' {
			t.Errorf("invalid UTF-8 at byte offset %d", i)
		}
	}
}

func TestEncodedCharsInAlphabet(t *testing.T) {
	thaiSet := make(map[rune]bool)
	for _, r := range ThaiAlphabet {
		thaiSet[r] = true
	}
	bugSet := make(map[rune]bool)
	for _, r := range BugineseAlphabet {
		bugSet[r] = true
	}

	input := make([]byte, 256)
	for i := range input {
		input[i] = byte(i)
	}

	encoded := Encode(input)

	for i, r := range []rune(encoded) {
		if !thaiSet[r] && !bugSet[r] {
			t.Errorf("rune at index %d (%U) is not in any alphabet", i, r)
		}
	}
}

func TestDecodeInvalidCharacter(t *testing.T) {
	_, err := Decode("ABC")
	if err == nil {
		t.Error("expected error for invalid input, got nil")
	}
}

func TestDecodeTruncatedThaiGroup(t *testing.T) {
	// Encode 2 bytes to get 3 Thai chars, then truncate to 2
	input := []byte{0x42, 0x43}
	encoded := Encode(input)

	runes := []rune(encoded)
	truncated := string(runes[:2]) // only 2 of 3 Thai chars

	_, err := Decode(truncated)
	if err == nil {
		t.Error("expected error for truncated Thai group, got nil")
	}
}

func TestDecodeSingleBuginese(t *testing.T) {
	// A single Buginese char is invalid (need exactly 0 or 2 for padding)
	single := string(BugineseAlphabet[0])
	_, err := Decode(single)
	if err == nil {
		t.Error("expected error for single Buginese character, got nil")
	}
}

func TestDecodeWhitespaceSkipped(t *testing.T) {
	input := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	encoded := Encode(input)

	// Inject whitespace between encoded characters
	runes := []rune(encoded)
	var withSpaces []rune
	for i, r := range runes {
		withSpaces = append(withSpaces, r)
		if i < len(runes)-1 {
			withSpaces = append(withSpaces, ' ')
		}
	}
	withNewlines := "\n" + string(withSpaces) + "\n"

	decoded, err := Decode(withNewlines)
	if err != nil {
		t.Fatalf("decode with whitespace: %v", err)
	}
	if !bytes.Equal(decoded, input) {
		t.Errorf("whitespace roundtrip mismatch: got %x, want %x", decoded, input)
	}
}

func TestRoundTripAllZeros(t *testing.T) {
	for _, size := range []int{1, 2, 3, 4, 100} {
		input := make([]byte, size)

		encoded := Encode(input)

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}

		if !bytes.Equal(decoded, input) {
			t.Errorf("size %d: roundtrip mismatch for all zeros", size)
		}
	}
}

func TestRoundTripAllOnes(t *testing.T) {
	for _, size := range []int{1, 2, 3, 4, 100} {
		input := make([]byte, size)
		for i := range input {
			input[i] = 0xFF
		}

		encoded := Encode(input)

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}

		if !bytes.Equal(decoded, input) {
			t.Errorf("size %d: roundtrip mismatch for all 0xFF", size)
		}
	}
}

func TestKnownEncodingStructure(t *testing.T) {
	// 2 bytes -> exactly 3 Thai chars
	twoBytes := Encode([]byte{0x00, 0x01})
	runes := []rune(twoBytes)
	if len(runes) != 3 {
		t.Fatalf("expected 3 runes for 2 bytes, got %d", len(runes))
	}
	for _, r := range runes {
		if !isThai(r) {
			t.Errorf("expected Thai character, got %U", r)
		}
	}

	// 3 bytes -> 3 Thai chars + 2 Buginese chars
	threeBytes := Encode([]byte{0x00, 0x01, 0x02})
	runes = []rune(threeBytes)
	if len(runes) != 5 {
		t.Fatalf("expected 5 runes for 3 bytes, got %d", len(runes))
	}
	for _, r := range runes[:3] {
		if !isThai(r) {
			t.Errorf("expected Thai character, got %U", r)
		}
	}
	for _, r := range runes[3:] {
		if !isBuginese(r) {
			t.Errorf("expected Buginese character, got %U", r)
		}
	}
}

func TestMaxUint16Value(t *testing.T) {
	// 0xFFFF is the max 16-bit value; make sure it roundtrips
	input := []byte{0xFF, 0xFF}
	encoded := Encode(input)
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Equal(decoded, input) {
		t.Errorf("max uint16 roundtrip mismatch: got %x, want %x", decoded, input)
	}
}

func BenchmarkEncode(b *testing.B) {
	input := make([]byte, 4096)
	_, _ = io.ReadFull(rand.Reader, input)

	b.SetBytes(int64(len(input)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Encode(input)
	}
}

func BenchmarkDecode(b *testing.B) {
	input := make([]byte, 4096)
	_, _ = io.ReadFull(rand.Reader, input)

	encoded := Encode(input)

	b.SetBytes(int64(len(input)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decode(encoded)
	}
}
