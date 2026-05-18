package main

import (
    "context"
    "fmt"
    "sort"
    "strings"
    "time"
    
    tb "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
)

var schedulerStarted = false

func registerHandlers(b *tb.Bot) {
    // Основные команды
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
    b.RegisterHandler(tb.HandlerTypeMessageText, "/addmed", tb.MatchTypePrefix, addMedHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/delmed", tb.MatchTypePrefix, delMedHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/editmed", tb.MatchTypePrefix, editMedHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/togglem ed", tb.MatchTypePrefix, toggleMedHandler)
    
    // Обработчики кнопок (текстовые сообщения)
    b.RegisterHandler(tb.HandlerTypeMessageText, "📅 План на сегодня", tb.MatchTypeExact, dailyHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "📊 Моя неделя", tb.MatchTypeExact, weekHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "💊 Все лекарства", tb.MatchTypeExact, medListHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "🏥 Ближайший визит", tb.MatchTypeExact, nextVetHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "❓ Помощь", tb.MatchTypeExact, helpHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "➕ Добавить визит", tb.MatchTypeExact, promptAddVetHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "🗑 Удалить визит", tb.MatchTypeExact, promptDelVetHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "💊 Добавить лекарство", tb.MatchTypeExact, promptAddMedHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "✏️ Редактировать лекарство", tb.MatchTypeExact, promptEditMedHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "📈 Анализы за дату", tb.MatchTypeExact, promptGetAnalysisHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "📋 Список анализов", tb.MatchTypeExact, listAnalysisHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "🔄 Переключить неделю", tb.MatchTypeExact, promptSetWeekHandler)
}

func startHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    storage.NotifyChat = update.Message.Chat.ID
    storage.Load()
    
    // Клавиатура с кнопками
    keyboard := &models.ReplyKeyboardMarkup{
        Keyboard: [][]models.KeyboardButton{
            {
                {Text: "📅 План на сегодня"},
                {Text: "📊 Моя неделя"},
            },
            {
                {Text: "💊 Все лекарства"},
                {Text: "🏥 Ближайший визит"},
            },
            {
                {Text: "➕ Добавить визит"},
                {Text: "🗑 Удалить визит"},
            },
            {
                {Text: "💊 Добавить лекарство"},
                {Text: "✏️ Редактировать лекарство"},
            },
            {
                {Text: "📈 Анализы за дату"},
                {Text: "📋 Список анализов"},
            },
            {
                {Text: "🔄 Переключить неделю"},
                {Text: "❓ Помощь"},
            },
        },
        ResizeKeyboard: true,
    }
    
    msg := fmt.Sprintf(`🐱 Привет! Я бот для кота %s.

👆 Используй кнопки внизу для управления.

📌 Быстрые команды (можно вводить вручную):
/daily - план на сегодня
/week - текущая неделя
/medlist - все лекарства
/nextvet - ближайший визит
/addvet YYYY-MM-DD HH:MM описание адрес
/delvet YYYY-MM-DD
/addmed Название|Время|Дозировка|дни|неделя
/delmed "название"
/editmed "название"|поле=значение
/addanalysis ГГГГ-ММ-ДД показатель=значение
/getanalysis ГГГГ-ММ-ДД
/listanalysis

Текущая неделя: %s`, storage.Name, getWeekType())
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID:      update.Message.Chat.ID,
        Text:        msg,
        ReplyMarkup: keyboard,
    })
    
    // Запускаем планировщик только один раз
    if !schedulerStarted {
        schedulerStarted = true
        startScheduler(b, update.Message.Chat.ID)
    }
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
    if len(storage.Medicines) == 0 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "💊 Список лекарств пуст. Добавьте через /addmed",
        })
        return
    }
    
    oddMeds := []string{}
    evenMeds := []string{}
    allMeds := []string{}

    for i, m := range storage.Medicines {
        status := "✅"
        if !m.IsActive {
            status = "❌"
        }
        line := fmt.Sprintf("%s #%d %s в %s - %s (дни: %s)", 
            status, i+1, m.Name, m.Time, m.Dosage, strings.Join(m.Days, ","))

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
    
    msg += "\n━━━━━━━━━━━━━━━━━━\n"
    msg += "Команды:\n"
    msg += "/delmed \"название\" - удалить\n"
    msg += "/editmed \"название\" - редактировать\n"
    msg += "/togglem ed \"название\" - вкл/выкл"

    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
}

