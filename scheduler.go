package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"
    tb "github.com/go-telegram/bot"
)

var (
    schedulerOnce   sync.Once
    schedulerRunning bool
)

func startScheduler(bot *tb.Bot, chatID int64) {
    schedulerOnce.Do(func() {
        schedulerRunning = true
        log.Println("🕐 Планировщик напоминаний запущен (один раз)")
        
        // Запускаем горутину для проверки каждую минуту
        go func() {
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("Паника в планировщике: %v", r)
                    // Можно перезапустить планировщик через некоторое время
                    time.Sleep(10 * time.Second)
                    go startScheduler(bot, chatID)
                }
            }()
            
            for {
                now := time.Now().Local()
                // Ждём до следующей минуты
                nextMinute := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
                time.Sleep(time.Until(nextMinute))
                
                currentTime := now.Format("15:04")
                currentDay := getDayRussian(now.Weekday())
                
                log.Printf("Проверка напоминаний: время %s, день %s", currentTime, currentDay)
                
                // Копируем список лекарств для безопасного обхода
                storage.mu.Lock()
                medicines := make([]Medicine, len(storage.Medicines))
                copy(medicines, storage.Medicines)
                storage.mu.Unlock()
                
                for _, med := range medicines {
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
                        
                        log.Printf("Отправка напоминания: %s в %s", med.Name, med.Time)
                        _, err := bot.SendMessage(context.Background(), &tb.SendMessageParams{
                            ChatID: chatID,
                            Text:   msg,
                        })
                        if err != nil {
                            log.Printf("Ошибка отправки сообщения: %v", err)
                        }
                    }
                }
            }
        }()
        
        // Запускаем горутину для ежедневного плана в 8:00
        go func() {
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("Паника в ежедневном планировщике: %v", r)
                    time.Sleep(10 * time.Second)
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
                }
            }()
            
            for {
                now := time.Now().Local()
                next8am := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
                if now.After(next8am) {
                    next8am = next8am.Add(24 * time.Hour)
                }
                time.Sleep(time.Until(next8am))
                
                plan := getDailyPlan()
                log.Println("📅 Отправка ежедневного плана")
                _, err := bot.SendMessage(context.Background(), &tb.SendMessageParams{
                    ChatID: chatID,
                    Text:   plan,
                })
                if err != nil {
                    log.Printf("Ошибка отправки ежедневного плана: %v", err)
                }
            }
        }()
    })
}
