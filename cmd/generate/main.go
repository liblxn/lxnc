package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	mprot_generator "github.com/mprot/mprotc/generator"

	"github.com/liblxn/lxnc/cmd/generate/generate_cldr"
	"github.com/liblxn/lxnc/internal/cldr"
	"github.com/liblxn/lxnc/internal/filetree"
	"github.com/liblxn/lxnc/internal/generator"
)

type options struct {
	outDir      string
	packageName string

	cldrDataDir string
	cldrVersion string

	schemaPath string
}

func main() {
	var opts options
	flag.StringVar(&opts.outDir, "out", "out", "path to the output directory")
	flag.StringVar(&opts.packageName, "pkg", "", "name of the package")
	flag.StringVar(&opts.cldrDataDir, "cldr-data", "", "path to the directory containing the CLDR data")
	flag.StringVar(&opts.cldrVersion, "cldr-version", "", "the version of the CLDR data")
	flag.StringVar(&opts.schemaPath, "schema", "", "the path to the schema definition file")
	flag.Parse()

	if opts.outDir == "" {
		fatal("no output directory specified (see flag 'out')")
	}

	if opts.packageName == "" {
		opts.packageName = filepath.Base(opts.outDir)
		if opts.packageName == "" || opts.packageName == "." {
			fatal("invalid package name (see flag 'pkg')")
		}
	}

	tasks := []struct {
		exec  func(options) error
		skip  bool
		descr string
	}{
		{
			exec:  generateCldr,
			skip:  opts.cldrDataDir == "",
			descr: "generating cldr code",
		},
		{
			exec:  generateSchema,
			skip:  opts.schemaPath == "",
			descr: "generating schema",
		},
	}

	for _, task := range tasks {
		if task.skip {
			continue
		}

		fmt.Fprint(os.Stdout, task.descr, "... ")
		if err := task.exec(opts); err != nil {
			fmt.Fprintln(os.Stdout)
			fatal(err)
		}
		fmt.Fprintln(os.Stdout, "done")
	}
}

func generateCldr(opts options) error {
	files := filetree.New(opts.cldrDataDir)
	data, err := cldr.Decode(files)
	files.Close()
	if err != nil {
		return err
	}

	header := ""
	if opts.cldrVersion != "" {
		header = "CLDR version: " + opts.cldrVersion
	}

	gen, err := generator.New(generator.Options{
		OutputDirectory: opts.outDir,
		PackageName:     opts.packageName,
		Header:          header,
	})
	if err != nil {
		return err
	}

	snippets := generate_cldr.FileSnippets(opts.packageName, data)
	return gen.GenerateFiles(snippets)
}

func generateSchema(opts options) error {
	gen := mprot_generator.NewGolang(mprot_generator.GolangOptions{
		ImportRoot:   "",
		ScopedEnums:  false,
		UnwrapUnions: false,
		TypeID:       false,
	})

	err := gen.Generate(mprot_generator.Options{
		RootDirectory:    filepath.Dir(opts.schemaPath),
		GlobPatterns:     []string{filepath.Base(opts.schemaPath)},
		RemoveDeprecated: true,
		OutputDirectory:  opts.outDir,
	})
	if err != nil {
		return err
	}
	return gen.Dump()
}

func fatal(args ...any) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}
