
# Yadro Test Task

Это CLI-приложение для обработки событий биатлона. 
Оно читает входные данные из файлов конфигурации и событий, формирует отчёт и сохраняет лог.

## 🔧 Сборка

Для сборки проекта используйте стандартную команду Go:

```bash
go build -o biathlon-simulator ./cmd
```

## ▶️ Запуск

```bash
./biathlon-simulator -config config.json -events events.txt
```

### Доступные флаги:

| Флаг       | Описание                                      | Обязательный | По умолчанию   |
|------------|-----------------------------------------------|--------------|----------------|
| `-config`  | Путь к JSON-файлу конфигурации                | ✅           | —              |
| `-events`  | Путь к файлу с событиями                      | ✅           | —              |
| `-output`  | Путь к лог-файлу                              | ❌           | `output.log`   |
| `-report`  | Путь к файлу, в который будет записан отчёт   | ❌           | `report.txt`   |

Пример:

```bash
./biathlon-simulator -config ./data/config.json -events ./data/events.txt -output log.txt -report result.txt
```

## 📂 Структура проекта

```
.
├── cmd/                  # main.go — точка входа
├── internal/
│   ├── config/           # Загрузка и обработка конфигурации
│   ├── event/            # Основная логика обработки событий
│   └── flags/            # Парсинг CLI флагов
├── testdata/             # Тестовые данные (пример config и events)
├── go.mod
├── go.sum
└── README.md
```

## 🧪 Тесты

Запуск всех тестов:

```bash
go test ./...
```

