package models

import (
    "time"
)

type Prompt struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Type      string    `json:"type"`      // system, user, assistant
    Content   string    `json:"content"`
    Depth     int       `json:"depth"`     // Order in the prompt list
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type PromptDefinitions struct {
    Description    string `json:"description"`
    Personality   string `json:"personality"`
    Scenario      string `json:"scenario"`
    FirstMessage  string `json:"first_message"`
    AuthorsNote   string `json:"authors_note"`
}

type UserSettings struct {
    UserID      string    `json:"user_id"`
    UserPersona string    `json:"user_persona"`
    UserToken   string    `json:"user_token"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type PromptList struct {
    UserID      string    `json:"user_id"`
    Prompts     []Prompt  `json:"prompts"`
    Definitions PromptDefinitions `json:"definitions"`
    Settings    UserSettings      `json:"settings"`
}

func NewPromptList(userID string) *PromptList {
    return &PromptList{
        UserID:  userID,
        Prompts: make([]Prompt, 0),
        Definitions: PromptDefinitions{},
        Settings: UserSettings{
            UserID:    userID,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }
}

func (pl *PromptList) AddPrompt(content string, promptType string, depth int) {
    prompt := Prompt{
        ID:        GenerateID(),
        UserID:    pl.UserID,
        Type:      promptType,
        Content:   content,
        Depth:     depth,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    pl.Prompts = append(pl.Prompts, prompt)
}

func (pl *PromptList) UpdatePrompt(promptID string, content string) bool {
    for i, prompt := range pl.Prompts {
        if prompt.ID == promptID {
            pl.Prompts[i].Content = content
            pl.Prompts[i].UpdatedAt = time.Now()
            return true
        }
    }
    return false
}

func (pl *PromptList) RemovePrompt(promptID string) bool {
    for i, prompt := range pl.Prompts {
        if prompt.ID == promptID {
            pl.Prompts = append(pl.Prompts[:i], pl.Prompts[i+1:]...)
            return true
        }
    }
    return false
}

func GenerateID() string {
    // Implementation for generating unique IDs
    return time.Now().Format("20060102150405") + RandomString(6)
}

func RandomString(length int) string {
    // Implementation for generating random strings
    // Used for unique ID generation
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, length)
    for i := range result {
        result[i] = charset[rand.Intn(len(charset))]
    }
    return string(result)
}
