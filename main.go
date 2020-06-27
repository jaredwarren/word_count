package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

/**
* in Linux, design a multithreaded program that will
* take a directory, traverse the directory,
* read all files ending in *.txt ,
* count the words collectively in all files (where words is unambiguously defined),
* and print out the top 10 words.
 */

func main() {
	// matches, err := glob("./test")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%+v\n", matches)

	wg.Add(1)
	walkDir("./test")
	wg.Wait()
}

var wg sync.WaitGroup

func walkDir(dir string) error {
	defer wg.Done()

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != dir {
			wg.Add(1)
			go walkDir(path)

			// Tell walk to not scan sub-directories, since we're walking concurrently
			return filepath.SkipDir
		}
		if !info.IsDir() && info.Mode().IsRegular() && strings.HasSuffix(path, ".txt") {
			fmt.Println(path)
			readFile(path)
		}
		return nil
	})
}

func readFile(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	// var words []string
	numWords := 0
	for scanner.Scan() {
		numWords++
		// words = append(words, scanner.Text())
	}

	fmt.Println("word list:", numWords)
	// for _, word := range words {
	// 	fmt.Println(word)
	// }
}