// Добавление лекарства
func addMedHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    text := strings.TrimPrefix(update.Message.Text, "/addmed ")
    parts := strings.Split(text, "|")
    
    if len(parts) != 5 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   `❌ Неверный формат!

Правильный формат:
/addmed Название|Время(ЧЧ:ММ)|Дозировка|Дни(пн,вт,ср)|Неделя(all/odd/even)

Примеры:
/addmed Креон|10:55|1/2 капсулы|пн,вт,ср,чт,пт,сб,вс|all
/addmed Чистка зубов|22:00|обработка|пн,ср,пт|odd
/addmed Витамины|09:00|1 мл|сб,вс|even

Дни недели: пн, вт, ср, чт, пт, сб, вс
Неделя: all (каждую), odd (1 неделя), even (2 неделя)`,
        })
        return
    }
    
    name := strings.TrimSpace(parts[0])
    timeStr := strings.TrimSpace(parts[1])
    dosage := strings.TrimSpace(parts[2])
    daysStr := strings.TrimSpace(parts[3])
    weekPattern := strings.TrimSpace(parts[4])
    
    if _, err := time.Parse("15:04", timeStr); err != nil {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Неверный формат времени. Используйте ЧЧ:ММ (например, 14:30)",
        })
        return
    }
    
    days := strings.Split(daysStr, ",")
    validDays := map[string]bool{"пн": true, "вт": true, "ср": true, "чт": true, "пт": true, "сб": true, "вс": true}
    for _, d := range days {
        if !validDays[d] {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: update.Message.Chat.ID,
                Text:   "❌ Неверный формат дней. Используйте: пн,вт,ср,чт,пт,сб,вс",
            })
            return
        }
    }
    
    if weekPattern != "all" && weekPattern != "odd" && weekPattern != "even" {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Неверный формат недели. Используйте: all, odd, even",
        })
        return
    }
    
    storage.Medicines = append(storage.Medicines, Medicine{
        Name:        name,
        Time:        timeStr,
        Dosage:      dosage,
        Days:        days,
        WeekPattern: weekPattern,
        IsActive:    true,
    })
    storage.Save()
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   fmt.Sprintf("✅ Лекарство добавлено!\n\n📋 %s\n⏰ %s\n💊 %s\n📅 Дни: %s\n📆 Неделя: %s",
            name, timeStr, dosage, daysStr, weekPattern),
    })
}

// Удаление лекарства
func delMedHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    name := strings.TrimPrefix(update.Message.Text, "/delmed ")
    name = strings.TrimSpace(name)
    name = strings.Trim(name, "\"'")
    
    if name == "" {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /delmed \"Название лекарства\"\n\nПример: /delmed Креон",
        })
        return
    }
    
    found := false
    newMedicines := []Medicine{}
    for _, m := range storage.Medicines {
        if m.Name == name {
            found = true
            continue
        }
        newMedicines = append(newMedicines, m)
    }
    
    if !found {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   fmt.Sprintf("❌ Лекарство \"%s\" не найдено. Используйте /medlist для просмотра списка", name),
        })
        return
    }
    
    storage.Medicines = newMedicines
    storage.Save()
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   fmt.Sprintf("🗑 Лекарство \"%s\" удалено", name),
    })
}

