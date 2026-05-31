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
    // Команды
    b.RegisterHandler(tb.HandlerTypeMessageText, "/start", tb.MatchTypeExact, startHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/menu", tb.MatchTypeExact, menuHandler)
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
    b.RegisterHandler(tb.HandlerTypeMessageText, "/addmed", tb.MatchTypeExact, addMedHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/delmed", tb.MatchTypePrefix, delMedHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/editmed", tb.MatchTypePrefix, editMedHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/togglem ed", tb.MatchTypePrefix, toggleMedHandler)
    b.RegisterHandler(tb.HandlerTypeMessageText, "/cancel", tb.MatchTypeExact, cancelHandler)
    
    // Callback-обработчик для инлайн-кнопок
    b.RegisterHandler(tb.HandlerTypeCallbackQueryData, "", tb.MatchTypePrefix, callbackHandler)
    
    // Обработчик текстовых сообщений для диалогов
    b.RegisterHandler(tb.HandlerTypeMessageText, "", tb.MatchTypePrefix, textMessageHandler)
}

// ----------------------------------------------------------------
//  СТАРТ (убираем reply-клавиатуру)
// ----------------------------------------------------------------
func startHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    storage.Load()
    storage.NotifyChat = update.Message.Chat.ID
    storage.Save()
    
    // Убираем старую reply-клавиатуру, если она была
    removeKeyboard := &models.ReplyKeyboardRemove{
        RemoveKeyboard: true,
    }
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID:      update.Message.Chat.ID,
        Text:        fmt.Sprintf("🐱 Привет! Я бот для кота %s.\n\nТекущая неделя: %s\n\nИспользуйте /menu для вызова главного меню.", storage.Name, getWeekType()),
        ReplyMarkup: removeKeyboard,
    })
    
    if !schedulerStarted {
        schedulerStarted = true
        startScheduler(b, update.Message.Chat.ID)
    }
}

// ----------------------------------------------------------------
//  МЕНЮ (инлайн-кнопки под сообщением)
// ----------------------------------------------------------------
func menuHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    keyboard := &models.InlineKeyboardMarkup{
        InlineKeyboard: [][]models.InlineKeyboardButton{
            {
                {Text: "📅 План на сегодня", CallbackData: "daily"},
                {Text: "📊 Моя неделя", CallbackData: "week"},
            },
            {
                {Text: "💊 Все лекарства", CallbackData: "medlist"},
                {Text: "🏥 Ближайший визит", CallbackData: "nextvet"},
            },
            {
                {Text: "➕ Добавить визит", CallbackData: "addvet_start"},
                {Text: "🗑 Удалить визит", CallbackData: "delvet_prompt"},
            },
            {
                {Text: "💊 Добавить лекарство", CallbackData: "addmed_start"},
                {Text: "✏️ Редактировать лекарство", CallbackData: "editmed_start"},
            },
            {
                {Text: "📈 Анализы за дату", CallbackData: "getanalysis_start"},
                {Text: "📋 Список анализов", CallbackData: "listanalysis"},
            },
            {
                {Text: "🔄 Переключить неделю", CallbackData: "setweek_start"},
                {Text: "❓ Помощь", CallbackData: "help"},
            },
        },
    }
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID:      update.Message.Chat.ID,
        Text:        "📋 Главное меню:",
        ReplyMarkup: keyboard,
    })
}

// ----------------------------------------------------------------
//  ОБРАБОТЧИК INLINE-КНОПОК
// ----------------------------------------------------------------
func callbackHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    chatID := update.CallbackQuery.Message.Message.Chat.ID
    data := update.CallbackQuery.Data
    
    b.AnswerCallbackQuery(ctx, &tb.AnswerCallbackQueryParams{
        CallbackQueryID: update.CallbackQuery.ID,
    })
    
    fakeUpdate := &models.Update{
        Message: &models.Message{
            Chat: models.Chat{ID: chatID},
        },
    }
    
    switch data {
    case "daily":
        dailyHandler(ctx, b, fakeUpdate)
    case "week":
        weekHandler(ctx, b, fakeUpdate)
    case "medlist":
        medListHandler(ctx, b, fakeUpdate)
    case "nextvet":
        nextVetHandler(ctx, b, fakeUpdate)
    case "listanalysis":
        listAnalysisHandler(ctx, b, fakeUpdate)
    case "help":
        helpHandler(ctx, b, fakeUpdate)
    case "addvet_start":
        startAddVetDialog(ctx, b, fakeUpdate)
    case "delvet_prompt":
        promptDelVetHandler(ctx, b, fakeUpdate)
    case "addmed_start":
        addMedHandler(ctx, b, fakeUpdate)
    case "editmed_start":
        startEditMedDialog(ctx, b, fakeUpdate)
    case "getanalysis_start":
        startGetAnalysisDialog(ctx, b, fakeUpdate)
    case "setweek_start":
        startSetWeekDialog(ctx, b, fakeUpdate)
    default:
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "Неизвестная команда. Используйте /menu",
        })
    }
}

