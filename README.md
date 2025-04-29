# Балансировщик нагрузки

## Описание

Балансировщик нагрузки, работающий на основе алгоритма round robin и использующий ограничитель частоты запросов.  

## Установка и запуск

Для запуска проекта вам потребуется [Docker](https://www.docker.com/get-started) и [Docker Compose](https://docs.docker.com/compose/install/).

### 1. Клонирование репозитория

```bash
git clone https://github.com/emilien-gus/load-balancer
cd load-balancer
```

### 2. Запуск проекта

```bash
docker-compose up --build
```