// Редактирование лекарства
func editMedHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    text := strings.TrimPrefix(update.Message.Text, "/editmed ")
    parts := strings.Split(text, "|")
    
    if len(parts) < 2 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   `❌ Формат редактирования:

/editmed "старое название"|поле=значение

Доступные поля:
name - новое название
time - новое время (ЧЧ:ММ)
dosage - новая дозировка
days - новые дни (пн,вт,ср)
week - all/odd/even
active - true/false

Примеры:
/editmed Креон|time=14:00
/editmed Креон|dosage=1 капсула
/editmed Креон|days=пн,ср,пт
/editmed Креон|week=odd
/editmed Креон|active=false`,
        })
        return
    }
    
    oldName := strings.TrimSpace(parts[0])
    oldName = strings.Trim(oldName, "\"'")
    
    index := -1
    for i, m := range storage.Medicines {
        if m.Name == oldName {
            index = i
            break
        }
    }
    
    if index == -1 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   fmt.Sprintf("❌ Лекарство \"%s\" не найдено", oldName),
        })
        return
    }
    
    med := &storage.Medicines[index]
    changes := []string{}
    
    for i := 1; i < len(parts); i++ {
        kv := strings.SplitN(parts[i], "=", 2)
        if len(kv) != 2 {
            continue
        }
        
        field := strings.TrimSpace(kv[0])
        value := strings.TrimSpace(kv[1])
        
        switch field {
        case "name":
            changes = append(changes, fmt.Sprintf("имя: %s → %s", med.Name, value))
            med.Name = value
        case "time":
            if _, err := time.Parse("15:04", value); err == nil {
                changes = append(changes, fmt.Sprintf("время: %s → %s", med.Time, value))
                med.Time = value
            } else {
                b.SendMessage(ctx, &tb.SendMessageParams{
                    ChatID: update.Message.Chat.ID,
                    Text:   "❌ Неверный формат времени. Используйте ЧЧ:ММ",
                })
                return
            }
        case "dosage":
            changes = append(changes, fmt.Sprintf("дозировка: %s → %s", med.Dosage, value))
            med.Dosage = value
        case "days":
            days := strings.Split(value, ",")
            validDays := map[string]bool{"пн": true, "вт": true, "ср": true, "чт": true, "пт": true, "сб": true, "вс": true}
            valid := true
            for _, d := range days {
                if !validDays[d] {
                    valid = false
                    break
                }
            }
            if valid {
                changes = append(changes, fmt.Sprintf("дни: %s → %s", strings.Join(med.Days, ","), value))
                med.Days = days
            } else {
                b.SendMessage(ctx, &tb.SendMessageParams{
                    ChatID: update.Message.Chat.ID,
                    Text:   "❌ Неверный формат дней. Используйте: пн,вт,ср,чт,пт,сб,вс",
                })
                return
            }
        case "week":
            if value == "all" || value == "odd" || value == "even" {
                changes = append(changes, fmt.Sprintf("неделя: %s → %s", med.WeekPattern, value))
                med.WeekPattern = value
            } else {
                b.SendMessage(ctx, &tb.SendMessageParams{
                    ChatID: update.Message.Chat.ID,
                    Text:   "❌ Неверный формат недели. Используйте: all, odd, even",
                })
                return
            }
        case "active":
            if value == "true" || value == "false" {
                oldActive := "активно"
                if !med.IsActive {
                    oldActive = "неактивно"
                }
                newActive := "активно"
                if value == "false" {
                    newActive = "неактивно"
                }
                changes = append(changes, fmt.Sprintf("статус: %s → %s", oldActive, newActive))
                med.IsActive = value == "true"
            } else {
                b.SendMessage(ctx, &tb.SendMessageParams{
                    ChatID: update.Message.Chat.ID,
                    Text:   "❌ active может быть true или false",
                })
                return
            }
        }
    }
    
    storage.Save()
    
    if len(changes) == 0 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Не указаны изменения",
        })
        return
    }
    
    msg := fmt.Sprintf("✅ Лекарство \"%s\" обновлено:\n\n", oldName)
    for _, ch := range changes {
        msg += "• " + ch + "\n"
    }
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
}

// Включение/выключение лекарства
func toggleMedHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    name := strings.TrimPrefix(update.Message.Text, "/togglem ed ")
    name = strings.TrimSpace(name)
    name = strings.Trim(name, "\"'")
    
    if name == "" {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /togglem ed \"Название лекарства\"\n\nПример: /togglem ed Креон",
        })
        return
    }
    
    for i, m := range storage.Medicines {
        if m.Name == name {
            storage.Medicines[i].IsActive = !storage.Medicines[i].IsActive
            storage.Save()
            
            status := "включено ✅"
            if !storage.Medicines[i].IsActive {
                status = "выключено ❌"
            }
            
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: update.Message.Chat.ID,
                Text:   fmt.Sprintf("🔄 Лекарство \"%s\" %s", name, status),
            })
            return
        }
    }
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   fmt.Sprintf("❌ Лекарство \"%s\" не найдено", name),
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