// ----------------------------------------------------------------
//  ОСНОВНЫЕ ФУНКЦИИ (без изменений)
// ----------------------------------------------------------------
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

// ----------------------------------------------------------------
//  ДИАЛОГИ
// ----------------------------------------------------------------
func addMedHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    chatID := update.Message.Chat.ID
    ClearUserState(chatID)
    state := &UserState{
        Step:   StepAddMedName,
        AddMed: AddMedData{},
    }
    SetUserState(chatID, state)
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: chatID,
        Text:   "💊 Давайте добавим новое лекарство.\n\nВведите НАЗВАНИЕ лекарства.\n\n(Для отмены отправьте /cancel)",
    })
}

func startEditMedDialog(ctx context.Context, b *tb.Bot, update *models.Update) {
    chatID := update.Message.Chat.ID
    ClearUserState(chatID)
    
    if len(storage.Medicines) == 0 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "❌ Нет лекарств для редактирования. Сначала добавьте лекарство через /addmed",
        })
        return
    }
    
    list := "📋 Список лекарств:\n\n"
    for i, m := range storage.Medicines {
        list += fmt.Sprintf("%d. %s\n", i+1, m.Name)
    }
    list += "\nВведите НАЗВАНИЕ лекарства, которое хотите редактировать.\n(Для отмены отправьте /cancel)"
    
    state := &UserState{
        Step: StepEditMedSelect,
    }
    SetUserState(chatID, state)
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: chatID,
        Text:   list,
    })
}

func editMedHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    startEditMedDialog(ctx, b, update)
}

func startGetAnalysisDialog(ctx context.Context, b *tb.Bot, update *models.Update) {
    chatID := update.Message.Chat.ID
    ClearUserState(chatID)
    
    state := &UserState{
        Step: StepGetAnalysisDate,
    }
    SetUserState(chatID, state)
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: chatID,
        Text:   "📈 Введите дату в формате ГГГГ-ММ-ДД (например, 2026-04-01)\n\n(Для отмены отправьте /cancel)",
    })
}

func startSetWeekDialog(ctx context.Context, b *tb.Bot, update *models.Update) {
    chatID := update.Message.Chat.ID
    ClearUserState(chatID)
    
    state := &UserState{
        Step: StepSetWeek,
    }
    SetUserState(chatID, state)
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: chatID,
        Text:   "🔄 Выберите режим недели:\n\n• odd — нечётная (1 неделя)\n• even — чётная (2 неделя)\n• auto — автоматический\n\nВведите: odd, even, auto\n\n(Для отмены отправьте /cancel)",
    })
}

func startAddVetDialog(ctx context.Context, b *tb.Bot, update *models.Update) {
    chatID := update.Message.Chat.ID
    ClearUserState(chatID)
    
    state := &UserState{
        Step:   StepAddVetDate,
        AddVet: AddVetData{},
    }
    SetUserState(chatID, state)
    
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: chatID,
        Text:   "🏥 Добавление нового визита.\n\nВведите ДАТУ в формате ГГГГ-ММ-ДД (например, 2026-05-20)\n\n(Для отмены отправьте /cancel)",
    })
}

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
            Text:   fmt.Sprintf("❌ Лекарство \"%s\" не найдено", name),
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

func toggleMedHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    name := strings.TrimPrefix(update.Message.Text, "/togglem ed ")
    name = strings.TrimSpace(name)
    name = strings.Trim(name, "\"'")
    
    if name == "" {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /togglem ed \"Название лекарства\"",
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
            Text:   "✅ Установлена РУЧНАЯ НЕЧЁТНАЯ (1) неделя",
        })
    case "even":
        storage.OverrideWeek = "even"
        storage.Save()
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "✅ Установлена РУЧНАЯ ЧЁТНАЯ (2) неделя",
        })
    case "auto":
        storage.OverrideWeek = ""
        storage.Save()
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "✅ Возврат к АВТОМАТИЧЕСКОМУ режиму",
        })
    default:
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Неверный параметр. Используйте: odd, even, auto",
        })
    }
}

// ----------------------------------------------------------------
//  АНАЛИЗЫ
// ----------------------------------------------------------------
func addAnalysisHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    parts := strings.Split(update.Message.Text, " ")
    if len(parts) < 3 {
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "❌ Формат: /addanalysis ГГГГ-ММ-ДД показатель=значение\nПример: /addanalysis 2026-04-01 лейкоциты=8.2",
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
            Text:   "❌ Формат: /getanalysis ГГГГ-ММ-ДД",
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
            Text:   "❌ Формат: /delanalysis ГГГГ-ММ-ДД",
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
        msg += fmt.Sprintf("📅 %s (%d показателей)\n", a.Date, len(a.Values))
    }
    msg += "\nЧтобы посмотреть анализы: /getanalysis ГГГГ-ММ-ДД"
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   msg,
    })
}

