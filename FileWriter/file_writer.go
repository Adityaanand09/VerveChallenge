package FileWriter

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type Configs struct {
	FileName      string
	WriteInterval int
}

type Counter struct {
	Mutex     *sync.RWMutex
	uniqueIDs map[int]struct{}
	redisConn *redis.Client
}

type FileWriter struct {
	fileName      string
	WriteInterval int
	*Counter
}

func initRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password
		DB:       0,                // Default DB
	})
	return rdb
}

func New(c Configs) FileWriter {
	rdb := initRedisClient()
	counter := &Counter{Mutex: &sync.RWMutex{}, uniqueIDs: make(map[int]struct{}), redisConn: rdb}
	go counter.logUniqueRequests()
	return FileWriter{fileName: c.FileName, WriteInterval: c.WriteInterval, Counter: counter}
}

func (r *Counter) IncrementCounter(idValue int) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.addIfUnique(strconv.Itoa(idValue))
	//r.uniqueIDs[idValue] = struct{}{}

}

//func (r *Counter) incrementUniqueIDCount() error {
//	// Increment the counter for unique IDs
//	_, err := r.redisConn.Incr(context.Background(), "unique_id_count").Result()
//	//_, err :=r.redisConn.SCard(context.Background(),"unique_ids").Result()
//	return err
//}

func (r *Counter) getUniqueIDCount() (int, error) {
	// Retrieve the unique ID count
	//val, err := r.redisConn.SCard(context.Background(), "unique_ids").Result()
	//if err != nil {
	//	return 0, err
	//}
	currentTimestamp := float64(time.Now().Unix())

	// Count the number of IDs within the last 60 seconds
	val, _ := r.redisConn.ZCount(context.Background(), "unique_ids", fmt.Sprintf("%f", currentTimestamp-60), fmt.Sprintf("%f", currentTimestamp)).Result()
	return strconv.Atoi(strconv.FormatInt(val, 10))
}

func (r *Counter) removeOldIDs(ctx context.Context, setKey string) error {
	// Get the timestamp for 60 seconds ago
	oneMinuteAgo := float64(time.Now().Add(-60 * time.Second).Unix())

	// Remove all IDs older than one minute
	return r.redisConn.ZRemRangeByScore(ctx, setKey, "-inf", fmt.Sprintf("%f", oneMinuteAgo)).Err()
}

func (r *Counter) addIfUnique(id string) {
	timestamp := float64(time.Now().Unix())
	err := r.redisConn.ZAdd(context.Background(), "unique_ids", &redis.Z{
		Score:  timestamp,
		Member: id,
	}).Err()
	if err != nil {
		log.Fatal(err)
	}
	return // Returns true if it's a new ID, false otherwise
}

func (r *Counter) logUniqueRequests() {
	for {
		time.Sleep(30 * time.Second)
		r.Write()
		r.removeOldIDs(context.Background(), "unique_ids")
		r.uniqueIDs = make(map[int]struct{}) // Reset the store every minute
	}
}

func (r *Counter) Write() {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	uniqueRequests, err := r.getUniqueIDCount()
	if err != nil {
		return
	}

	// Log the unique request count
	file, err := os.OpenFile("uniqueCount.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
	}
	file.Write([]byte(strconv.Itoa(uniqueRequests) + "\n"))
	log.Printf("Unique requests in the last minute: %d\n", uniqueRequests)
}

func (r *Counter) GetValue() int {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	count, err := r.getUniqueIDCount()
	if err != nil {
		return 0
	}
	return count
}
