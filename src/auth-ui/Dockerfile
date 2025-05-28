FROM python:3.10-slim

WORKDIR /app

# Установка зависимостей
RUN pip install --no-cache-dir flask kubernetes

COPY . .

CMD ["python", "app.py"]