// ----------------------------------------------------------------
//  ОСНОВНОЙ ОБРАБОТЧИК ДИАЛОГОВ (textMessageHandler)
// ----------------------------------------------------------------
func textMessageHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    chatID := update.Message.Chat.ID
    text := strings.TrimSpace(update.Message.Text)
    
    if text == "/cancel" {
        ClearUserState(chatID)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "✅ Диалог отменён.",
        })
        return
    }
    
    if strings.HasPrefix(text, "/") {
        return
    }
    
    state := GetUserState(chatID)
    if state.Step == StepNone {
        return
    }
    
    switch state.Step {
    case StepAddMedName:
        state.AddMed.Name = text
        state.Step = StepAddMedTime
        SetUserState(chatID, state)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "⏰ Введите время приёма (ЧЧ:ММ, например 10:55):\n(Для отмены /cancel)",
        })
    case StepAddMedTime:
        if _, err := time.Parse("15:04", text); err != nil {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Неверный формат. Используйте ЧЧ:ММ. Попробуйте снова:\n(Для отмены /cancel)",
            })
            return
        }
        state.AddMed.Time = text
        state.Step = StepAddMedDosage
        SetUserState(chatID, state)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "💊 Введите дозировку (например: 1/2 капсулы):\n(Для отмены /cancel)",
        })
    case StepAddMedDosage:
        state.AddMed.Dosage = text
        state.Step = StepAddMedDays
        SetUserState(chatID, state)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "📅 Введите дни недели через запятую (пн,вт,ср,чт,пт,сб,вс):\nПример: пн,ср,пт\n(Для отмены /cancel)",
        })
    case StepAddMedDays:
        days := strings.Split(text, ",")
        valid := true
        for _, d := range days {
            if !isValidDay(d) {
                valid = false
                break
            }
        }
        if !valid {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Неверные дни. Используйте: пн, вт, ср, чт, пт, сб, вс. Попробуйте снова:\n(Для отмены /cancel)",
            })
            return
        }
        state.AddMed.Days = days
        state.Step = StepAddMedWeekPattern
        SetUserState(chatID, state)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "📆 На какой неделе давать?\n• all — каждую неделю\n• odd — только нечётная (1 неделя)\n• even — только чётная (2 неделя)\n\nВведите: all, odd или even\n(Для отмены /cancel)",
        })
    case StepAddMedWeekPattern:
        pattern := strings.ToLower(text)
        if pattern != "all" && pattern != "odd" && pattern != "even" {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Неверно. Введите: all, odd или even\n(Для отмены /cancel)",
            })
            return
        }
        newMed := Medicine{
            Name:        state.AddMed.Name,
            Time:        state.AddMed.Time,
            Dosage:      state.AddMed.Dosage,
            Days:        state.AddMed.Days,
            WeekPattern: pattern,
            IsActive:    true,
        }
        storage.Medicines = append(storage.Medicines, newMed)
        storage.Save()
        ClearUserState(chatID)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   fmt.Sprintf("✅ Лекарство добавлено!\n\n📋 %s\n⏰ %s\n💊 %s\n📅 Дни: %s\n📆 Неделя: %s",
                newMed.Name, newMed.Time, newMed.Dosage, strings.Join(newMed.Days, ","), newMed.WeekPattern),
        })
    case StepEditMedSelect:
        found := false
        for _, m := range storage.Medicines {
            if m.Name == text {
                found = true
                break
            }
        }
        if !found {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Лекарство не найдено. Попробуйте снова или /cancel",
            })
            return
        }
        state.EditMedName = text
        state.Step = StepEditMedField
        SetUserState(chatID, state)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "✏️ Что хотите изменить?\nДоступные поля: name, time, dosage, days, week, active\n\nПример: time=14:00\n(Для отмены /cancel)",
        })
    case StepEditMedField:
        kv := strings.SplitN(text, "=", 2)
        if len(kv) != 2 {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Неверный формат. Используйте: поле=значение\nПример: time=14:00\n(Для отмены /cancel)",
            })
            return
        }
        field := kv[0]
        value := kv[1]
        index := -1
        for i, m := range storage.Medicines {
            if m.Name == state.EditMedName {
                index = i
                break
            }
        }
        if index == -1 {
            ClearUserState(chatID)
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Ошибка: лекарство не найдено",
            })
            return
        }
        med := &storage.Medicines[index]
        switch field {
        case "name":
            med.Name = value
        case "time":
            if _, err := time.Parse("15:04", value); err != nil {
                b.SendMessage(ctx, &tb.SendMessageParams{
                    ChatID: chatID,
                    Text:   "❌ Неверный формат времени. Используйте ЧЧ:ММ",
                })
                return
            }
            med.Time = value
        case "dosage":
            med.Dosage = value
        case "days":
            days := strings.Split(value, ",")
            valid := true
            for _, d := range days {
                if !isValidDay(d) {
                    valid = false
                    break
                }
            }
            if !valid {
                b.SendMessage(ctx, &tb.SendMessageParams{
                    ChatID: chatID,
                    Text:   "❌ Неверные дни. Используйте: пн,вт,ср,чт,пт,сб,вс",
                })
                return
            }
            med.Days = days
        case "week":
            if value != "all" && value != "odd" && value != "even" {
                b.SendMessage(ctx, &tb.SendMessageParams{
                    ChatID: chatID,
                    Text:   "❌ Неверно. Используйте: all, odd, even",
                })
                return
            }
            med.WeekPattern = value
        case "active":
            if value != "true" && value != "false" {
                b.SendMessage(ctx, &tb.SendMessageParams{
                    ChatID: chatID,
                    Text:   "❌ active может быть true или false",
                })
                return
            }
            med.IsActive = value == "true"
        default:
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Неизвестное поле. Доступные: name, time, dosage, days, week, active",
            })
            return
        }
        storage.Save()
        ClearUserState(chatID)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   fmt.Sprintf("✅ Лекарство обновлено!\n\n📋 %s\n⏰ %s\n💊 %s\n📅 Дни: %s\n📆 Неделя: %s\n🔘 Активно: %v",
                med.Name, med.Time, med.Dosage, strings.Join(med.Days, ","), med.WeekPattern, med.IsActive),
        })
    case StepAddVetDate:
        if _, err := time.Parse("2006-01-02", text); err != nil {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Неверный формат даты. Используйте ГГГГ-ММ-ДД (например, 2026-05-20). Попробуйте снова:\n(Для отмены /cancel)",
            })
            return
        }
        state.AddVet.Date = text
        state.Step = StepAddVetTime
        SetUserState(chatID, state)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "⏰ Введите ВРЕМЯ в формате ЧЧ:ММ (например, 14:30):\n(Для отмены /cancel)",
        })
    case StepAddVetTime:
        if _, err := time.Parse("15:04", text); err != nil {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Неверный формат времени. Используйте ЧЧ:ММ (например, 14:30). Попробуйте снова:\n(Для отмены /cancel)",
            })
            return
        }
        state.AddVet.Time = text
        state.Step = StepAddVetDescription
        SetUserState(chatID, state)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "📝 Введите ОПИСАНИЕ визита (например, Плановый осмотр):\n(Для отмены /cancel)",
        })
    case StepAddVetDescription:
        state.AddVet.Description = text
        state.Step = StepAddVetAddress
        SetUserState(chatID, state)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "📍 Введите АДРЕС ветклиники:\n(Для отмены /cancel)",
        })
    case StepAddVetAddress:
        state.AddVet.Address = text
        storage.VetVisits = append(storage.VetVisits, VetVisit{
            Date:        state.AddVet.Date,
            Time:        state.AddVet.Time,
            Description: state.AddVet.Description,
            Address:     state.AddVet.Address,
        })
        storage.Save()
        ClearUserState(chatID)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   fmt.Sprintf("✅ Визит добавлен!\n\n📅 %s\n⏰ %s\n📝 %s\n📍 %s",
                state.AddVet.Date, state.AddVet.Time, state.AddVet.Description, state.AddVet.Address),
        })
    case StepGetAnalysisDate:
        if _, err := time.Parse("2006-01-02", text); err != nil {
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Неверный формат даты. Используйте ГГГГ-ММ-ДД (например, 2026-04-01). Попробуйте снова:\n(Для отмены /cancel)",
            })
            return
        }
        ClearUserState(chatID)
        for _, a := range storage.Analyses {
            if a.Date == text {
                msg := fmt.Sprintf("📊 Анализы Папуша за %s:\n\n", text)
                keys := make([]string, 0, len(a.Values))
                for k := range a.Values {
                    keys = append(keys, k)
                }
                sort.Strings(keys)
                for _, k := range keys {
                    msg += fmt.Sprintf("• %s: %s\n", k, a.Values[k])
                }
                b.SendMessage(ctx, &tb.SendMessageParams{
                    ChatID: chatID,
                    Text:   msg,
                })
                return
            }
        }
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   fmt.Sprintf("❌ Нет анализов за %s", text),
        })
    case StepSetWeek:
        switch strings.ToLower(text) {
        case "odd":
            storage.OverrideWeek = "odd"
            storage.Save()
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "✅ Установлена РУЧНАЯ НЕЧЁТНАЯ (1) неделя",
            })
        case "even":
            storage.OverrideWeek = "even"
            storage.Save()
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "✅ Установлена РУЧНАЯ ЧЁТНАЯ (2) неделя",
            })
        case "auto":
            storage.OverrideWeek = ""
            storage.Save()
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "✅ Возврат к АВТОМАТИЧЕСКОМУ режиму",
            })
        default:
            b.SendMessage(ctx, &tb.SendMessageParams{
                ChatID: chatID,
                Text:   "❌ Неверно. Введите: odd, even или auto\n(Для отмены /cancel)",
            })
            return
        }
        ClearUserState(chatID)
    default:
        ClearUserState(chatID)
        b.SendMessage(ctx, &tb.SendMessageParams{
            ChatID: chatID,
            Text:   "🔄 Диалог прерван. Начните заново.",
        })
    }
}

