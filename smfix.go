package main

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/macdylan/SMFix/fix"
)

func main() {
	var (
		in  *os.File
		err error
	)
	if len(os.Args) > 1 {
		in, err = os.Open(os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}
	}

	if in == nil {
		flag_usage()
	}

	var headers [][]byte
	if headers, err = fix.ExtractHeader(in); err != nil {
		log.Fatalf("Parse params failed: %s", err)
	}

	// write headers
	tmpfile, err := os.CreateTemp("", "smfix")
	if err != nil {
		log.Fatalf("Can not create temp file: %s", err)
	}

	if _, err := tmpfile.Write(bytes.Join(headers, []byte("\n"))); err != nil {
		log.Fatalln(err)
	}

	// 4. append raw gcodes
	if _, err := io.Copy(tmpfile, in); err != nil {
		log.Fatalf("Can not write gcodes: %s", err)
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatalf("Temp file error: %s", err)
	}

	// 5. finally, move tmpfile to in
	in.Close()
	if err := os.Rename(tmpfile.Name(), in.Name()); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
