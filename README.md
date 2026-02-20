# base-padthai

Like base64 encoding, but with Thai characters! 🍜

Made with love, imagination and Claude Opus 4.6 for Padthai enjoyers <3

## Encoding Scheme

**base-padthai** encodes arbitrary binary data into a UTF-8 string composed of
Thai Unicode characters, and decodes it back losslessly.

### Character Sets

| Set      | Range              | Count | Purpose                     |
|----------|--------------------|-------|-----------------------------|
| Thai     | `U+0E01`–`U+0E2F`, `U+0E3F` | 48    | Main encoding (base-48)     |
| Buginese | `U+1A00`–`U+1A0F` | 16    | Final byte padding (nibble) |

### How It Works

Input bytes are consumed **2 at a time** (big-endian) and converted to a
**base-48** triplet of Thai characters:

```
2 input bytes  →  16-bit value  →  3 Thai characters (base-48 digits, MSB first)
```

If the input has an **odd** number of bytes, the final byte is encoded as
**2 Buginese characters**, each representing one nibble (4 bits):

```
1 trailing byte  →  high nibble + low nibble  →  2 Buginese characters
```

#### Encoding Map

```
 Byte pair [B0, B1]
    │
    ▼
 val = B0<<8 | B1          (0 .. 65535)
    │
    ├── d0 = val / 48²     (0 .. 28)
    ├── d1 = val / 48 % 48 (0 .. 47)
    └── d2 = val % 48      (0 .. 47)
    │
    ▼
 ThaiAlphabet[d0], ThaiAlphabet[d1], ThaiAlphabet[d2]
```

Since `48³ = 110,592 > 65,536 = 2¹⁶`, every 16-bit value fits cleanly in
3 base-48 digits. The Buginese set covers `16² = 256` values for the
remaining single byte.

### Expansion Ratio

| Input        | Output              | Ratio (UTF-8 bytes) |
|--------------|---------------------|---------------------|
| 2 bytes      | 3 Thai chars (9 B)  | 4.5×                |
| 1 byte (pad) | 2 Buginese (6 B)    | 6.0×                |

## Installation

```sh
go install github.com/lynxnot/base-padthai/cmd/padthai@latest
```

Or build from source:

```sh
git clone https://github.com/lynxnot/base-padthai.git
cd base-padthai
go build -o padthai ./cmd/padthai/
```

## Usage

**padthai** works like `base64` — it reads from **stdin** and writes to **stdout**.

### Encode

```sh
$ echo -n "Hello, World!" | padthai
ฉฃฆญฃญญฑอคฝธญณณญฃฅᨂᨁ
```

### Decode

```sh
$ echo -n "Hello, World!" | padthai | padthai -d
Hello, World!
```

### Piping Binary Data

```sh
# Roundtrip a binary file
cat image.png | padthai > image.padthai
cat image.padthai | padthai -d > image_restored.png

# Verify integrity
md5sum image.png image_restored.png
```

### Options

```
Usage: padthai [-d]

  -d    decode mode: read Thai-encoded UTF-8 from stdin and write binary to stdout
```

## Project Structure

```
base-padthai/
├── cmd/
│   └── padthai/
│       └── main.go          # CLI entrypoint
├── pkg/
│   └── padthai/
│       ├── padthai.go        # Encoding/decoding library
│       └── padthai_test.go   # Tests and benchmarks
├── go.mod
└── README.md
```

## Library API

The `padthai` package exposes two functions:

```go
import "github.com/lynxnot/base-padthai/pkg/padthai"

// Encode binary data to a padthai string
encoded := padthai.Encode(data)

// Decode a padthai string back to bytes
decoded, err := padthai.Decode(encoded)
```

Whitespace (spaces, tabs, newlines) in the encoded string is silently skipped
during decoding, so encoded output can be safely wrapped or pretty-printed.

## Running Tests

```sh
go test ./pkg/padthai/ -v
go test ./pkg/padthai/ -bench=. -benchmem
```

## License

MIT
