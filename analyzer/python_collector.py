import asyncio
import aiohttp
import json
import time
from datetime import datetime

API_KEY = "test_key"  # Замени на реальный
CITIES = ["Moscow", "Saint Petersburg", "Novosibirsk"]

async def fetch_weather(session, city):
    url = f"https://api.openweathermap.org/data/2.5/weather?q={city}&appid={API_KEY}&units=metric"
    start = time.time()
    try:
        async with session.get(url) as resp:
            data = await resp.json()
            elapsed = time.time() - start
            return {
                "city": city,
                "temperature": data["main"]["temp"],
                "elapsed": elapsed
            }
    except:
        return None

async def collect_all():
    async with aiohttp.ClientSession() as session:
        tasks = [fetch_weather(session, city) for city in CITIES]
        return await asyncio.gather(*tasks)

async def benchmark(iterations=10):
    times = []
    for i in range(iterations):
        start = time.time()
        results = await collect_all()
        elapsed = time.time() - start
        times.append(elapsed)
        print(f"Iteration {i+1}: {elapsed:.2f}s")
    
    avg = sum(times) / len(times)
    print(f"\nAverage time: {avg:.2f}s")
    return avg

if __name__ == "__main__":
    asyncio.run(benchmark())