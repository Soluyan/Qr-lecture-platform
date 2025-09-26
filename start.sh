#!/bin/sh
# Сборка фронта
cd frontend
npm install
npm run build
cd ..

# Сборка Go и запуск
cd backend
go mod tidy
go build -o main .
./main