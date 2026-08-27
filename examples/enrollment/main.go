package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	dir := flag.String("dir", defaultDir(), "过程文件输出目录")
	flag.Parse()

	svc, err := run(*dir, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n完成。报名库共 %d 人，过程文件在 %s\n", svc.Count(), *dir)
}

func defaultDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "out"
	}
	return filepath.Join(filepath.Dir(file), "out")
}
