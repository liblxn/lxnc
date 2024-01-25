package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	msgpack "github.com/mprot/msgpack-go"

	"github.com/liblxn/lxnc/internal/locale"
	"github.com/liblxn/lxnc/lxn"
)

const targetExt = ".lxnc"

type compilation struct {
	files []string
	cat   lxn.Catalog
	bin   []byte
}

func main() {
	conf := parseCommandLine()

	loc, err := locale.New(conf.localeTag)
	if err != nil {
		fatalf("%v", err)
	}

	// compile input files
	compilations := make(map[string]compilation) // target => compilation
	if conf.merge {
		c, err := compile(loc, conf.inputFiles)
		if err != nil {
			fatalf("%v", err)
		}

		target := conf.mergeTarget
		if target == "" {
			target = loc.String() + targetExt
		}
		compilations[target] = c
	} else {
		for i := 0; i < len(conf.inputFiles); i++ {
			c, err := compile(loc, conf.inputFiles[i:i+1])
			if err != nil {
				fatalf("%v", err)
			}

			target := conf.inputFiles[i]
			if ext := filepath.Ext(target); ext != "" {
				target = strings.TrimSuffix(target, ext) + targetExt
			}
			compilations[target] = c
		}
	}

	// write binaries
	for target, c := range compilations {
		f, err := os.Create(filepath.Join(conf.outputPath, target))
		if err != nil {
			fatalf("%v", err)
		}
		_, err = f.Write(c.bin)
		f.Close()
		if err != nil {
			fatalf("error writing %s: %v", target, err)
		}
	}
}

func compile(loc locale.Locale, filenames []string) (c compilation, err error) {
	c.files = filenames
	c.cat, err = lxn.CompileFiles(loc, filenames...)
	if err != nil {
		return c, err
	}

	var buf bytes.Buffer
	if err = msgpack.Encode(&buf, &c.cat); err != nil {
		return c, err
	}
	c.bin = buf.Bytes()

	lxn.Validate(c.cat, lxn.Validator{
		Warn: func(msg string) {
			fmt.Fprintln(os.Stderr, "warning:", msg)
		},
	})
	return c, nil
}

func fatalf(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}
