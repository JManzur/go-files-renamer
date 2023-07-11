package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func main() {
	var dir string
	flag.StringVar(&dir, "dir", ".", "directory path")
	flag.Parse()

	listSubdirs(dir, 0)
}

func listSubdirs(dir string, indent int) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() && !isHidden(file.Name()) {
			fmt.Printf("%s%s/\n", strings.Repeat("  ", indent), file.Name())
			listSubdirs(filepath.Join(dir, file.Name()), indent+1)
		}
	}
}

func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}
