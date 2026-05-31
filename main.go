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
        tb.WithCallbackQueryDataHandler("", tb.MatchTypePrefix, callbackHandler),
    }

    bot, err := tb.New(BotToken, opts...)
    if err != nil {
        log.Fatal("Bot init error:", err)
    }

    registerHandlers(bot)

    log.Println("🐱 Бот для кота Папуш запущен!")
    log.Println("Текущая неделя:", getWeekType())

    if storage.NotifyChat != 0 {
        log.Printf("Запуск планировщика для chatID=%d", storage.NotifyChat)
        startScheduler(bot, storage.NotifyChat)
    } else {
        log.Println("ChatID не найден. Планировщик запустится после команды /start")
    }

    bot.Start(ctx)
}
