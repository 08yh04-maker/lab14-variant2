# Лабораторная работа №14

**Студент:** Харлашкин Юрий Дмитриевич
**Группа:** 221131
**Вариант:** 2 (Анализ погодных данных)
**Сложность:** повышенная
**Дата:** 24.05.2026

## Название работы
Конвейеры обработки данных: сбор на Go, анализ на Python

## Описание проекта
ETL-конвейер для сбора и анализа погодных данных:
- **Go-сборщик**: параллельный опрос OpenWeatherMap API, буферизация, graceful shutdown
- **Python-анализ**: Polars, DuckDB, Parquet, визуализация Plotly
- **Повышенная сложность**: оконная агрегация, Apache Arrow, веб-дашборд

## Технологии
- **Go**: горутины, каналы, JSON, net/http, graceful shutdown
- **Python**: Polars, DuckDB, PyArrow, Plotly, Streamlit
- **Форматы**: JSON, Parquet, Arrow

## Структура проекта
lab14-variant2/
├── .github/workflows/
├── collector/ # Go-сборщик
│ ├── main.go
│ ├── go.mod
│ └── go.sum
├── analyzer/ # Python-анализ
│ ├── main.py
│ ├── visualize.py
│ └── dashboard.py
├── data/ # Данные (игнорируется)
├── docker-compose.yml
├── README.md
├── PROMPT_LOG.md
└── .gitignore
## Запуск

### Go-сборщик
```bash
cd collector
go run main.go
Python-анализ
pip install -r requirements.txt
python analyzer/main.py
Веб-дашборд
streamlit run analyzer/dashboard.py
API ключ
Для работы требуется API ключ OpenWeatherMap. Получить на https://openweathermap.org/api