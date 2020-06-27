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
}

// Swap ...
func (pq WordQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	// pq[i].index = i
	// pq[j].index = j
}

// Push ...
func (pq *WordQueue) Push(x interface{}) {
	// n := len(*pq)
	item := x.(*Word)
	// item.index = n
	item.Count++
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
	scanner.Split(bufio.ScanWords)

	// var words []string
	numWords := 0
	for scanner.Scan() {
		numWords++
		word := strings.ToLower(scanner.Text())
		mutex.Lock()
		// wc := wordList[word] + 1
		// wordList[word] = wc

		heap.Push(&wq, word)

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

func addToTop10(word string) {

	for k, v := range top10 {

	}

}
