package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Database struct {
	data    []string
	mutex   sync.Mutex
	readers int
}

func NewDatabase(filename string) (*Database, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return &Database{data: data}, nil
}

func (db *Database) read(pos int) string {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	return db.data[pos]
}

func (db *Database) write(pos int, value string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.data[pos] = value
}

func reader(db *Database, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		pos := rand.Intn(len(db.data))
		_ = db.read(pos) // Leitura apenas para simulação
	}
	time.Sleep(1 * time.Millisecond)
}

func writer(db *Database, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		pos := rand.Intn(len(db.data))
		db.write(pos, "MODIFICADO")
	}
	time.Sleep(1 * time.Millisecond)
}

func runExperiment(readers, writers int) time.Duration {
	db, err := NewDatabase("bd.txt")
	if err != nil {
		fmt.Println("Error loading database:", err)
		return 0
	}

	var wg sync.WaitGroup
	threadCount := 100
	threads := make([]func(), threadCount)

	// Populate threads array with readers and writers
	for i := 0; i < readers; i++ {
		threads[i] = func() { reader(db, &wg) }
	}
	for i := readers; i < readers+writers; i++ {
		threads[i] = func() { writer(db, &wg) }
	}

	// Shuffle threads array
	rand.Shuffle(threadCount, func(i, j int) { threads[i], threads[j] = threads[j], threads[i] })

	start := time.Now()
	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go threads[i]()
	}
	wg.Wait()
	return time.Since(start)
}

func main() {
	var proportions [][2]int
	for i := 0; i <= 100; i++ {
		proportions = append(proportions, [2]int{i, 100 - i})
	}

	for _, proportion := range proportions {
		var totalDuration time.Duration
		for i := 0; i < 50; i++ {
			totalDuration += runExperiment(proportion[0], proportion[1])
		}
		averageDuration := totalDuration / 50
		fmt.Printf("Readers: %d, Writers: %d, Avg Duration: %v\n", proportion[0], proportion[1], averageDuration)
	}
}
