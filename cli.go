package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const binName = "lxnc"

type config struct {
	localeTag   string   // locale tag
	outputPath  string   // output path
	inputFiles  []string // .lxn files
	merge       bool
	mergeTarget string
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage:`)
	fmt.Fprintln(w, ` `, binName, `[options] [file ...]`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `Options:`)
	fmt.Fprintln(w, `  --locale <locale tag>`)
	fmt.Fprintln(w, `      Specify the locale for which the lxnc catalog should be created.`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `  --out <path>`)
	fmt.Fprintln(w, `      Specify the output path of the generated lxnc catalog files.`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `  --merge`)
	fmt.Fprintln(w, `      Merge all input files into a single catalog.`)
	fmt.Fprintln(w, `      If --merge-into is not specified, the catalog is named after the locale.`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `  --merge-into <catalog name>`)
	fmt.Fprintln(w, `      Specify the name of the merged catalog. This flag automatically sets`)
	fmt.Fprintln(w, `      the --merge option.`)
	fmt.Fprintln(w)
}

func parseCommandLine() config {
	fset := flag.NewFlagSet(binName, flag.ContinueOnError)
	fset.Usage = func() { printUsage(os.Stderr) }

	printHelp := false
	locale := fset.String("locale", "", "locale for the lxnc catalog")
	output := fset.String("out", ".", "output path for the lxnc catalogs")
	merge := fset.Bool("merge", false, "merge into single catalog")
	mergeInto := fset.String("merge-into", "", "name of the merged target catalog")
	if err := fset.Parse(os.Args[1:]); err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		printHelp = true
	}

	switch {
	case printHelp:
		os.Exit(0)
	case *locale == "":
		fatalf("no locale specified (see 'locale' flag)")
	}

	return config{
		localeTag:   *locale,
		outputPath:  *output,
		inputFiles:  fset.Args(),
		merge:       *merge || *mergeInto != "",
		mergeTarget: *mergeInto,
	}
}
