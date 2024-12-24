# Финальный проект 1 семестра (уровень сложности - простой)

REST API сервис для загрузки и выгрузки данных о ценах

## Требования к системе

- **Операционная система**: Linux (рекомендуется Ubuntu 20.04 или выше)
- **Аппаратные требования**: минимум 2 ГБ ОЗУ и 2 ГБ свободного места на диске
- **СУБД**: PostgreSQL, порт 5432

## Установка и запуск

1. Установите PostgreSQL и настройте базу данных:

```bash
sudo apt update
sudo apt install postgresql
sudo -u <user> psql -c 'CREATE DATABASE "project-sem-1";'
sudo -u <user> psql -c "CREATE USER validator WITH PASSWORD 'val1dat0r';"
sudo -u <user> psql -c 'GRANT ALL PRIVILEGES ON DATABASE "project-sem-1" TO validator;'
sudo -u <user> psql -d project-sem-1 -c "GRANT CREATE ON SCHEMA public TO validator;"
```

2. Установите Go (рекомендуется версия 1.20 или выше)

3. Склонируйте репозиторий проекта:

```bash
git clone git@github.com:b-o-e-v/itmo-devops-sem1-project-template.git
cd itmo-devops-sem1-project-template
```

4. Запустите скрипт подготовки:

```bash
sh ./scripts/prepare.sh
```

5. Запустите сервер локально:

```bash
sh ./scripts/run.sh
```

## Тестирование

1. Запустите тесты API-запросов с помощью скрипта:

```bash
sh ./scripts/tests.sh 1
```

2. Скрипт `tests.sh` проверяет:

- Корректность добавления данных через `POST /api/v0/prices`
- Корректность выгрузки данных через `GET /api/v0/prices`

### Пример успешного выполнения тестов:

- **POST запрос** загружает данные из архива и возвращает JSON с метриками
- **GET запрос** возвращает zip-архив с корректным содержимым файла `data.csv`

## Скрипты

### `prepare.sh`

- Устанавливает зависимости приложения
- Настраивает базу данных

```bash
sh ./scripts/prepare.sh
```

### `run.sh`

- Запускает приложение локально

Пример использования:

```bash
sh ./scripts/run.sh
```

### `tests.sh`

- Выполняет тесты API-запросов.

Пример использования:

```bash
sh ./scripts/tests.sh 1
```

## API эндпоинты

### `POST /api/v0/prices`

**Входные данные:**

- Тело запроса: архив с данными

**Логика работы:**

- Разархивация архива
- Построчная запись данных в базу данных
- Возврат JSON-объекта:

```json
{
  "total_items": 100,        // общее количество добавленных элементов
  "total_categories": 15,    // общее количество категорий
  "total_price": 100000      // суммарная стоимость всех объектов
}
```

**Пример запроса:**

```bash
curl -s -F "file=@test_data.zip" http://localhost:8080/api/v0/prices
```

### `GET /api/v0/prices`

**Входные данные:**

- Отсутствуют

**Логика работы:**

- Извлечение всех записей из базы данных
- Формирование zip-архива с файлом `data.csv`
- Возврат zip-архива

**Пример запроса:**

```bash
curl -o ./sample_data/prices.zip http://localhost:8080/api/v0/prices
```

## Контакт

В случае вопросов можно обращаться:

- **Telegram**: @boevvv
