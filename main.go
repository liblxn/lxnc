package main

import (
	"bytes"
	"fmt"
	"os"

	msgpack "github.com/mprot/msgpack-go"

	"github.com/liblxn/lxnc/locale"
	"github.com/liblxn/lxnc/lxn"
)

const targetExt = ".lxnc"

type compilation struct {
	files []string
	cat   *lxn.Catalog
	bin   []byte
}

func main() {
	opts := parseCommandLine()

	var cat *lxn.Catalog
	switch opts.command {
	case compileCommand:
		cat = compile(opts.locale, opts.inputFiles)
	case bundleCommand:
		cat = bundle(opts.inputFiles)
	default:
		fatalf("unknown command %q", opts.command)
	}

	if cat == nil {
		return
	}

	lxn.ValidateMessages(cat.Messages, warner{})

	var out bytes.Buffer
	if opts.catalog {
		if err := msgpack.Encode(&out, cat); err != nil {
			fatalf("error encoding catalog: %v", err)
		}
	} else {
		loc, err := locale.New(cat.LocaleID)
		if err != nil {
			fatalf("%v", err)
		}

		dic := lxn.NewDictionary(loc, cat.Messages)
		if err := msgpack.Encode(&out, dic); err != nil {
			fatalf("error encoding dictionary: %v", err)
		}
	}

	output := opts.outputFile
	if output == "" {
		output = cat.LocaleID + ".lxnc"
	}

	err := os.WriteFile(output, out.Bytes(), 0666)
	if err != nil {
		fatalf("error writing %s: %v", output, err)
	}
}

func compile(localeID string, inputFiles []string) *lxn.Catalog {
	switch {
	case localeID == "":
		fatalf("missing locale")
	case len(inputFiles) == 0:
		return nil
	}

	loc, err := locale.New(localeID)
	if err != nil {
		fatalf("%v", err)
	}

	cat, err := lxn.CompileCatalog(loc, inputFiles...)
	if err != nil {
		fatalf("%v", err)
	}
	return cat
}

func bundle(inputFiles []string) *lxn.Catalog {
	if len(inputFiles) == 0 {
		return nil
	}

	locale := ""
	messages := make([]lxn.Message, 0, 128)
	for _, inputFile := range inputFiles {
		f, err := os.Open(inputFile)
		if err != nil {
			fatalf("%v", err)
		}

		cat := &lxn.Catalog{}
		err = msgpack.Decode(f, cat)
		f.Close()
		switch {
		case err != nil:
			fatalf("unable to decode catalog file %q", inputFile)
		case locale != "" && cat.LocaleID != locale:
			fatalf("multiple locales detected: %s and %s", locale, cat.LocaleID)
		}

		locale = cat.LocaleID
		messages = append(messages, cat.Messages...)
	}

	return &lxn.Catalog{
		LocaleID: locale,
		Messages: messages,
	}
}

type warner struct{}

func (w warner) Warn(msg string) {
	fmt.Fprintln(os.Stderr, "warning:", msg)
}

func fatalf(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}
