package main

import (
    "context"
    "fmt"
    "time"
    
    tb "github.com/go-telegram/bot"
)

func startScheduler(bot *tb.Bot, chatID int64) {
    go func() {
        for {
            now := time.Now().Local()
            nextMinute := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
            time.Sleep(time.Until(nextMinute))

            currentTime := now.Format("15:04")
            currentDay := getDayRussian(now.Weekday())

            // ПРОВЕРКА ЛЕКАРСТВ (каждую минуту)
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
                }
            }
        }
    }()

    // ЕЖЕДНЕВНЫЙ ПЛАН В 08:00
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
        }
    }()

    // НАПОМИНАНИЯ О ВИЗИТАХ (каждый день в 09:00)
    go func() {
        for {
            now := time.Now().Local()
            next9am := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
            if now.After(next9am) {
                next9am = next9am.Add(24 * time.Hour)
            }
            time.Sleep(time.Until(next9am))
            
            tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
            
            for _, v := range storage.VetVisits {
                if v.Date == tomorrow {
                    msg := fmt.Sprintf("🏥 НАПОМИНАНИЕ! Завтра визит к ветеринару:\n📅 %s в %s\n📝 %s\n📍 %s",
                        v.Date, v.Time, v.Description, v.Address)
                    bot.SendMessage(context.Background(), &tb.SendMessageParams{
                        ChatID: chatID,
                        Text:   msg,
                    })
                }
            }
        }
    }()
}
