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
	writer  sync.Mutex
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

// Funções SEM controle de prioridade para leitores e escritores
func (db *Database) read(pos int) string {
	db.mutex.Lock() // Bloqueio geral para leitura e escrita
	defer db.mutex.Unlock()
	return db.data[pos]
}

func (db *Database) write(pos int, value string) {
	db.mutex.Lock() // Bloqueio geral para leitura e escrita
	defer db.mutex.Unlock()
	db.data[pos] = value
}

// Funções COM controle de prioridade para leitores e escritores
func (db *Database) readWithPriority(pos int) string {
	db.mutex.Lock()
	db.readers++
	if db.readers == 1 {
		db.writer.Lock() // O primeiro leitor bloqueia os escritores
	}
	db.mutex.Unlock()

	// Realiza a leitura
	result := db.data[pos]

	db.mutex.Lock()
	db.readers--
	if db.readers == 0 {
		db.writer.Unlock() // O último leitor libera os escritores
	}
	db.mutex.Unlock()

	return result
}

func (db *Database) writeWithPriority(pos int, value string) {
	db.writer.Lock() // Escritores bloqueiam outros leitores e escritores
	defer db.writer.Unlock()

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

// Funções de leitura e escrita com prioridade (para threads)
func readerWithPriority(db *Database, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		pos := rand.Intn(len(db.data))
		_ = db.readWithPriority(pos) // Leitura apenas para simulação
	}
	time.Sleep(1 * time.Millisecond)
}

func writerWithPriority(db *Database, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		pos := rand.Intn(len(db.data))
		db.writeWithPriority(pos, "MODIFICADO")
	}
	time.Sleep(1 * time.Millisecond)
}

// Função para executar o experimento sem prioridade
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

// Função para executar o experimento com prioridade para leitores
func runExperimentWithPriority(readers, writers int) time.Duration {
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
		threads[i] = func() { readerWithPriority(db, &wg) }
	}
	for i := readers; i < readers+writers; i++ {
		threads[i] = func() { writerWithPriority(db, &wg) }
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

// Main function
func main() {
	var proportions [][2]int
	for i := 0; i <= 100; i++ {
		proportions = append(proportions, [2]int{i, 100 - i})
	}

	for _, proportion := range proportions {
		// Executa o experimento com prioridade para leitores
		var totalDurationPriority time.Duration
		for i := 0; i < 50; i++ {
			totalDurationPriority += runExperimentWithPriority(proportion[0], proportion[1])
		}
		averageDurationPriority := totalDurationPriority / 50

		// Executa o experimento sem prioridade
		var totalDurationNoPriority time.Duration
		for i := 0; i < 50; i++ {
			totalDurationNoPriority += runExperiment(proportion[0], proportion[1])
		}
		averageDurationNoPriority := totalDurationNoPriority / 50

		difference := averageDurationPriority - averageDurationNoPriority

		fmt.Printf("Readers: %d, Writers: %d, Avg Duration with Priority: %v, Avg Duration without Priority: %v, Difference: %v\n", proportion[0], proportion[1], averageDurationPriority, averageDurationNoPriority, difference)
	}
}
