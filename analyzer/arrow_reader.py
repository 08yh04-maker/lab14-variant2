import pyarrow as pa
import pyarrow.ipc as ipc
import pandas as pd
import plotly.express as px
from pathlib import Path

def read_arrow_file(file_path: str):
    """Читает Arrow файл и возвращает DataFrame"""
    if not Path(file_path).exists():
        print(f"File {file_path} not found")
        return None
    
    with open(file_path, 'rb') as f:
        reader = ipc.open_file(f)
        table = reader.read_all()
        df = table.to_pandas()
    
    return df

def read_arrow_from_grpc():
    """Читает Arrow данные через gRPC (упрощённо — читаем файл)"""
    # В реальном проекте здесь был бы gRPC клиент
    return read_arrow_file('data/weather.arrow')

def analyze_arrow_data(df):
    """Анализирует данные из Arrow"""
    if df is None or df.empty:
        print("No data to analyze")
        return
    
    print("=== Arrow Data Analysis ===")
    print(f"Records: {len(df)}")
    print(f"Columns: {list(df.columns)}")
    print("\nFirst 5 rows:")
    print(df.head())
    
    # Базовая статистика
    print("\n=== Temperature Statistics ===")
    print(f"Avg temperature: {df['avg_temperature'].mean():.2f}°C")
    print(f"Min temperature: {df['min_temperature'].min():.2f}°C")
    print(f"Max temperature: {df['max_temperature'].max():.2f}°C")
    
    # По городам
    city_stats = df.groupby('city').agg({
        'avg_temperature': 'mean',
        'sample_count': 'sum'
    }).round(2)
    print("\n=== City Statistics ===")
    print(city_stats)
    
    return df

def plot_arrow_data(df):
    """Визуализирует данные из Arrow"""
    if df is None or df.empty:
        return
    
    # График температур по городам
    fig1 = px.bar(df, x='city', y='avg_temperature', 
                  title='Average Temperature by City',
                  labels={'avg_temperature': 'Temperature (°C)', 'city': 'City'})
    fig1.write_html('data/temperature_chart.html')
    
    # Временной ряд
    df['window_start_dt'] = pd.to_datetime(df['window_start'])
    fig2 = px.line(df, x='window_start_dt', y='avg_temperature', color='city',
                   title='Temperature Trends Over Time',
                   labels={'avg_temperature': 'Temperature (°C)', 'window_start_dt': 'Time'})
    fig2.write_html('data/temperature_trend.html')
    
    print("Charts saved to data/ directory")

if __name__ == "__main__":
    # Читаем Arrow файл
    df = read_arrow_file('data/weather.arrow')
    
    if df is not None:
        analyze_arrow_data(df)
        plot_arrow_data(df)