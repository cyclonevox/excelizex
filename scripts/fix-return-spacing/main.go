// One-off: insert blank line before return when preceded by another statement in the same block.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	root := "v2"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	changed := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		ok, err := fixFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if ok {
			changed++
			fmt.Println(path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("changed %d files\n", changed)
}

func fixFile(path string) (bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(src), "\n")
	insertBefore := map[int]struct{}{}

	var inspect func(node ast.Node) bool
	inspect = func(node ast.Node) bool {
		if node == nil {
			return true
		}
		switch n := node.(type) {
		case *ast.BlockStmt:
			markReturns(fset, lines, n.List, insertBefore)
		case *ast.CaseClause:
			markReturns(fset, lines, n.Body, insertBefore)
		case *ast.CommClause:
			markReturns(fset, lines, n.Body, insertBefore)
		}
		return true
	}
	ast.Inspect(f, inspect)

	if len(insertBefore) == 0 {
		return false, nil
	}

	// Apply from bottom to top so line numbers stay valid.
	sorted := make([]int, 0, len(insertBefore))
	for line := range insertBefore {
		sorted = append(sorted, line)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))

	for _, line := range sorted {
		idx := line - 1
		if idx < 0 || idx > len(lines) {
			continue
		}
		lines = append(lines[:idx], append([]string{""}, lines[idx:]...)...)
	}

	out := strings.Join(lines, "\n")
	if len(src) > 0 && src[len(src)-1] == '\n' && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	if out == string(src) {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(out), infoMode(path))
}

func infoMode(path string) os.FileMode {
	fi, err := os.Stat(path)
	if err != nil {
		return 0o644
	}
	return fi.Mode()
}

func markReturns(fset *token.FileSet, lines []string, stmts []ast.Stmt, insertBefore map[int]struct{}) {
	for i := 1; i < len(stmts); i++ {
		ret, ok := stmts[i].(*ast.ReturnStmt)
		if !ok {
			continue
		}
		prevEnd := fset.Position(stmts[i-1].End()).Line
		retLine := fset.Position(ret.Pos()).Line
		if needsBlankLine(lines, retLine, prevEnd) {
			insertBefore[retLine] = struct{}{}
		}
	}
}

func needsBlankLine(lines []string, retLine, prevEnd int) bool {
	if retLine <= prevEnd+1 {
		return retLine == prevEnd+1
	}
	if retLine == prevEnd+2 {
		return strings.TrimSpace(lines[prevEnd]) != ""
	}
	return false
}
