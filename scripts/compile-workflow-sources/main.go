package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr *os.File) error {
	fs := flag.NewFlagSet("compile-workflow-sources", flag.ContinueOnError)
	fs.SetOutput(stderr)

	check := fs.Bool("check", false, "check whether generated workflow sources are up to date")
	manifest := fs.String("manifest", "", "path to workflow manifest")
	output := fs.String("output", "", "path to generated workflow output")
	root := fs.String("root", "", "repository root")
	template := fs.String("template", "", "path to workflow template")

	if err := fs.Parse(args); err != nil {
		return err
	}

	rootDir := *root
	if rootDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		rootDir = wd
	}

	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return err
	}

	manifestPath := *manifest
	if manifestPath == "" {
		manifestPath = filepath.Join(rootDir, ".github", "workflows-src", "manifest.json")
	} else {
		manifestPath = resolvePath(rootDir, manifestPath)
	}

	var results []CompileResult

	switch {
	case *template != "" || *output != "":
		if *template == "" || *output == "" {
			return fmt.Errorf("both --template and --output are required")
		}

		result, err := CompileWorkflow(CompileOptions{
			TemplatePath: resolvePath(rootDir, *template),
			OutputPath:   resolvePath(rootDir, *output),
			RootDir:      rootDir,
			Check:        *check,
		})
		if err != nil {
			return err
		}
		results = []CompileResult{result}
	default:
		results, err = CompileFromManifest(manifestPath, rootDir, *check)
		if err != nil {
			return err
		}
	}

	if *check {
		var changed []string
		for _, result := range results {
			if result.Changed {
				changed = append(changed, "- "+normalizeRelativePath(rootDir, result.OutputPath))
			}
		}
		if len(changed) > 0 {
			return fmt.Errorf("generated workflow sources are out of date:\n%s", strings.Join(changed, "\n"))
		}
		return nil
	}

	for _, result := range results {
		_, _ = fmt.Fprintf(stdout, "Generated %s\n", normalizeRelativePath(rootDir, result.OutputPath))
	}

	return nil
}
