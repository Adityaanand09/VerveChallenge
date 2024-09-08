package FileWriter

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	redisConn *redis.Client
	RedisKey  string
}

type Configs struct {
	FileName         string
	WriteIntervalMin int
	RedisServer      string
	RedisPassword    string
	RedisKey         string
}

type Counter struct {
	Mutex        *sync.RWMutex
	uniqueIDs    map[int]struct{}
	redisConfigs RedisConfig
}

type FileWriter struct {
	fileName      string
	WriteInterval int
	*Counter
}

func initRedisClient(server, password, key string) RedisConfig {
	rdb := redis.NewClient(&redis.Options{
		Addr:     server,   // Redis server address
		Password: password, // No password
		DB:       0,        // Default DB
	})
	return RedisConfig{
		redisConn: rdb,
		RedisKey:  key,
	}
}

func New(c Configs) FileWriter {
	rdb := initRedisClient(c.RedisServer, c.RedisPassword, c.RedisKey)
	counter := &Counter{Mutex: &sync.RWMutex{}, uniqueIDs: make(map[int]struct{}), redisConfigs: rdb}
	go counter.logUniqueRequests(c.WriteIntervalMin, c.FileName, c.RedisKey)
	return FileWriter{fileName: c.FileName, WriteInterval: c.WriteIntervalMin, Counter: counter}
}

func (r *Counter) IncrementCounter(idValue int) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.addIfUnique(strconv.Itoa(idValue))
}

func (r *Counter) getUniqueIDCount() (int, error) {
	currentTimestamp := float64(time.Now().Unix())

	// Count the number of IDs within the last 60 seconds
	val, _ := r.redisConfigs.redisConn.ZCount(context.Background(), r.redisConfigs.RedisKey, fmt.Sprintf("%f", currentTimestamp-60), fmt.Sprintf("%f", currentTimestamp)).Result()
	return strconv.Atoi(strconv.FormatInt(val, 10))
}

func (r *Counter) removeOldIDs(ctx context.Context, setKey string) error {
	// Get the timestamp for 60 seconds ago
	oneMinuteAgo := float64(time.Now().Add(-60 * time.Second).Unix())

	// Remove all IDs older than one minute
	return r.redisConfigs.redisConn.ZRemRangeByScore(ctx, setKey, "-inf", fmt.Sprintf("%f", oneMinuteAgo)).Err()
}

func (r *Counter) addIfUnique(id string) {
	timestamp := float64(time.Now().Unix())
	err := r.redisConfigs.redisConn.ZAdd(context.Background(), "unique_ids", &redis.Z{
		Score:  timestamp,
		Member: id,
	}).Err()
	if err != nil {
		log.Fatal(err)
	}
	return // Returns true if it's a new ID, false otherwise
}

func (r *Counter) logUniqueRequests(writeInterval int, fileName, redisKey string) {
	for {
		time.Sleep(time.Duration(writeInterval) * time.Minute)
		r.Write(fileName)
		r.removeOldIDs(context.Background(), redisKey)
		r.uniqueIDs = make(map[int]struct{}) // Reset the store every minute
	}
}

func (r *Counter) Write(fileName string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	uniqueRequests, err := r.getUniqueIDCount()
	if err != nil {
		return
	}

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
	count, err := r.getUniqueIDCount()
	if err != nil {
		return 0
	}
	return count
}
