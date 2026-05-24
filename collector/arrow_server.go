package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/ipc"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"google.golang.org/grpc"
)

type ArrowServer struct {
	// gRPC сервер (упрощённо: будем отдавать файл через HTTP)
}

// ReadAggregatedData читает агрегированные данные из JSON-файла
func ReadAggregatedData() ([]AggregatedData, error) {
	file, err := os.Open("data/aggregated.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []AggregatedData
	decoder := json.NewDecoder(file)
	for {
		var record AggregatedData
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		data = append(data, record)
	}
	return data, nil
}

// ConvertToArrow преобразует данные в Arrow RecordBatch
func ConvertToArrow(data []AggregatedData) (*arrow.RecordBatch, error) {
	pool := memory.NewGoAllocator()
	
	// Определяем схемы полей
	fields := []arrow.Field{
		{Name: "city", Type: arrow.BinaryTypes.String},
		{Name: "window_start", Type: arrow.FixedWidthTypes.Timestamp_us},
		{Name: "window_end", Type: arrow.FixedWidthTypes.Timestamp_us},
		{Name: "avg_temperature", Type: arrow.PrimitiveTypes.Float64},
		{Name: "min_temperature", Type: arrow.PrimitiveTypes.Float64},
		{Name: "max_temperature", Type: arrow.PrimitiveTypes.Float64},
		{Name: "avg_humidity", Type: arrow.PrimitiveTypes.Float64},
		{Name: "avg_wind_speed", Type: arrow.PrimitiveTypes.Float64},
		{Name: "sample_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "most_common_condition", Type: arrow.BinaryTypes.String},
	}
	schema := arrow.NewSchema(fields, nil)

	// Создаём билдеры
	bldrCity := array.NewStringBuilder(pool)
	bldrWindowStart := array.NewTimestampBuilder(pool, arrow.Us)
	bldrWindowEnd := array.NewTimestampBuilder(pool, arrow.Us)
	bldrAvgTemp := array.NewFloat64Builder(pool)
	bldrMinTemp := array.NewFloat64Builder(pool)
	bldrMaxTemp := array.NewFloat64Builder(pool)
	bldrAvgHumidity := array.NewFloat64Builder(pool)
	bldrAvgWindSpeed := array.NewFloat64Builder(pool)
	bldrSampleCount := array.NewInt64Builder(pool)
	bldrCondition := array.NewStringBuilder(pool)

	defer func() {
		bldrCity.Release()
		bldrWindowStart.Release()
		bldrWindowEnd.Release()
		bldrAvgTemp.Release()
		bldrMinTemp.Release()
		bldrMaxTemp.Release()
		bldrAvgHumidity.Release()
		bldrAvgWindSpeed.Release()
		bldrSampleCount.Release()
		bldrCondition.Release()
	}()

	for _, d := range data {
		bldrCity.Append(d.City)
		bldrWindowStart.Append(arrow.Timestamp(d.WindowStart.UnixMicro()))
		bldrWindowEnd.Append(arrow.Timestamp(d.WindowEnd.UnixMicro()))
		bldrAvgTemp.Append(d.AvgTemp)
		bldrMinTemp.Append(d.MinTemp)
		bldrMaxTemp.Append(d.MaxTemp)
		bldrAvgHumidity.Append(d.AvgHumidity)
		bldrAvgWindSpeed.Append(d.AvgWindSpeed)
		bldrSampleCount.Append(int64(d.SampleCount))
		bldrCondition.Append(d.MostCommonCond)
	}

	// Создаём RecordBatch
	batch := arrow.NewRecordBatch(
		schema,
		[]arrow.Array{
			bldrCity.NewArray(),
			bldrWindowStart.NewArray(),
			bldrWindowEnd.NewArray(),
			bldrAvgTemp.NewArray(),
			bldrMinTemp.NewArray(),
			bldrMaxTemp.NewArray(),
			bldrAvgHumidity.NewArray(),
			bldrAvgWindSpeed.NewArray(),
			bldrSampleCount.NewArray(),
			bldrCondition.NewArray(),
		},
	)

	return batch, nil
}

// ServeArrowFile отдаёт Arrow файл через HTTP
func ServeArrowFile() {
	data, err := ReadAggregatedData()
	if err != nil {
		log.Printf("Error reading aggregated data: %v", err)
		return
	}

	if len(data) == 0 {
		log.Println("No aggregated data available yet")
		return
	}

	batch, err := ConvertToArrow(data)
	if err != nil {
		log.Printf("Error converting to Arrow: %v", err)
		return
	}
	defer batch.Release()

	// Сохраняем в Arrow файл
	f, err := os.Create("data/weather.arrow")
	if err != nil {
		log.Printf("Error creating arrow file: %v", err)
		return
	}
	defer f.Close()

	writer := ipc.NewWriter(f, ipc.WithSchema(batch.Schema()))
	defer writer.Close()

	if err := writer.Write(batch); err != nil {
		log.Printf("Error writing arrow batch: %v", err)
		return
	}

	log.Printf("Arrow file saved: data/weather.arrow (%d records)", len(data))
}

// StartArrowServer запускает gRPC сервер для передачи Arrow данных
func StartArrowServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Printf("Failed to listen: %v", err)
		return
	}
	
	grpcServer := grpc.NewServer()
	log.Println("Arrow gRPC server listening on :50051")
	
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			ServeArrowFile()
		}
	}()
	
	if err := grpcServer.Serve(lis); err != nil {
		log.Printf("gRPC server error: %v", err)
	}
}

func init() {
	// Запускаем Arrow сервер в фоне
	go StartArrowServer()
}