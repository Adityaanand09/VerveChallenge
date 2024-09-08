package FileWriter

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

//import (
//	"log"
//	"log/slog"
//	"os"
//)

type Configs struct {
	FileName      string
	WriteInterval int
}

type Counter struct {
	Mutex     *sync.RWMutex
	uniqueIDs map[int]struct{}
}

type FileWriter struct {
	fileName      string
	WriteInterval int
	*Counter
}

func New(c Configs) FileWriter {
	counter := &Counter{Mutex: &sync.RWMutex{}, uniqueIDs: make(map[int]struct{})}
	go counter.updateUniqueIds(c.FileName, c.WriteInterval)
	return FileWriter{fileName: c.FileName, WriteInterval: c.WriteInterval, Counter: counter}
}

func (r *Counter) IncrementCounter(idValue int) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.uniqueIDs[idValue] = struct{}{}

}

func (r *Counter) updateUniqueIds(fileName string, writeInterval int) {
	for {
		time.Sleep(time.Duration(writeInterval) * time.Minute)
		r.Write(fileName)
		r.uniqueIDs = make(map[int]struct{}) // Reset the store every minute
	}
}

func (r *Counter) Write(fileName string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	uniqueRequests := len(r.uniqueIDs)

	// Log the unique request count
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
	}
	file.Write([]byte(strconv.Itoa(uniqueRequests) + "\n"))
	log.Printf("Unique requests in the last minute: %d\n", uniqueRequests)
}

func (r *Counter) GetValue() int {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	return len(r.uniqueIDs)
}
