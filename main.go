package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "fmt"

    tb "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
)

func main() {
    fmt.Println("1. Загрузка данных...")
    err := storage.Load()
    if err != nil {
        log.Fatal("Load error:", err)
    }
    fmt.Println("2. Данные загружены")

    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    fmt.Println("3. Проверка токена...")
    if BotToken == "" {
        log.Fatal("Токен не задан! Укажите токен в config.go")
    }
    if len(BotToken) < 10 {
        log.Fatal("Токен слишком короткий! Проверьте config.go")
    }
    fmt.Printf("4. Токен загружен (длина: %d символов)\n", len(BotToken))

    opts := []tb.Option{
        tb.WithDefaultHandler(func(ctx context.Context, b *tb.Bot, update *models.Update) {
            fmt.Println("Получено неизвестное сообщение")
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: update.Message.Chat.ID,
                Text:   "Неизвестная команда. Используйте /start",
            })
        }),
    }

    fmt.Println("5. Подключение к Telegram API...")
    bot, err := tb.New(BotToken, opts...)
    if err != nil {
        log.Fatal("Bot init error:", err)
    }
    fmt.Println("6. Бот успешно подключен!")

    registerHandlers(bot)

    log.Println("🐱 Бот для кота Папуш запущен!")
    log.Println("Текущая неделя:", getWeekType())
    fmt.Println("7. Бот готов к работе!")

    bot.Start(ctx)
}