// --- Анализы ---
func addAnalysisHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    parts := strings.Split(update.Message.Text, " ")
    if len(parts) < 3 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /addanalysis ГГГГ-ММ-ДД показатель=значение\n\nПример: /addanalysis 2026-04-01 лейкоциты=8.2 глюкоза=5.1",
        })
        return
    }
    
    date := parts[1]
    if _, err := time.Parse("2006-01-02", date); err != nil {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Неверный формат даты. Используйте ГГГГ-ММ-ДД",
        })
        return
    }
    
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
    
    found := false
    for i, a := range storage.Analyses {
        if a.Date == date {
            for k, v := range values {
                storage.Analyses[i].Values[k] = v
            }
            found = true
            break
        }
    }
    
    if !found {
        storage.Analyses = append(storage.Analyses, Analysis{
            Date:   date,
            Values: values,
        })
    }
    
    storage.Save()
    
    msg := fmt.Sprintf("✅ Анализы за %s сохранены:\n\n", date)
    for k, v := range values {
        msg += fmt.Sprintf("📊 %s: %s\n", k, v)
    }
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
}

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
    for _, a := range storage.Analyses {
        if a.Date == date {
            msg := fmt.Sprintf("📊 Анализы Папуша за %s:\n\n", date)
            keys := make([]string, 0, len(a.Values))
            for k := range a.Values {
                keys = append(keys, k)
            }
            sort.Strings(keys)
            for _, k := range keys {
                msg += fmt.Sprintf("• %s: %s\n", k, a.Values[k])
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

func listAnalysisHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    if len(storage.Analyses) == 0 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "📊 Нет сохранённых анализов",
        })
        return
    }
    
    msg := "📊 Список дат с анализами:\n\n"
    sort.Slice(storage.Analyses, func(i, j int) bool {
        return storage.Analyses[i].Date > storage.Analyses[j].Date
    })
    
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

// --- Обработчики кнопок-подсказок ---
func promptAddVetHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "➕ Чтобы добавить визит, введите команду:\n`/addvet 2026-05-20 14:30 Осмотр ул.Ленина 5`\n\n📅 Формат: `/addvet ГГГГ-ММ-ДД ЧЧ:ММ Описание Адрес`",
        ParseMode: "Markdown",
    })
}

func promptDelVetHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "🗑 Чтобы удалить визит, введите команду:\n`/delvet 2026-05-20`\n\n📅 Формат: `/delvet ГГГГ-ММ-ДД`",
        ParseMode: "Markdown",
    })
}

func promptAddMedHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "💊 Чтобы добавить лекарство, используйте формат:\n`/addmed Название|Время|Дозировка|дни|неделя`\n\nПример:\n`/addmed Креон|10:55|1/2 капсулы|пн,вт,ср,чт,пт,сб,вс|all`\n\n📌 Дни: пн,вт,ср,чт,пт,сб,вс\n📌 Неделя: all (каждую), odd (1 неделя), even (2 неделя)",
        ParseMode: "Markdown",
    })
}

func promptEditMedHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "✏️ Чтобы редактировать лекарство:\n`/editmed \"Название\"|поле=значение`\n\nДоступные поля: name, time, dosage, days, week, active\n\nПримеры:\n`/editmed Креон|time=14:00`\n`/editmed Креон|active=false`",
        ParseMode: "Markdown",
    })
}

func promptGetAnalysisHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "📈 Чтобы посмотреть анализы за дату:\n`/getanalysis 2026-04-01`\n\n📅 Формат: `/getanalysis ГГГГ-ММ-ДД`",
        ParseMode: "Markdown",
    })
}

func promptSetWeekHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "🔄 Чтобы переключить неделю:\n`/setweek odd` — нечётная (1)\n`/setweek even` — чётная (2)\n`/setweek auto` — автоматически",
        ParseMode: "Markdown",
    })
}

func helpHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    msg := `❓ **Помощь по командам**

📅 **Планирование:**
/daily - план на сегодня
/week - текущая неделя
/setweek odd/even/auto

🏥 **Визиты:**
/addvet ГГГГ-ММ-ДД ЧЧ:ММ Описание Адрес
/delvet ГГГГ-ММ-ДД
/nextvet - ближайший визит

💊 **Лекарства:**
/medlist - список
/addmed Название|Время|Дозировка|дни|неделя
/delmed "название"
/editmed "название"|поле=значение
/togglem ed "название"

📊 **Анализы:**
/addanalysis ГГГГ-ММ-ДД показатель=значение
/getanalysis ГГГГ-ММ-ДД
/delanalysis ГГГГ-ММ-ДД
/listanalysis

👆 Также можно пользоваться кнопками внизу экрана.`
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID:    update.Message.Chat.ID,
        Text:      msg,
        ParseMode: "Markdown",
    })
}
