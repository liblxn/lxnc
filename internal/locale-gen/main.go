package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/liblxn/lxnc/internal/cldr"
)

func main() {
	cldrPath := flag.String("cldr", "", "path to the CLDR data directory")
	outDir := flag.String("out", "../../locale", "path to the output directory")
	pkg := flag.String("pkg", "locale", "name of the package")
	flag.Parse()

	switch {
	case *cldrPath == "":
		fatal("no cldr path specified (see flag 'cldr')")
	case *outDir == "":
		fatal("no output directory specified (see flag 'out')")
	case *pkg == "":
		fatal("no package name specified (see flag 'pkg')")
	}

	data, err := cldr.Decode(*cldrPath)
	if err != nil {
		fatal("decode error:", err)
	}

	gen, err := newGenerator(options{outputDir: *outDir, packageName: *pkg})
	if err != nil {
		fatal(err)
	}

	if err = gen.Generate(data); err != nil {
		fatal("generate error:", err)
	}
}

func fatal(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}
