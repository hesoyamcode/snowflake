package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	workerIDBits     = uint64(5) // 5bit workerID out of 10bit worker machine ID
	dataCenterIDBits = uint64(5) // 5bit workerID out of 10bit worker dataCenterID
	sequenceBits     = uint64(12)

	maxWorkerID     = int64(-1) ^ (int64(-1) << workerIDBits) // The maximum value of the node ID used to prevent overflow
	maxDataCenterID = int64(-1) ^ (int64(-1) << dataCenterIDBits)
	maxSequence     = int64(-1) ^ (int64(-1) << sequenceBits)

	timeLeft = uint8(22) // timeLeft = workerIDBits + sequenceBits // Timestamp offset left
	dataLeft = uint8(17) // dataLeft = dataCenterIDBits + sequenceBits
	workLeft = uint8(12) // workLeft = sequenceBits // Node IDx offset to the left

	twepoch = int64(1659674040000) // constant timestamp (milliseconds)
)

type Worker struct {
	mu           sync.Mutex
	LastStamp    int64 // Record the timestamp of the last ID
	WorkerID     int64 // the ID of the node
	DataCenterID int64 // The data center ID of the node
	Sequence     int64 // ID sequence numbers that have been generated in the current millisecond (accumulated from 0) A maximum of 4096 IDs are generated within 1 millisecond
}

// In distributed cases, we should assign each machine an independent id through an external configuration file or other means
func NewWorker(workerID, dataCenterID int64) *Worker {
	return &Worker{
		WorkerID:     workerID,
		LastStamp:    0,
		Sequence:     0,
		DataCenterID: dataCenterID,
	}
}

func (w *Worker) getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

func (w *Worker) NextID() (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.nextID()
}

func (w *Worker) nextID() (uint64, error) {
	timeStamp := w.getMilliSeconds()
	if timeStamp < w.LastStamp {
		return 0, errors.New("time is moving backwards,waiting until")
	}

	if w.LastStamp == timeStamp {

		w.Sequence = (w.Sequence + 1) & maxSequence

		if w.Sequence == 0 {
			for timeStamp <= w.LastStamp {
				timeStamp = w.getMilliSeconds()
			}
		}
	} else {
		w.Sequence = 0
	}

	w.LastStamp = timeStamp
	id := ((timeStamp - twepoch) << timeLeft) |
		(w.DataCenterID << dataLeft) |
		(w.WorkerID << workLeft) |
		w.Sequence

	return uint64(id), nil
}

var wg sync.WaitGroup

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}

func main() {
	defer timer("main function")()

	w := NewWorker(5, 5)
	ch := make(chan uint64, 100000)
	count := 100000
	wg.Add(count)
	defer close(ch)
	// Concurrently count goroutines for snowFlake ID generation
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			id, _ := w.NextID()
			ch <- id
		}()
	}
	wg.Wait()
	m := make(map[uint64]int)
	for i := 0; i < count; i++ {
		id := <-ch
		// If there is a key with id in the map, it means that the generated snowflake ID is duplicated
		_, ok := m[id]
		if ok {
			fmt.Printf("repeat id %d\n", id)
			return
		}
		// store id as key in map
		m[id] = i
	}
	// successfully generated snowflake ID
	fmt.Println("All", len(m), "snowflake ID successed!")

}
