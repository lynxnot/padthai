package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lynxnot/base-padthai/pkg/padthai"
)

func main() {
	decode := flag.Bool("d", false, "decode mode: read Thai-encoded UTF-8 from stdin and write binary to stdout")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-d]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Encode binary data to Thai Unicode characters, or decode back.\n")
		fmt.Fprintf(os.Stderr, "Reads from stdin, writes to stdout.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *decode {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "padthai: read error: %v\n", err)
			os.Exit(1)
		}
		decoded, err := padthai.Decode(string(input))
		if err != nil {
			fmt.Fprintf(os.Stderr, "padthai: decode error: %v\n", err)
			os.Exit(1)
		}
		if _, err := os.Stdout.Write(decoded); err != nil {
			fmt.Fprintf(os.Stderr, "padthai: write error: %v\n", err)
			os.Exit(1)
		}
	} else {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "padthai: read error: %v\n", err)
			os.Exit(1)
		}
		encoded := padthai.Encode(input)
		if _, err := fmt.Fprint(os.Stdout, encoded); err != nil {
			fmt.Fprintf(os.Stderr, "padthai: write error: %v\n", err)
			os.Exit(1)
		}
	}
}
