package Kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

type Configs struct {
	Topic         string
	WriteInterval int
	Brokers       []string
}

type Counter struct {
	Mutex     *sync.RWMutex
	uniqueIDs map[int]struct{}
}

type writer interface {
	WriteMessages(context.Context, ...kafka.Message) error
	Close() error
}

type Producer struct {
	WriteInterval int
	*Counter
	topic string
	writer
}

type ProducerConfig struct {
	Topic           string
	BrokerAddresses []string
}

func New(c Configs) Producer {
	counter := &Counter{Mutex: &sync.RWMutex{}, uniqueIDs: make(map[int]struct{})}
	p := ProducerConfig{Topic: c.Topic, BrokerAddresses: c.Brokers}
	producer, err := InitializeProducerFromConfigs(p)
	if err != nil {
		return Producer{}
	}
	fw := Producer{WriteInterval: c.WriteInterval, Counter: counter, writer: producer}
	go fw.logUniqueRequests()
	return fw
}

func InitializeProducerFromConfigs(config ProducerConfig) (Producer, error) {
	w := &kafka.Writer{
		Addr:  kafka.TCP(config.BrokerAddresses...),
		Topic: config.Topic,
	}

	return Producer{topic: w.Topic, writer: w}, nil
}

func (fw Producer) IncrementCounter(idValue int) {
	fw.Mutex.Lock()
	defer fw.Mutex.Unlock()
	fw.uniqueIDs[idValue] = struct{}{}

}

func (fw Producer) logUniqueRequests() {
	for {
		time.Sleep(30 * time.Second)
		fw.ProduceHelper()
		fw.uniqueIDs = make(map[int]struct{}) // Reset the store every minute
	}
}

func (fw Producer) ProduceHelper() {
	fw.Mutex.Lock()
	defer fw.Mutex.Unlock()
	uniqueRequests := len(fw.uniqueIDs)

	// Log the unique request count
	err := fw.Produce(context.Background(), "", []byte(strconv.Itoa(uniqueRequests)))
	if err != nil {
		slog.Error("error while producing message", "error", err)
		return
	}
	log.Printf("Unique requests in the last minute: %d\n", uniqueRequests)
}

func (fw Producer) GetValue() int {
	fw.Mutex.RLock()
	defer fw.Mutex.RUnlock()
	return len(fw.uniqueIDs)
}

func (fw Producer) Produce(ctx context.Context, key string, payload []byte) error {

	// In the case of invalid topics, the message will be published to the 'bad_event' topic.
	//if len(topic) == 0 {
	//	slog.Warn("Unable to find topic for the event", "eventName", eventName, "topic", topic, "osType", osType)
	//	metrics.BadTopicEvents.Inc()
	//	topic = p.badTopic
	//}

	err := fw.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: payload,
		Topic: fw.topic,
	})

	if err != nil {
		return err
	}

	return nil
}
