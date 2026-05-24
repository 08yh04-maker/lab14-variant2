use std::ffi::{CStr, CString};
use std::os::raw::{c_char, c_double, c_int};
use serde_json::Value;

// Структура для результата валидации
#[repr(C)]
pub struct ValidationResult {
    pub is_valid: bool,
    pub error_message: *mut c_char,
}

// Освобождает память строки ошибки
#[no_mangle]
pub extern "C" fn free_error_message(ptr: *mut c_char) {
    if !ptr.is_null() {
        unsafe {
            let _ = CString::from_raw(ptr);
        }
    }
}

// Валидация температуры
#[no_mangle]
pub extern "C" fn validate_temperature(temp: c_double) -> bool {
    temp >= -90.0 && temp <= 60.0
}

// Валидация влажности
#[no_mangle]
pub extern "C" fn validate_humidity(humidity: c_int) -> bool {
    humidity >= 0 && humidity <= 100
}

// Валидация скорости ветра
#[no_mangle]
pub extern "C" fn validate_wind_speed(speed: c_double) -> bool {
    speed >= 0.0 && speed <= 150.0
}

// Валидация давления
#[no_mangle]
pub extern "C" fn validate_pressure(pressure: c_int) -> bool {
    pressure >= 800 && pressure <= 1100
}

// Полная валидация погодных данных из JSON строки
#[no_mangle]
pub extern "C" fn validate_weather_json(json_ptr: *const c_char) -> *mut ValidationResult {
    let mut result = Box::new(ValidationResult {
        is_valid: true,
        error_message: std::ptr::null_mut(),
    });

    if json_ptr.is_null() {
        return Box::into_raw(result);
    }

    let json_str = unsafe {
        match CStr::from_ptr(json_ptr).to_str() {
            Ok(s) => s,
            Err(_) => return Box::into_raw(result),
        }
    };

    let parsed: Result<Value, _> = serde_json::from_str(json_str);
    let data = match parsed {
        Ok(d) => d,
        Err(_) => return Box::into_raw(result),
    };

    let mut errors = Vec::new();

    // Проверка температуры
    if let Some(temp) = data.get("temperature").and_then(|v| v.as_f64()) {
        if !validate_temperature(temp) {
            errors.push(format!("Invalid temperature: {}", temp));
        }
    }

    // Проверка влажности
    if let Some(humidity) = data.get("humidity").and_then(|v| v.as_i64()) {
        if !validate_humidity(humidity as i32) {
            errors.push(format!("Invalid humidity: {}", humidity));
        }
    }

    // Проверка скорости ветра
    if let Some(wind_speed) = data.get("wind_speed").and_then(|v| v.as_f64()) {
        if !validate_wind_speed(wind_speed) {
            errors.push(format!("Invalid wind speed: {}", wind_speed));
        }
    }

    // Проверка давления
    if let Some(pressure) = data.get("pressure").and_then(|v| v.as_i64()) {
        if !validate_pressure(pressure as i32) {
            errors.push(format!("Invalid pressure: {}", pressure));
        }
    }

    if !errors.is_empty() {
        let error_msg = errors.join("; ");
        let c_error_msg = CString::new(error_msg).unwrap();
        result.is_valid = false;
        result.error_message = c_error_msg.into_raw();
    }

    Box::into_raw(result)
}