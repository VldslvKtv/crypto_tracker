# Crypto Tracker

Проект предназначен для отслеживания цен криптовалют. Он позволяет добавлять криптовалюты в список отслеживаемых, удалять их из этого списка и получать текущую цену на указанный момент времени.

---

## Структура проекта

Проект организован с четким разделением ответственности, что упрощает поддержку и расширение функциональности.

### Описание папок и файлов

- **cmd/**:
  - `main.go`: Основной файл приложения, точка входа.

- **config/**:  
  - `config.go`: Настройки конфигурации проекта.

- **docs/**:  
  - `docs.go`: Сгенерированная документация API.  
  - `swagger.json`: Swagger-спецификация API.  
  - `swagger.yaml`: Swagger-спецификация API в формате YAML.

- **internal/**:  
  - **handlers/**: Обработчики HTTP-запросов.  
    - **add/**:  
      - `add_currency.go`: Обработчик для добавления криптовалюты в список отслеживаемых.  
    - **get/**:  
      - `get_currency.go`: Обработчик для получения цены криптовалюты.  
    - **remove/**:  
      - `remove_currency.go`: Обработчик для удаления криптовалюты из списка отслеживаемых.  

  - **models/**:  
    - `models.go`: Модели данных, используемые в проекте.  

  - **storage/pg/**:  
    - `pg.go`: Реализация хранения данных в PostgreSQL.  

  - **tracker/**:  
    - `tracker.go`: Логика отслеживания криптовалют.  

- **migrations/**:  
  - `001_create_table_coins.down.sql`: SQL-скрипт для отката миграции.  
  - `001_create_table_coins.up.sql`: SQL-скрипт для применения миграции.  

- **.env**: Файл переменных окружения.  
- **.env.example**: Пример файла переменных окружения.  
- **.gitignore**: Файл исключений для git.  
- **go.mod**: Файл с описанием зависимостей модуля.  
- **go.sum**: Файл с контрольными суммами зависимостей.  
- **README.md**: Описание проекта.  

---

## Инструкции по установке

1. Склонируйте репозиторий:

  git clone https://github.com/ваш-репозиторий/crypto-tracker.git

2. Создайте и настройте файл .env (пример в .env.example):

3. Отредактируйте .env, указав необходимые настройки (например, параметры подключения к базе данных и внешнему API).

4. Установите зависимости:

  go mod download

5. Запустите проект:

  go run cmd/main.go

## Инструкции по запуску с помощью Docker

1. Настроить .env (как в Инструкции по установке)

2. Собрать с запустить контейнеры:

  docker-compose up --build

3. Остановить контейнеры:

  docker-compose down