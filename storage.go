package main

import (
    "encoding/json"
    "fmt"
    "os"
    "sync"
    "time"
)

type Medicine struct {
    Name        string   `json:"name"`
    Time        string   `json:"time"`
    Dosage      string   `json:"dosage"`
    WeekPattern string   `json:"week_pattern"`
    Days        []string `json:"days"`
    IsActive    bool     `json:"is_active"`
}

type VetVisit struct {
    Date        string `json:"date"`
    Time        string `json:"time"`
    Description string `json:"description"`
    Address     string `json:"address"`
}

type Analysis struct {
    Date   string            `json:"date"`
    Values map[string]string `json:"values"`
}

type PetData struct {
    Name       string     `json:"name"`
    Medicines  []Medicine `json:"medicines"`
    VetVisits  []VetVisit `json:"vet_visits"`
    Analyses   []Analysis `json:"analyses"`
    NotifyChat int64      `json:"-"`
    OverrideWeek string   `json:"override_week"` 
    mu         sync.Mutex
}

var storage = &PetData{
    Name: "Папуш",
    Medicines: []Medicine{
        {Name: "Креон", Time: "10:55", Dosage: "1/2 капсулы", WeekPattern: "all", Days: []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}, IsActive: true},
        {Name: "Креон", Time: "16:55", Dosage: "1/2 капсулы", WeekPattern: "all", Days: []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}, IsActive: true},
        {Name: "Ренал Адванс", Time: "10:55", Dosage: "1 мерную ложку", WeekPattern: "all", Days: []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}, IsActive: true},
        {Name: "Нефроспас", Time: "12:00", Dosage: "0,65 мл", WeekPattern: "all", Days: []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}, IsActive: true},
        {Name: "Мальт паста", Time: "12:00", Dosage: "1 доза", WeekPattern: "all", Days: []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}, IsActive: true},
        {Name: "Мальт паста", Time: "23:30", Dosage: "1 доза", WeekPattern: "all", Days: []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}, IsActive: true},
        {Name: "Чистка Зубов + хлоргексидин", Time: "22:00", Dosage: "обработка", WeekPattern: "odd", Days: []string{"пн", "ср", "пт", "вс"}, IsActive: true},
        {Name: "Помыть миски", Time: "15:00", Dosage: "обязательно", WeekPattern: "odd", Days: []string{"пн", "ср", "пт", "вс"}, IsActive: true},
        {Name: "Чистка Зубов + хлоргексидин", Time: "22:00", Dosage: "обработка", WeekPattern: "even", Days: []string{"вт", "чт", "сб"}, IsActive: true},
        {Name: "Помыть миски", Time: "15:00", Dosage: "обязательно", WeekPattern: "even", Days: []string{"вт", "чт", "сб"}, IsActive: true},
    },
    VetVisits: []VetVisit{},
    Analyses:  []Analysis{},
}

func (p *PetData) Save() error {
    p.mu.Lock()
    defer p.mu.Unlock()
    data, err := json.MarshalIndent(p, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile("data/pet_data.json", data, 0644)
}

func (p *PetData) Load() error {
    p.mu.Lock()
    defer p.mu.Unlock()
    file, err := os.ReadFile("data/pet_data.json")
    if err != nil {
        if os.IsNotExist(err) {
            return p.Save()
        }
        return err
    }
    return json.Unmarshal(file, p)
}

func isOddWeek() bool {
    // Если есть ручное переопределение
    if storage.OverrideWeek == "odd" {
        return true
    }
    if storage.OverrideWeek == "even" {
        return false
    }
    // Иначе автоматически
    _, week := time.Now().ISOWeek()
    return week%2 == 0
}


func isWeekMatch(weekPattern string) bool {
    if weekPattern == "all" {
        return true
    }
    currentOdd := isOddWeek()
    if weekPattern == "odd" && currentOdd {
        return true
    }
    if weekPattern == "even" && !currentOdd {
        return true
    }
    return false
}

func getDayRussian(day time.Weekday) string {
    days := map[time.Weekday]string{
        time.Monday:    "пн",
        time.Tuesday:   "вт",
        time.Wednesday: "ср",
        time.Thursday:  "чт",
        time.Friday:    "пт",
        time.Saturday:  "сб",
        time.Sunday:    "вс",
    }
    return days[day]
}

func getWeekType() string {
    if storage.OverrideWeek == "odd" {
        return "НЕЧЁТНАЯ (1 неделя) [РУЧНОЙ РЕЖИМ]"
    }
    if storage.OverrideWeek == "even" {
        return "ЧЁТНАЯ (2 неделя) [РУЧНОЙ РЕЖИМ]"
    }
    if isOddWeek() {
        return "НЕЧЁТНАЯ (1 неделя) [АВТО]"
    }
    return "ЧЁТНАЯ (2 неделя) [АВТО]"
}

func getDailyPlan() string {
    today := time.Now().Format("2006-01-02")
    currentDay := getDayRussian(time.Now().Weekday())
    weekType := getWeekType()

    result := fmt.Sprintf("📅 План на %s (%s) для кота %s\n🏷 Неделя: %s\n\n",
        today, currentDay, storage.Name, weekType)

    medsToday := []string{}
    for _, m := range storage.Medicines {
        if !m.IsActive {
            continue
        }

        dayMatch := false
        for _, d := range m.Days {
            if d == currentDay {
                dayMatch = true
                break
            }
        }

        if dayMatch && isWeekMatch(m.WeekPattern) {
            medsToday = append(medsToday, fmt.Sprintf("  💊 %s в %s - %s", m.Name, m.Time, m.Dosage))
        }
    }

    if len(medsToday) > 0 {
        result += "Лекарства:\n" + joinLines(medsToday)
    } else {
        result += "Лекарств на сегодня нет."
    }

    visitsToday := []string{}
    for _, v := range storage.VetVisits {
        if v.Date == today {
            visitsToday = append(visitsToday, fmt.Sprintf("  🏥 %s в %s - %s (%s)", v.Description, v.Time, v.Address))
        }
    }
    if len(visitsToday) > 0 {
        result += "\n\nВизиты к ветеринару:\n" + joinLines(visitsToday)
    }

    return result
}

func joinLines(lines []string) string {
    res := ""
    for _, l := range lines {
        res += l + "\n"
    }
    return res
}
