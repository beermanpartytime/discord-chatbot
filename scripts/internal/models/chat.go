package models

import (
    "time"
)

type Chat struct {
    ID            string    `json:"id"`
    UserID        string    `json:"user_id"`
    Messages      []Message `json:"messages"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    LastMessageAt time.Time `json:"last_message_at"`
}

type Message struct {
    ID        string    `json:"id"`
    ChatID    string    `json:"chat_id"`
    Role      string    `json:"role"`      // system, user, assistant
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    Metadata  MessageMetadata `json:"metadata"`
}

type MessageMetadata struct {
    TokenCount   int     `json:"token_count"`
    Temperature  float64 `json:"temperature"`
    Model       string  `json:"model"`
    IsStreaming bool    `json:"is_streaming"`
}

func NewChat(userID string) *Chat {
    return &Chat{
        ID:            GenerateID(),
        UserID:        userID,
        Messages:      make([]Message, 0),
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
        LastMessageAt: time.Now(),
    }
}

func (c *Chat) AddMessage(role, content string, metadata MessageMetadata) Message {
    message := Message{
        ID:        GenerateID(),
        ChatID:    c.ID,
        Role:      role,
        Content:   content,
        CreatedAt: time.Now(),
        Metadata:  metadata,
    }
    
    c.Messages = append(c.Messages, message)
    c.UpdatedAt = time.Now()
    c.LastMessageAt = time.Now()
    
    return message
}

func (c *Chat) GetLastMessage() *Message {
    if len(c.Messages) == 0 {
        return nil
    }
    return &c.Messages[len(c.Messages)-1]
}

func (c *Chat) RemoveLastMessage() bool {
    if len(c.Messages) == 0 {
        return false
    }
    
    c.Messages = c.Messages[:len(c.Messages)-1]
    c.UpdatedAt = time.Now()
    
    if len(c.Messages) > 0 {
        c.LastMessageAt = c.Messages[len(c.Messages)-1].CreatedAt
    }
    
    return true
}

func (c *Chat) GetMessageHistory() []Message {
    return c.Messages
}

func (c *Chat) ClearHistory() {
    c.Messages = make([]Message, 0)
    c.UpdatedAt = time.Now()
    c.LastMessageAt = time.Now()
}

func (c *Chat) GetMessageCount() int {
    return len(c.Messages)
}

func (c *Chat) GetTotalTokens() int {
    total := 0
    for _, msg := range c.Messages {
        total += msg.Metadata.TokenCount
    }
    return total
}

func (c *Chat) IsActive(timeout time.Duration) bool {
    return time.Since(c.LastMessageAt) < timeout
}
