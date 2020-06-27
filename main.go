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
		}
		return nil
	})
}

func glob(dir string) (m []string, e error) {

	e = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".txt") {
			fmt.Println(path)
		}
		// if err != nil {
		// 	fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
		// 	return err
		// }
		// if info.IsDir() && info.Name() == subDirToSkip {
		// 	fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
		// 	return filepath.SkipDir
		// }
		// fmt.Printf("visited file or dir: %q\n", path)
		return nil
	})

	return
	fi, err := os.Stat(dir)
	if err != nil {
		return
	}
	fmt.Printf("%+v\n", fi)
	if !fi.IsDir() {
		return
	}
	d, err := os.Open(dir)
	if err != nil {
		return
	}
	defer d.Close()

	// names, _ := d.Readdirnames(-1)
	// sort.Strings(names)

	// for _, n := range names {
	// 	matched, err := Match(pattern, n)
	// 	if err != nil {
	// 		return m, err
	// 	}
	// 	if matched {
	// 		m = append(m, Join(dir, n))
	// 	}
	// }
	return
}

func readFile() {
	file, err := os.Open("test.csv")
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	var words []string

	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	fmt.Println("word list:")
	for _, word := range words {
		fmt.Println(word)
	}
}
