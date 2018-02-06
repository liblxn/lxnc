# lxnc
`lxnc` is a compiler for [lxn](https://github.com/liblxn/lxn) to translate schema definitions into the binary representation, which is used by the client libraries.

## Usage
```
Synopsis:
  lxnc [options] [schema-file ...]

Description:
  The lxnc tool compiles lxn schema definition files into their binary representation.

  The following options are available:
    --locale <locale tag>
        Specify the locale for which the lxnc catalog should be created. This option is
        required.
    --out <path>
        Specify the output path of the generated lxnc catalog files. The default is the
        current directory.
    --merge
        Merge all input files into a single catalog.
        If --merge-into is not specified, the catalog is named after the locale.
    --merge-into <catalog name>
        Specify the name of the merged catalog. This flag automatically sets
        the --merge option.
```
