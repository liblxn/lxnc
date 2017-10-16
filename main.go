package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	msgpack "github.com/tsne/msgpack-go"

	"github.com/liblxn/lxnc/locale"
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
	c.cat, err = lxn.CompileFile(loc, filenames...)
	if err != nil {
		return c, err
	}

	var buf bytes.Buffer
	if err = msgpack.Encode(&buf, &c.cat); err != nil {
		return c, err
	}
	c.bin = buf.Bytes()

	check(c)
	return c, nil
}

func check(c compilation) {
	messageKeys := make(map[string]map[string]struct{}) // section => key set
	warnedDuplicates := make(map[string]struct{})       // (section, message key) set
	for _, msg := range c.cat.Messages {
		keys, has := messageKeys[msg.Section]
		if !has {
			keys = make(map[string]struct{})
			messageKeys[msg.Section] = keys
		}

		if _, has = keys[msg.Key]; has {
			s := msg.Section + "." + msg.Key
			if _, warned := warnedDuplicates[s]; !warned {
				if msg.Section == "" {
					warnf("duplicate message key %q", msg.Key)
				} else {
					warnf("duplicate message key %q for section %q", msg.Key, msg.Section)
				}
				warnedDuplicates[s] = struct{}{}
			}
		}
		keys[msg.Key] = struct{}{}
	}
}

func fatalf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

func warnf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "warning: "+msg, args...)
	fmt.Fprintln(os.Stderr)
}