// ----------------------------------------------------------------
//  ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ----------------------------------------------------------------
func promptDelVetHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID:    update.Message.Chat.ID,
        Text:      "🗑 Чтобы удалить визит, введите:\n`/delvet 2026-05-20`\n\n(Для отмены /cancel)",
        ParseMode: "Markdown",
    })
}

func helpHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    msg := `❓ **Помощь по командам**

📅 **Планирование:**
/daily, /week, /setweek

🏥 **Визиты:**
/addvet, /delvet, /nextvet

💊 **Лекарства:**
/medlist, /addmed, /delmed, /editmed, /togglem ed

📊 **Анализы:**
/addanalysis, /getanalysis, /delanalysis, /listanalysis

🛑 **Отмена диалога:**
/cancel

👆 Главное меню: /menu`
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID:    update.Message.Chat.ID,
        Text:      msg,
        ParseMode: "Markdown",
    })
}

func cancelHandler(ctx context.Context, b *tb.Bot, update *models.Update) {
    chatID := update.Message.Chat.ID
    ClearUserState(chatID)
    b.SendMessage(ctx, &tb.SendMessageParams{
        ChatID: chatID,
        Text:   "✅ Диалог отменён.",
    })
}

func isValidDay(day string) bool {
    switch day {
    case "пн", "вт", "ср", "чт", "пт", "сб", "вс":
        return true
    }
    return false
}
