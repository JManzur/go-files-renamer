package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	dir := flag.String("dir", ".", "the root directory to search")
	verbose := flag.Bool("v", false, "enable verbose output")
	flag.Parse()

	re := regexp.MustCompile(`[A-Z]`)

	file, err := os.Create("output.log")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)

	err = filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name()[0] == '.' {
			return filepath.SkipDir
		}
		if re.MatchString(info.Name()) {
			logger.Println(path)
			if *verbose {
				fmt.Println(path)
			}
		}
		return nil
	})
	if err != nil {
		logger.Println(err)
	}
}
