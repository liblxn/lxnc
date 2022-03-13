package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/filetree"
)

func main() {
	cldrData := flag.String("cldr-data", "", "path to the directory containing the CLDR data")
	cldrVersion := flag.String("cldr-version", "<local>", "the version of the CLDR data")
	outDir := flag.String("out", "localedata", "path to the output directory")
	pkg := flag.String("pkg", "localedata", "name of the package")
	flag.Parse()

	switch {
	case *cldrData == "":
		fatal("no cldr data specified (see flag 'cldr-data')")
	case *cldrVersion == "":
		fatal("no cldr version specified (see flag 'cldr-version')")
	case *outDir == "":
		fatal("no output directory specified (see flag 'out')")
	case *pkg == "":
		fatal("no package name specified (see flag 'pkg')")
	}

	files := filetree.New(*cldrData)
	data, err := cldr.Decode(files)
	files.Close()
	if err != nil {
		fatal("decode error:", err)
	}

	gen, err := newGenerator(options{
		outputDir:   *outDir,
		packageName: *pkg,
		cldrVersion: *cldrVersion,
	})
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
