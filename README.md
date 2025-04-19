# TWB — Telegram Weather Bot 🌦

## 🤖 Бота можно протестировать в телеграмме: @bumus_bot

## ⚠️ Бот может выдавать иногда неверные данные, так как используется бесплатное API. Прошу прощения за предоставленные неудобства

## 🌇 Телеграм-бот, написанный на Go, который умеет:

- отвечать на команды
- запрашивать погоду по названию города / страна, город (более точный поиск)
- сохранять избранные локации
- вести логирование активности по дням
- и молча уходить в `AFK`, если ты его игнорируешь

> 🛠 Архитектура построена с акцентом на модульность и устойчивость.

## 💫 В чем идея проекта?

Данный проект - это моя первая работа. Главная цель -  отточить общие знания в программировании, а также свои знания в Go. Здесь сделан уклон на асинхроннное программирование, дабы в будущих проектах не было проблем с пониманием этого аспекта. Телеграмм бота можно с легкостью дальше улучшать, все зависит от вашего желания 😎.

## 🚀 Быстрый старт

### 📦 Требования

- Go 1.20+
- PostgreSQL (база данных должна быть запущена локально или удалённо)
- API Token из WeatherApi
- Telegram Bot API Token
- Make (опционально, если используешь `makefile`)

---

### 🔧 Установка

```bash
git clone https://github.com/Ummuys/TWB.git
cd TG_W
go mod tidy
make or go run main.go
```

### 🤫.env

```
BOT_API=your_tg_bot_token
CONNECT=postgres://user:pass@your_host:your_port/your_database?sslmode=disable # +
WEATHER_API=your_token_from_https://www.weatherapi.com/
```

## 🧩 Структура проекта

```
TG_W/
├── bot/
│   └── bot_commands.go├── database/
│   └── pg_commands.go
│
├── logs/ This folder might not exist. You can either create it manually, or use make to handle it for you.
│   
├── weather/
│   ├── example.json
│   └── weather.go
│
├── .env / This file not exists. Create him!
├── .gitignore
├── go.mod
├── go.sum
├── main.go
├── makefile
└── README.md
```

## ⚙️ Основной функционал

* 📬 `/start` и `/menu` команды
* 🔄 Состояния пользователя (`CHS`, `WEA`, `END`)
* 💾 Сохранение до 3 любимых городов
* 💤 Автоматический выход в AFK через 1 минуту неактивности
* 🔍 Получение прогноза погоды по названию города
* 📓 Логирование в файлы `logs/YYYY-MM-DD.log`
* 📦 Безопасно выключает бота командой `exit`

## ⚡️Автор: Евгений Егоров/Ummuys
