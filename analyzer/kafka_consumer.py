from kafka import KafkaConsumer
import json
from collections import deque
from datetime import datetime

window_size = 5  # секунд
window = deque()

consumer = KafkaConsumer(
    'weather-data',
    bootstrap_servers=['localhost:9092'],
    auto_offset_reset='earliest',
    value_deserializer=lambda x: json.loads(x.decode())
)

print("Waiting for weather data...")

for msg in consumer:
    data = msg.value
    now = datetime.now()
    window.append((now, data))
    
    # Очищаем старые записи
    while window and (now - window[0][0]).total_seconds() > window_size:
        window.popleft()
    
    if window:
        temps = [w[1]["temperature"] for w in window]
        print(f"Sliding window ({len(window)} records): avg temp = {sum(temps)/len(temps):.2f}°C")