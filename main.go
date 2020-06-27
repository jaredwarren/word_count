package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"
)

/**
* in Linux, design a multithreaded program that will
* take a directory, traverse the directory,
* read all files ending in *.txt ,
* count the words collectively in all files (where words is unambiguously defined),
* and print out the top 10 words.
 */

var (
	wg    sync.WaitGroup
	mutex sync.Mutex
	wq    WordQueue
)

// TOP number of most words to print
const TOP = 10

func main() {
	args := os.Args

	// Get path from args (default to current working dir)
	searchPath := "./"
	if len(args) > 1 {
		searchPath = args[1]
	}

	// validate input path
	pi, err := os.Stat(searchPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if !pi.IsDir() {
		log.Fatalf("Input path must be directory")
		os.Exit(1)
	}

	// Init word queue
	wq = WordQueue{}
	heap.Init(&wq)

	// Start walking directory
	wg.Add(1)
	walkDir(searchPath)
	wg.Wait()

	// Print out top words
	for i := 0; i < TOP; i++ {
		// check that there are even enough words
		if i > len(wq)-1 {
			break
		}
		word := heap.Pop(&wq).(*Word)
		fmt.Println(word.value, word.count)
	}
}

func walkDir(dir string) error {
	defer wg.Done()

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != dir {
			wg.Add(1)
			go walkDir(path)

			// Tell walk to not scan sub-directories, since we're walking concurrently
			return filepath.SkipDir
		}

		// Scan all regular files ending in ".txt"
		if !info.IsDir() && info.Mode().IsRegular() && strings.HasSuffix(path, ".txt") {
			go func() {
				wg.Add(1)
				readFile(path)
				wg.Done()
			}()
		}
		return nil
	})
}

// readFile scans a file and adds words to the word queue
func readFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
		return err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(ScanWords)

	numWords := 0
	for scanner.Scan() {
		numWords++
		word := strings.ToLower(scanner.Text())

		// Add word to queue
		mutex.Lock()
		heap.Push(&wq, &Word{value: word})
		mutex.Unlock()
	}
	return nil
}

/**
* Word Queue Heap
 */

// A Word ...
type Word struct {
	value string // The word
	count int    // The total number of occurances
	index int    // The index of the word in the heap.
}

// A WordQueue implements heap.Interface and holds Words.
type WordQueue []*Word

// Len implements sort.Interface returns current length of queue
func (wq WordQueue) Len() int {
	return len(wq)
}

// Less implements sort.Interface compares word.count
func (wq WordQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, count so we use greater than here.
	return wq[i].count > wq[j].count
}

// Swap implements sort.Interface
func (wq WordQueue) Swap(i, j int) {
	wq[i], wq[j] = wq[j], wq[i]
	wq[i].index = i
	wq[j].index = j
}

// Push implements heap.Interface adds new words to queue, increments count of existing words
func (wq *WordQueue) Push(x interface{}) {
	item := x.(*Word)
	_, w := wq.Find(item.value)

	// don't add duplicates to queue
	if w == nil {
		n := len(*wq)
		item.index = n
		item.count = 1
		*wq = append(*wq, item)
	} else {
		w.count++
		heap.Fix(wq, w.index)
	}
}

// Pop implements heap.interface removes word from queue
func (wq *WordQueue) Pop() interface{} {
	old := *wq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*wq = old[0 : n-1]
	return item
}

// Find returns the index and *Word for a given string
func (wq *WordQueue) Find(word string) (int, *Word) {
	for i, v := range *wq {
		if v.value == word {
			return i, v
		}
	}
	return 0, nil
}

/**
* The following functions were copied from the standard libary (bufio/scan.go)
* and have been modified to ignore puncuation when scanning for words
 */

// isSpace reports whether the character is a Unicode white space character.
// We avoid dependency on the unicode package, but check validity of the implementation
// in the tests.
func isSpace(r rune) bool {
	if r <= '\u00FF' {
		// Obvious ASCII ones: \t through \r plus space. Plus two Latin-1 oddballs.
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r', ',', '.', '-', '_', '?', '!', ';', ':', '=', '>', '<':
			return true
		case '\u0085', '\u00A0':
			return true
		}
		return false
	}
	// High-valued ones.
	if '\u2000' <= r && r <= '\u200a' {
		return true
	}
	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}
	return false
}

// ScanWords is a split function for a Scanner that returns each
// space-separated word of text, with surrounding spaces deleted. It will
// never return an empty string. The definition of space is set by
// unicode.IsSpace.
func ScanWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !isSpace(r) {
			break
		}
	}
	// Scan until space, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if isSpace(r) {
			return i + width, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}
