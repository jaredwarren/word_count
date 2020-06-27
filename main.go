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

var wg sync.WaitGroup
var mutex sync.Mutex
var wordList map[string]int
var top10 map[string]int
var min int

// Word ...
type Word struct {
	String string
	Count  int
}

// A WordQueue implements heap.Interface and holds Items.
type WordQueue []*Word

// Len ...
func (pq WordQueue) Len() int { return len(pq) }

// Less ...
func (pq WordQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Count > pq[j].Count
	// si := pq[i]
	// return wordList[*pq[i]] > wordList[*pq[j]]
}

// Swap ...
func (pq WordQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	// pq[i].index = i
	// pq[j].index = j
}

// Push ...
func (pq *WordQueue) Push(x interface{}) {
	// item := x.(*string)
	item := x.(*Word)
	*pq = append(*pq, item)
}

// Pop ...
func (pq *WordQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	// item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

/**
* main
 */

var wq WordQueue

func main() {
	wq = WordQueue{}
	heap.Init(&wq)

	wordList = map[string]int{}
	top10 = map[string]int{}
	wg.Add(1)
	walkDir("./test")
	wg.Wait()

	for i := 0; i < 10; i++ {
		// word := heap.Pop(&wq).(*string)
		// fmt.Println(*word, wordList[*word])

		word := heap.Pop(&wq).(*Word)
		fmt.Println(word.String, word.Count)
	}

	fmt.Println(wordList)
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
	scanner.Split(ScanWords)
	// scanner.Split(bufio.ScanWords)
	// scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// 	if atEOF && len(data) == 0 {
	// 		return 0, nil, nil
	// 	}
	// 	if i := strings.Index(string(data), "\n#"); i >= 0 {
	// 		return i + 1, data[0:i], nil
	// 	}

	// 	if atEOF {
	// 		return len(data), data, nil
	// 	}

	// 	return
	// })

	// var words []string
	numWords := 0
	for scanner.Scan() {
		numWords++
		word := strings.ToLower(scanner.Text())
		mutex.Lock()
		wordList[word]++
		heap.Push(&wq, &Word{word, wordList[word]})
		// heap.Push(&wq, &word)

		// heap.Push(&pq, item)
		// if wc > min {
		// 	addToTop10(word)
		// }
		mutex.Unlock()
		// TODO: if case
		// words = append(words, scanner.Text())
	}

	fmt.Println("word list:", numWords)
	// for _, word := range words {
	// 	fmt.Println(word)
	// }
}

// isSpace reports whether the character is a Unicode white space character.
// We avoid dependency on the unicode package, but check validity of the implementation
// in the tests.
func isSpace(r rune) bool {
	if r <= '\u00FF' {
		// Obvious ASCII ones: \t through \r plus space. Plus two Latin-1 oddballs.
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r', ',', '.', '-', '_':
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
