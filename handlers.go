package main

import (
    "context"
    "fmt"
    "strings"
    "time"
    
    tb "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
)

func registerHandlers(b *tb.Bot) {
    b.RegisterHandler(tb.HandlerTypeMessageText, "/start", tb.MatchTypeExact, startHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/daily", tb.MatchTypeExact, dailyHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/nextvet", tb.MatchTypeExact, nextVetHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/addvet", tb.MatchTypePrefix, addVetHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/delvet", tb.MatchTypePrefix, delVetHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/medlist", tb.MatchTypeExact, medListHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/week", tb.MatchTypeExact, weekHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/setweek", tb.MatchTypePrefix, setWeekHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/addanalysis", tb.MatchTypePrefix, addAnalysisHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/getanalysis", tb.MatchTypePrefix, getAnalysisHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/delanalysis", tb.MatchTypePrefix, delAnalysisHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/listanalysis", tb.MatchTypeExact, listAnalysisHandler)
}

func startHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    storage.NotifyChat = update.Message.Chat.ID
    storage.Load()
    
    msg := fmt.Sprintf(`🐱 Привет! Я бот для кота %s.

📋 КОМАНДЫ:

📅 Планирование:
/daily - план на сегодня
/nextvet - ближайший визит
/week - какая сейчас неделя
/setweek odd/even/auto - ручное переключение недель

🏥 Визиты:
/addvet ГГГГ-ММ-ДД ЧЧ:ММ Описание Адрес
/delvet ГГГГ-ММ-ДД

💊 Лекарства:
/medlist - список всех лекарств

📊 Анализы:
/addanalysis ГГГГ-ММ-ДД показатель=значение
/getanalysis ГГГГ-ММ-ДД
/delanalysis ГГГГ-ММ-ДД
/listanalysis - все даты с анализами

Текущая неделя: %s

Пример добавления анализов:
/addanalysis 2026-04-01 лейкоциты=8.2 глюкоза=5.1`, storage.Name, getWeekType())
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
    
    startScheduler(b, update.Message.Chat.ID)
}

func dailyHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    plan := getDailyPlan()
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   plan,
    })
}

func nextVetHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    today := time.Now().Format("2006-01-02")
    var nearest *VetVisit
    for _, v := range storage.VetVisits {
        if v.Date >= today {
            if nearest == nil || v.Date < nearest.Date {
                nearest = &v
            }
        }
    }
    
    var msg string
    if nearest == nil {
        msg = "Нет запланированных визитов."
    } else {
        msg = fmt.Sprintf("🏥 Ближайший визит:\n📅 %s в %s\n📝 %s\n📍 %s",
            nearest.Date, nearest.Time, nearest.Description, nearest.Address)
    }
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
}

func addVetHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    parts := strings.SplitN(update.Message.Text, " ", 5)
    if len(parts) < 5 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /addvet 2026-05-10 14:30 Осмотр ул.Ленина 5",
        })
        return
    }
    
    storage.VetVisits = append(storage.VetVisits, VetVisit{
        Date:        parts[1],
        Time:        parts[2],
        Description: parts[3],
        Address:     parts[4],
    })
    storage.Save()
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "✅ Визит добавлен",
    })
}

func delVetHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    parts := strings.Split(update.Message.Text, " ")
    if len(parts) != 2 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /delvet 2026-05-10",
        })
        return
    }
    
    date := parts[1]
    newVisits := []VetVisit{}
    for _, v := range storage.VetVisits {
        if v.Date != date {
            newVisits = append(newVisits, v)
        }
    }
    storage.VetVisits = newVisits
    storage.Save()
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "🗑 Визит(ы) удалён(ы)",
    })
}

func medListHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    oddMeds := []string{}
    evenMeds := []string{}
    allMeds := []string{}

    for _, m := range storage.Medicines {
        status := "✅"
        if !m.IsActive {
            status = "❌"
        }
        line := fmt.Sprintf("%s %s в %s - %s (%s)", status, m.Name, m.Time, m.Dosage, strings.Join(m.Days, ","))

        switch m.WeekPattern {
        case "odd":
            oddMeds = append(oddMeds, line)
        case "even":
            evenMeds = append(evenMeds, line)
        default:
            allMeds = append(allMeds, line)
        }
    }

    msg := "💊 ПОЛНЫЙ СПИСОК ЛЕКАРСТВ ДЛЯ ПАПУША:\n\n"
    msg += "📌 Каждый день:\n" + joinLines(allMeds) + "\n"
    if len(oddMeds) > 0 {
        msg += "\n📌 1 неделя (нечётная):\n" + joinLines(oddMeds) + "\n"
    }
    if len(evenMeds) > 0 {
        msg += "\n📌 2 неделя (чётная):\n" + joinLines(evenMeds)
    }

    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
}

func weekHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    _, weekNum := time.Now().ISOWeek()
    weekType := getWeekType()
    
    modeText := ""
    if storage.OverrideWeek == "odd" {
        modeText = "\n\n🔧 Режим: РУЧНОЙ (нечётная)"
    } else if storage.OverrideWeek == "even" {
        modeText = "\n\n🔧 Режим: РУЧНОЙ (чётная)"
    } else {
        modeText = "\n\n🤖 Режим: АВТОМАТИЧЕСКИЙ (по ISO)"
    }
    
    msg := fmt.Sprintf("📅 Сейчас %s неделя года (№%d)\n🏷 Для Папуша это %s%s",
        weekType, weekNum, weekType, modeText)
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
}

func setWeekHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    parts := strings.Split(update.Message.Text, " ")
    if len(parts) != 2 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /setweek odd  или  /setweek even  или  /setweek auto",
        })
        return
    }

    mode := parts[1]
    switch mode {
    case "odd":
        storage.OverrideWeek = "odd"
        storage.Save()
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "✅ Установлена РУЧНАЯ НЕЧЁТНАЯ (1) неделя\n\nИспользуйте /setweek auto для возврата к автоматическому режиму.",
        })
    case "even":
        storage.OverrideWeek = "even"
        storage.Save()
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "✅ Установлена РУЧНАЯ ЧЁТНАЯ (2) неделя\n\nИспользуйте /setweek auto для возврата к автоматическому режиму.",
        })
    case "auto":
        storage.OverrideWeek = ""
        storage.Save()
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "✅ Возврат к АВТОМАТИЧЕСКОМУ режиму\n\nБот будет определять чётность недели по ISO-стандарту.",
        })
    default:
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Неверный параметр. Используйте: /setweek odd, /setweek even или /setweek auto",
        })
    }
}

// Добавить анализы
func addAnalysisHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    // Формат: /addanalysis 2026-04-01 лейкоциты=8.2 глюкоза=5.1
    parts := strings.Split(update.Message.Text, " ")
    if len(parts) < 3 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /addanalysis ГГГГ-ММ-ДД показатель=значение показатель=значение ...\n\nПример: /addanalysis 2026-04-01 лейкоциты=8.2 глюкоза=5.1",
        })
        return
    }
    
    date := parts[1]
    
    // Проверяем формат даты
    _, err := time.Parse("2006-01-02", date)
    if err != nil {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Неверный формат даты. Используйте ГГГГ-ММ-ДД",
        })
        return
    }
    
    // Парсим показатели
    values := make(map[string]string)
    for i := 2; i < len(parts); i++ {
        kv := strings.SplitN(parts[i], "=", 2)
        if len(kv) == 2 {
            values[kv[0]] = kv[1]
        }
    }
    
    if len(values) == 0 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Не указаны показатели. Формат: показатель=значение",
        })
        return
    }
    
    // Проверяем, есть ли уже анализы за эту дату
    found := false
    for i, a := range storage.Analyses {
        if a.Date == date {
            // Обновляем существующие
            for k, v := range values {
                storage.Analyses[i].Values[k] = v
            }
            found = true
            break
        }
    }
    
    if !found {
        // Добавляем новые
        storage.Analyses = append(storage.Analyses, Analysis{
            Date:   date,
            Values: values,
        })
    }
    
    storage.Save()
    
    // Формируем сообщение
    msg := fmt.Sprintf("✅ Анализы за %s сохранены:\n\n", date)
    for k, v := range values {
        msg += fmt.Sprintf("📊 %s: %s\n", k, v)
    }
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
}

// Получить анализы за дату
func getAnalysisHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    parts := strings.Split(update.Message.Text, " ")
    if len(parts) != 2 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /getanalysis ГГГГ-ММ-ДД\n\nПример: /getanalysis 2026-04-01",
        })
        return
    }
    
    date := parts[1]
    
    // Ищем анализы
    for _, a := range storage.Analyses {
        if a.Date == date {
            msg := fmt.Sprintf("📊 Анализы Папуша за %s:\n\n", date)
            for k, v := range a.Values {
                msg += fmt.Sprintf("• %s: %s\n", k, v)
            }
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: update.Message.Chat.ID,
                Text:   msg,
            })
            return
        }
    }
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   fmt.Sprintf("❌ Нет анализов за %s", date),
    })
}

// Удалить анализы за дату
func delAnalysisHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    parts := strings.Split(update.Message.Text, " ")
    if len(parts) != 2 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /delanalysis ГГГГ-ММ-ДД\n\nПример: /delanalysis 2026-04-01",
        })
        return
    }
    
    date := parts[1]
    
    newAnalyses := []Analysis{}
    for _, a := range storage.Analyses {
        if a.Date != date {
            newAnalyses = append(newAnalyses, a)
        }
    }
    
    if len(newAnalyses) == len(storage.Analyses) {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   fmt.Sprintf("❌ Нет анализов за %s", date),
        })
        return
    }
    
    storage.Analyses = newAnalyses
    storage.Save()
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   fmt.Sprintf("🗑 Анализы за %s удалены", date),
    })
}

// Список всех дат с анализами
func listAnalysisHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    if len(storage.Analyses) == 0 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "📊 Нет сохранённых анализов",
        })
        return
    }
    
    msg := "📊 Список дат с анализами:\n\n"
    for _, a := range storage.Analyses {
        count := len(a.Values)
        msg += fmt.Sprintf("📅 %s (%d показателей)\n", a.Date, count)
    }
    msg += "\nЧтобы посмотреть анализы: /getanalysis ГГГГ-ММ-ДД"
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
}
