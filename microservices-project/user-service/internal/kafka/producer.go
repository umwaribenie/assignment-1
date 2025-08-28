package kafka

import (
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
)

type Producer interface {
	SendNotification(topic string, notification interface{}) error
	Close() error
}

type kafkaProducer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	log.Println("Kafka producer connected successfully")
	return &kafkaProducer{producer: producer}, nil
}

func (p *kafkaProducer) SendNotification(topic string, notification interface{}) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		return err
	}

	log.Printf("Message sent to partition %d at offset %d", partition, offset)
	return nil
}

func (p *kafkaProducer) Close() error {
	return p.producer.Close()
}