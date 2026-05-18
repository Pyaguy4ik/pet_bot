package main

import (
    "context"
    "fmt"
    "sync"
    "time"
    tb "github.com/go-telegram/bot"
)

var (
    schedulerOnce sync.Once
    schedulerRunning = false
)

func startScheduler(bot *tb.Bot, chatID int64) {
    schedulerOnce.Do(func() {
        schedulerRunning = true
        fmt.Println("🕐 Планировщик запущен (один раз)")
        
        go func() {
            // Проверка каждую минуту
            for {
                now := time.Now().Local()
                // Ждём до следующей минуты
                nextMinute := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
                time.Sleep(time.Until(nextMinute))
                
                currentTime := now.Format("15:04")
                currentDay := getDayRussian(now.Weekday())

                for _, med := range storage.Medicines {
                    if !med.IsActive {
                        continue
                    }

                    dayMatch := false
                    for _, d := range med.Days {
                        if d == currentDay {
                            dayMatch = true
                            break
                        }
                    }
                    if !dayMatch {
                        continue
                    }

                    if !isWeekMatch(med.WeekPattern) {
                        continue
                    }

                    if med.Time == currentTime {
                        msg := fmt.Sprintf("💊 Напоминание для кота %s!\n📋 %s в %s - %s\n📅 Неделя: %s",
                            storage.Name, med.Name, med.Time, med.Dosage, getWeekType())
                        bot.SendMessage(context.Background(), &tb.SendMessageParams{
                            ChatID: chatID,
                            Text:   msg,
                        })
                        fmt.Printf("⏰ Отправлено напоминание: %s в %s\n", med.Name, med.Time)
                    }
                }
            }
        }()
        
        // Ежедневный план в 08:00
        go func() {
            for {
                now := time.Now().Local()
                next8am := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
                if now.After(next8am) {
                    next8am = next8am.Add(24 * time.Hour)
                }
                time.Sleep(time.Until(next8am))
                
                plan := getDailyPlan()
                bot.SendMessage(context.Background(), &tb.SendMessageParams{
                    ChatID: chatID,
                    Text:   plan,
                })
                fmt.Println("📅 Отправлен ежедневный план")
            }
        }()
    })
}
