# ForgeHost

## Требования

- Docker и Docker Compose
- Go
- `make`
- [golangci-lint](https://golangci-lint.run)

## Docker-зависимости

Loki (плагин для логов):

```sh
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
```

Образы:

```sh
docker pull confluentinc/confluent-local:7.5.0
docker pull mongo:7
```

## Тесты

1. Перейти в корень проекта:

   ```sh
   cd forgehost
   ```

2. Запустить тесты всех сервисов:

   ```sh
   make test
   ```

## Линтер

1. Перейти в корень проекта:

   ```sh
   cd forgehost
   ```

2. Запустить линтер:

   ```sh
   make lint
   ```