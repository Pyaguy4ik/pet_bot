package main

import (
    "context"
    "log"
    "os"
    "os/signal"

    tb "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
)

func main() {
    err := storage.Load()
    if err != nil {
        log.Fatal("Load error:", err)
    }

    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    opts := []tb.Option{
        tb.WithDefaultHandler(func(ctx context.Context, b *tb.Bot, update *models.Update) {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: update.Message.Chat.ID,
                Text:   "Неизвестная команда. Используйте /start",
            })
        }),
    }

    bot, err := tb.New(BotToken, opts...)
    if err != nil {
        log.Fatal("Bot init error:", err)
    }

    registerHandlers(bot)

    log.Println("🐱 Бот для кота Папуш запущен!")
    log.Println("Текущая неделя:", getWeekType())
    
    // НЕ ЗАПУСКАЕМ ПЛАНИРОВЩИК ЗДЕСЬ!
    // Он запустится только при первом /start

    bot.Start(ctx)
}
