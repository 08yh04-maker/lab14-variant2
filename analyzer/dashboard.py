import streamlit as st
import polars as pl
import plotly.express as px
import plotly.graph_objects as go
import pandas as pd
from pathlib import Path
import time

st.set_page_config(
    page_title="Weather Data Dashboard",
    page_icon="🌤️",
    layout="wide"
)

st.title("🌤️ Weather Data Pipeline Dashboard")
st.markdown("### Real-time weather monitoring and analytics")

# Функция загрузки данных
@st.cache_data(ttl=30)
def load_aggregated_data():
    json_file = Path("data/aggregated.json")
    if not json_file.exists():
        return None
    
    data = []
    with open(json_file, 'r') as f:
        for line in f:
            if line.strip():
                data.append(eval(line))
    
    if not data:
        return None
    
    df = pl.DataFrame(data)
    return df

@st.cache_data(ttl=30)
def load_arrow_data():
    arrow_file = Path("data/weather.arrow")
    if not arrow_file.exists():
        return None
    
    import pyarrow as pa
    import pyarrow.ipc as ipc
    
    with open(arrow_file, 'rb') as f:
        reader = ipc.open_file(f)
        table = reader.read_all()
        df = table.to_pandas()
    return pl.from_pandas(df)

# Sidebar
st.sidebar.header("Filters")
auto_refresh = st.sidebar.checkbox("Auto refresh (30 sec)", value=True)

# Основная панель
col1, col2, col3, col4 = st.columns(4)

# Загрузка данных
df = load_aggregated_data()

if df is None:
    st.warning("No data available yet. Waiting for collector...")
    st.info("Make sure the Go collector is running and collecting weather data.")
    
    # Автообновление
    if auto_refresh:
        time.sleep(30)
        st.rerun()
    st.stop()

# Метрики
total_cities = df["city"].n_unique()
total_samples = df["sample_count"].sum()
avg_temp = df["avg_temperature"].mean()

col1.metric("🏙️ Cities Monitored", total_cities)
col2.metric("📊 Total Samples", f"{total_samples:,}")
col3.metric("🌡️ Avg Temperature", f"{avg_temp:.1f}°C")
col4.metric("🕐 Last Update", "Just now")

st.divider()

# Графики
tab1, tab2, tab3, tab4 = st.tabs(["📈 Temperature Trends", "📊 City Comparison", "🌡️ Temperature Distribution", "📋 Raw Data"])

with tab1:
    st.subheader("Temperature Trends Over Time")
    
    # Преобразуем для Plotly
    plot_df = df.to_pandas()
    plot_df['window_start'] = pd.to_datetime(plot_df['window_start'])
    
    fig = px.line(plot_df, x='window_start', y='avg_temperature', color='city',
                  title='Average Temperature by City',
                  labels={'avg_temperature': 'Temperature (°C)', 'window_start': 'Time'})
    fig.update_layout(height=500)
    st.plotly_chart(fig, use_container_width=True)

with tab2:
    st.subheader("City Statistics Comparison")
    
    city_stats = df.group_by("city").agg([
        pl.col("avg_temperature").mean().alias("avg_temp"),
        pl.col("avg_humidity").mean().alias("avg_humidity"),
        pl.col("avg_wind_speed").mean().alias("avg_wind"),
        pl.col("sample_count").sum().alias("total_samples")
    ]).sort("avg_temp", descending=True)
    
    city_pd = city_stats.to_pandas()
    
    fig2 = px.bar(city_pd, x='city', y='avg_temp', 
                  title='Average Temperature by City',
                  color='avg_temp', color_continuous_scale='RdYlGn_r')
    fig2.update_layout(height=400)
    st.plotly_chart(fig2, use_container_width=True)
    
    col1, col2 = st.columns(2)
    
    with col1:
        fig3 = px.bar(city_pd, x='city', y='avg_humidity',
                      title='Average Humidity by City',
                      color='avg_humidity', color_continuous_scale='Blues')
        st.plotly_chart(fig3, use_container_width=True)
    
    with col2:
        fig4 = px.bar(city_pd, x='city', y='avg_wind',
                      title='Average Wind Speed by City',
                      color='avg_wind', color_continuous_scale='Greens')
        st.plotly_chart(fig4, use_container_width=True)

with tab3:
    st.subheader("Temperature Distribution")
    
    fig5 = px.histogram(plot_df, x='avg_temperature', nbins=20,
                        title='Distribution of Average Temperatures',
                        labels={'avg_temperature': 'Temperature (°C)'})
    fig5.update_layout(height=400)
    st.plotly_chart(fig5, use_container_width=True)
    
    # Статистика
    st.subheader("Statistics Summary")
    stats_df = df.select([
        pl.col("avg_temperature").min().alias("Min Temperature"),
        pl.col("avg_temperature").max().alias("Max Temperature"),
        pl.col("avg_temperature").mean().alias("Mean Temperature"),
        pl.col("avg_temperature").std().alias("Std Temperature")
    ]).round(2)
    st.dataframe(stats_df.to_pandas(), use_container_width=True)

with tab4:
    st.subheader("Raw Aggregated Data")
    st.dataframe(df.to_pandas(), use_container_width=True)

# Arrow Data Analysis
st.divider()
st.subheader("📊 Apache Arrow Data Analysis")

arrow_df = load_arrow_data()
if arrow_df is not None:
    arrow_pd = arrow_df.to_pandas()
    st.write(f"Arrow Data Shape: {arrow_pd.shape}")
    st.dataframe(arrow_pd.head(10), use_container_width=True)
else:
    st.info("Arrow data not available yet. Waiting for collector to generate weather.arrow file")

# Footer
st.divider()
st.caption("Data Pipeline: Go Collector → JSON → Aggregation → Arrow → Streamlit Dashboard")

# Автообновление
if auto_refresh:
    time.sleep(30)
    st.rerun()