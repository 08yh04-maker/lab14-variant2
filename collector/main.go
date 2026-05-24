package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// WeatherData структура для хранения погодных данных
type WeatherData struct {
	City        string    `json:"city"`
	Temperature float64   `json:"temperature"`
	Humidity    int       `json:"humidity"`
	Pressure    int       `json:"pressure"`
	WindSpeed   float64   `json:"wind_speed"`
	Condition   string    `json:"condition"`
	Timestamp   time.Time `json:"timestamp"`
}

// Collector структура сборщика
type Collector struct {
	client     *resty.Client
	etcdClient *clientv3.Client
	cities     []string
	apiKey     string
	dataChan   chan WeatherData
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewCollector создаёт новый сборщик
func NewCollector(apiKey string, cities []string) (*Collector, error) {
	// Подключаемся к etcd
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("Warning: etcd not available, running in standalone mode: %v", err)
		etcdClient = nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		client:     resty.New().SetTimeout(10 * time.Second),
		etcdClient: etcdClient,
		cities:     cities,
		apiKey:     apiKey,
		dataChan:   make(chan WeatherData, 1000),
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// fetchWeather собирает погоду для одного города
func (c *Collector) fetchWeather(city string) (*WeatherData, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric&lang=ru", city, c.apiKey)
	
	var response struct {
		Main struct {
			Temp     float64 `json:"temp"`
			Pressure int     `json:"pressure"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
	}

	resp, err := c.client.R().SetResult(&response).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode())
	}

	condition := ""
	if len(response.Weather) > 0 {
		condition = response.Weather[0].Description
	}

	return &WeatherData{
		City:        city,
		Temperature: response.Main.Temp,
		Humidity:    response.Main.Humidity,
		Pressure:    response.Main.Pressure,
		WindSpeed:   response.Wind.Speed,
		Condition:   condition,
		Timestamp:   time.Now(),
	}, nil
}

// saveToFile сохраняет данные в JSON-файл
func (c *Collector) saveToFile(data WeatherData) error {
	file, err := os.OpenFile("data/weather.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(jsonData) + "\n")
	return err
}

// runWorker запускает воркер для сбора данных
func (c *Collector) runWorker(city string) {
	defer c.wg.Done()
	
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("Worker for %s stopping", city)
			return
		case <-ticker.C:
			log.Printf("Fetching weather for %s", city)
						data, err := c.fetchWeather(city)
			if err != nil {
				log.Printf("Error fetching %s: %v", city, err)
				continue
			}
			
			// Валидация через Rust-библиотеку
			result := ValidateWeatherData(*data)
			if !result.IsValid {
				log.Printf("Validation failed for %s: %s", city, result.ErrorMessage)
				continue
			}
			
			c.dataChan <- *data
		}
	}
}

// writeWorker записывает данные в файл
func (c *Collector) writeWorker() {
	buffer := make([]WeatherData, 0, 100)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			// Записываем остатки при завершении
			for _, data := range buffer {
				c.saveToFile(data)
			}
			return
		case data := <-c.dataChan:
			buffer = append(buffer, data)
			if len(buffer) >= 100 {
				for _, d := range buffer {
					c.saveToFile(d)
				}
				buffer = buffer[:0]
				log.Printf("Flushed 100 records to file")
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				for _, d := range buffer {
					c.saveToFile(d)
				}
				buffer = buffer[:0]
				log.Printf("Flushed %d records to file (timeout)", len(buffer))
			}
		}
	}
}

// Run запускает сборщик
func (c *Collector) Run() {
	// Создаём папку для данных
	os.MkdirAll("data", 0755)

	// Запускаем воркеров для каждого города
	for _, city := range c.cities {
		c.wg.Add(1)
		go c.runWorker(city)
	}

	// Запускаем writer
	go c.writeWorker()

	log.Printf("Collector started with %d cities", len(c.cities))
}

// Stop останавливает сборщик
func (c *Collector) Stop() {
	log.Println("Shutting down collector...")
	c.cancel()
	c.wg.Wait()
	close(c.dataChan)
	log.Println("Collector stopped")
}

func main() {
	// Получаем API ключ из переменной окружения
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		apiKey = "test_key" // Для тестирования без API
		log.Println("Warning: OPENWEATHER_API_KEY not set, using test mode")
	}

	cities := []string{
		"Moscow",
		"Saint Petersburg",
		"Novosibirsk",
		"Yekaterinburg",
		"Kazan",
		"Nizhny Novgorod",
		"Chelyabinsk",
		"Omsk",
		"Samara",
		"Rostov-on-Don",
	}

	collector, err := NewCollector(apiKey, cities)
	if err != nil {
		log.Fatal(err)
	}

	// Обработка сигналов graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// Запускаем Arrow сервер в отдельной горутине
	go StartArrowServer()

	collector.Run()

	<-sigChan
	log.Println("Received shutdown signal")
	collector.Stop()
}