#!/bin/bash

# Завершаем выполнение скрипта при ошибке любой команды
set -e

# Загружаем переменные окружения из файла .env на уровень выше
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
else
    echo "Файл .env не найден на уровень выше. Пожалуйста, создайте его с необходимыми переменными окружения."
    exit 1
fi

# Проверяем, что все необходимые переменные окружения заданы
: "${POSTGRES_HOST:?Переменная POSTGRES_HOST не задана}"
: "${POSTGRES_PORT:?Переменная POSTGRES_PORT не задана}"
: "${POSTGRES_USER:?Переменная POSTGRES_USER не задана}"
: "${POSTGRES_PASSWORD:?Переменная POSTGRES_PASSWORD не задана}"
: "${POSTGRES_DB:?Переменная POSTGRES_DB не задана}"

# Устанавливаем зависимости Go
echo "Устанавливаем зависимости Go..."
go mod tidy
echo "Зависимости установлены"

# Проверяем, установлен ли PostgreSQL
if ! command -v psql &> /dev/null
then
    echo "PostgreSQL не установлен. Установите его и попробуйте снова"
    exit 1
fi

# Выводим отладочную информацию
echo "Подключение к PostgreSQL: HOST=$POSTGRES_HOST, PORT=$POSTGRES_PORT, USER=$POSTGRES_USER, DB=$POSTGRES_DB"
# Подключаемся к базе данных $POSTGRES_DB и создаём таблицу prices, если её нет
echo "Создание таблицы prices в базе данных $POSTGRES_DB..."
PGPASSWORD=$POSTGRES_PASSWORD psql -U $POSTGRES_USER -h $POSTGRES_HOST -p $POSTGRES_PORT -d $POSTGRES_DB -c "
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,           -- Автоматически увеличиваемый идентификатор
    created_at DATE NOT NULL,        -- Дата создания продукта
    name VARCHAR(255) NOT NULL,      -- Название продукта
    category VARCHAR(255) NOT NULL,  -- Категория продукта
    price DECIMAL(10, 2) NOT NULL    -- Цена продукта с точностью до 2 знаков после запятой
);"

echo "База данных подготовлена успешно"
