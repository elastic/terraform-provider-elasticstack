// Package main implements a targeted acceptance test package selector.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindAccTestPackages walks root (typically "internal") and returns the import
// paths of all Go packages that define at least one func TestAcc in a
// *_test.go file.
func FindAccTestPackages(root, modulePath string) ([]string, error) {
	seen := make(map[string]struct{})

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if !strings.HasSuffix(name, "_test.go") {
			return nil
		}

		ok, err := isAccTestFile(path)
		if err != nil {
			return fmt.Errorf("scan %s: %w", path, err)
		}
		if !ok {
			return nil
		}

		dir := filepath.Dir(path)
		if _, exists := seen[dir]; exists {
			return nil
		}
		seen[dir] = struct{}{}
		return nil
	}

	if err := filepath.WalkDir(root, walkFn); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(seen))
	for dir := range seen {
		result = append(result, modulePath+"/"+filepath.ToSlash(dir))
	}
	return stringsSorted(result), nil
}

// isAccTestFile reports whether path declares at least one func TestAcc.
func isAccTestFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, data, parser.AllErrors)
	if err != nil {
		// Files that cannot be parsed cannot declare valid acceptance tests.
		return false, nil
	}

	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if ok && strings.HasPrefix(fn.Name.Name, "TestAcc") {
			return true, nil
		}
	}
	return false, nil
}
