package main

import (
    "sync"
)

type DialogStep int

const (
    StepNone DialogStep = iota
    // Добавление лекарства
    StepAddMedName
    StepAddMedTime
    StepAddMedDosage
    StepAddMedDays
    StepAddMedWeekPattern
    // Добавление визита
    StepAddVetDate
    StepAddVetTime
    StepAddVetDescription
    StepAddVetAddress
    // Редактирование лекарства
    StepEditMedSelect
    StepEditMedField
    // Анализы
    StepGetAnalysisDate
    // Неделя
    StepSetWeek
)

type AddMedData struct {
    Name        string
    Time        string
    Dosage      string
    Days        []string
    WeekPattern string
}

type AddVetData struct {
    Date        string
    Time        string
    Description string
    Address     string
}

type UserState struct {
    Step        DialogStep
    AddMed      AddMedData
    AddVet      AddVetData
    EditMedName string
}

var (
    userStates = make(map[int64]*UserState)
    stateMutex sync.Mutex
)

func GetUserState(chatID int64) *UserState {
    stateMutex.Lock()
    defer stateMutex.Unlock()
    if _, ok := userStates[chatID]; !ok {
        userStates[chatID] = &UserState{Step: StepNone}
    }
    return userStates[chatID]
}

func SetUserState(chatID int64, state *UserState) {
    stateMutex.Lock()
    defer stateMutex.Unlock()
    userStates[chatID] = state
}

func ClearUserState(chatID int64) {
    stateMutex.Lock()
    defer stateMutex.Unlock()
    delete(userStates, chatID)
}
