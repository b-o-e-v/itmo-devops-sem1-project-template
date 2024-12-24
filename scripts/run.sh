#!/bin/bash

set -e

GO_APP_PATH="./main.go"

echo "Компилируем приложение..."
go build -o app $GO_APP_PATH

echo "Запускаем приложение..."
nohup ./app > output.log 2>&1 &

GO_APP_PID=$!

echo "Приложение запущено с PID: $GO_APP_PID"
