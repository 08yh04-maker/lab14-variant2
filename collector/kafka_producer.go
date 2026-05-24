package main

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

func StartKafkaProducer(dataChan <-chan WeatherData) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = 5
	
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Printf("Kafka not available: %v", err)
		return
	}
	defer producer.Close()
	
	for data := range dataChan {
		jsonData, _ := json.Marshal(data)
		msg := &sarama.ProducerMessage{
			Topic: "weather-data",
			Value: sarama.StringEncoder(jsonData),
		}
		_, _, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("Failed to send: %v", err)
		}
	}
}