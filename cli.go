package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type command string

const (
	compileCommand command = "compile"
	bundleCommand  command = "bundle"
)

type options struct {
	command    command
	catalog    bool
	locale     string
	outputFile string
	inputFiles []string
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `USAGE`)
	fmt.Fprintln(w, `  lxnc compile <locale> [<options>] <translation file> ...`)
	fmt.Fprintln(w, `  lxnc bundle [<options>] <catalog file> ...`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `DESCRIPTION`)
	fmt.Fprintln(w, `  lxnc converts the given input files into a single binary output file.`)
	fmt.Fprintln(w, `  The output file is either a dictionary or a catalog. A dictionary contains`)
	fmt.Fprintln(w, `  all messages of the translation files and the necessary locale data to format`)
	fmt.Fprintln(w, `  these messages. A catalog, on the other hand, contains only the messages with`)
	fmt.Fprintln(w, `  a locale id to define the language.`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `  The 'compile' command compiles translation files into a single binary output`)
	fmt.Fprintln(w, `  file of the specified locale.`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `  The 'bundle' command merges binary catalog files into a single binary output`)
	fmt.Fprintln(w, `  file. All input files must reference the same locale.`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `OPTIONS`)
	fmt.Fprintln(w, `  --catalog`)
	fmt.Fprintln(w, `      Tell the compiler that a catalog should be produces instead of a dictionary.`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `  -o <output-file>, --out=<output-file>`)
	fmt.Fprintln(w, `      Specify the output file of the generated dictionary or catalog files.`)
	fmt.Fprintln(w, `      Defaults to '<locale>.lxnc'.`)
	fmt.Fprintln(w)
}

func parseCommandLine() options {
	args := os.Args[1:]
	nextArg := func(errmsg string) string {
		if len(args) == 0 || strings.HasPrefix(args[0], "-") {
			fmt.Fprintln(os.Stderr, errmsg)
			fmt.Fprintln(os.Stderr)
			printUsage(os.Stderr)
			os.Exit(1)
		}
		arg := args[0]
		args = args[1:]
		return arg
	}

	opts := options{
		command: command(nextArg("missing command")),
	}

	if opts.command == compileCommand {
		opts.locale = nextArg("missing locale")
	}

	fset := flag.NewFlagSet("lxnc", flag.ContinueOnError)
	fset.Usage = func() {}
	fset.BoolVar(&opts.catalog, "catalog", false, "")
	fset.StringVar(&opts.outputFile, "out", "", "")
	fset.StringVar(&opts.outputFile, "o", "", "")

	switch fset.Parse(args) {
	case nil:
	case flag.ErrHelp:
		printUsage(os.Stdout)
		os.Exit(0)
	default:
		fmt.Fprintln(os.Stderr)
		printUsage(os.Stderr)
		os.Exit(1)
	}

	opts.inputFiles = fset.Args()
	return opts
}
