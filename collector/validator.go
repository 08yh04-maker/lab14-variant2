package main

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lweather_validator
#include <stdlib.h>
#include "validator.h"
*/
import "C"

import (
	"encoding/json"
	"log"
	"unsafe"
)

// ValidationResult структура результата валидации
type ValidationResult struct {
	IsValid      bool
	ErrorMessage string
}

// ValidateWeatherData валидирует WeatherData через Rust-библиотеку
func ValidateWeatherData(data WeatherData) ValidationResult {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ValidationResult{IsValid: false, ErrorMessage: err.Error()}
	}

	jsonStr := string(jsonBytes)
	cJSON := C.CString(jsonStr)
	defer C.free(unsafe.Pointer(cJSON))

	result := C.validate_weather_json(cJSON)
	defer C.free_error_message(result.error_message)

	if result.is_valid {
		return ValidationResult{IsValid: true}
	}

	errMsg := C.GoString(result.error_message)
	return ValidationResult{IsValid: false, ErrorMessage: errMsg}
}