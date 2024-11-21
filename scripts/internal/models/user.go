package models

import (
    "time"
)

type User struct {
    ID             string    `json:"id"`
    Username       string    `json:"username"`
    DiscordID      string    `json:"discord_id"`
    Token          string    `json:"token"`
    Persona        string    `json:"persona"`
    Settings       Settings  `json:"settings"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
    LastActiveAt   time.Time `json:"last_active_at"`
}

type Settings struct {
    Temperature      float64 `json:"temperature"`
    MaxTokens       int     `json:"max_tokens"`
    Model           string  `json:"model"`
    StreamResponses bool    `json:"stream_responses"`
    Language        string  `json:"language"`
    Theme           string  `json:"theme"`
}

func NewUser(discordID, username string) *User {
    return &User{
        ID:        GenerateID(),
        Username:  username,
        DiscordID: discordID,
        Settings: Settings{
            Temperature:      0.83,
            MaxTokens:       1096,
            Model:           "gpt-4-turbo",
            StreamResponses: true,
            Language:        "en",
            Theme:           "default",
        },
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
        LastActiveAt: time.Now(),
    }
}

func (u *User) UpdateSettings(settings Settings) {
    u.Settings = settings
    u.UpdatedAt = time.Now()
}

func (u *User) SetPersona(persona string) {
    u.Persona = persona
    u.UpdatedAt = time.Now()
}

func (u *User) SetToken(token string) {
    u.Token = token
    u.UpdatedAt = time.Now()
}

func (u *User) UpdateActivity() {
    u.LastActiveAt = time.Now()
    u.UpdatedAt = time.Now()
}

func (u *User) IsActive(timeout time.Duration) bool {
    return time.Since(u.LastActiveAt) < timeout
}

func (u *User) GetUserStats() map[string]interface{} {
    return map[string]interface{}{
        "user_id":      u.ID,
        "discord_id":   u.DiscordID,
        "created_at":   u.CreatedAt,
        "last_active":  u.LastActiveAt,
        "has_persona":  u.Persona != "",
        "has_token":    u.Token != "",
        "model":        u.Settings.Model,
        "temperature":  u.Settings.Temperature,
    }
}
