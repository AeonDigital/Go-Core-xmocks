package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("xmocks", flag.ContinueOnError)
	fs.SetOutput(stderr)

	filePath := fs.String("file", "", "Path to the Go source file containing the interface")
	interfaceName := fs.String("interface", "", "Name of the interface to mock")
	alias := fs.String("alias", "", "Alias to use for generated mock names; defaults to interface name")
	output := fs.String("output", "", "Output directory or file path for generated mock; defaults to internal/pkgxmock/<alias>.go")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *filePath == "" || *interfaceName == "" {
		fmt.Fprintln(stderr, "error: both --file and --interface are required")
		fs.Usage()
		return 1
	}

	opts := GeneratorOptions{
		InputFile:     *filePath,
		InterfaceName: *interfaceName,
	}

	if *alias != "" {
		opts.Alias = *alias
	}

	if *output != "" {
		opts.OutputPath = *output
	}

	outputPath, err := GenerateMockFile(opts)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Mock generated: %s\n", outputPath)
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
