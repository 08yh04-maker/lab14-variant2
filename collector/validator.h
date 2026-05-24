#ifndef VALIDATOR_H
#define VALIDATOR_H

#include <stdbool.h>

typedef struct {
    bool is_valid;
    char* error_message;
} ValidationResult;

extern ValidationResult* validate_weather_json(const char* json_ptr);
extern void free_error_message(char* ptr);
extern bool validate_temperature(double temp);
extern bool validate_humidity(int humidity);
extern bool validate_wind_speed(double speed);
extern bool validate_pressure(int pressure);

#endif