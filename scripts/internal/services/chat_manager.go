package services

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

var startTime = time.Now()

type ChatManager struct {
    openAI        *OpenAIService
    promptManager *PromptManager
    sessions      map[string]*ChatSession
    mu            sync.RWMutex
}

type ChatSession struct {
    Messages     []Message
    LastActivity time.Time
    IsStreaming  bool
}

func NewChatManager(openAI *OpenAIService, promptManager *PromptManager) *ChatManager {
    return &ChatManager{
        openAI:        openAI,
        promptManager: promptManager,
        sessions:      make(map[string]*ChatSession),
    }
}

func (cm *ChatManager) CreateNewChat(userID string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    prompts := cm.promptManager.BuildPromptList(userID)
    cm.sessions[userID] = &ChatSession{
        Messages:     prompts,
        LastActivity: time.Now(),
        IsStreaming:  false,
    }
}

func (cm *ChatManager) AddMessage(userID string, role string, content string) {
    cm.mu.Lock()
    session := cm.getOrCreateSession(userID)
    session.Messages = append(session.Messages, Message{
        ID:        GenerateID(),
        Role:      role,
        Content:   content,
        Timestamp: time.Now(),
    })
    session.LastActivity = time.Now()
    cm.mu.Unlock()
}

func (cm *ChatManager) GenerateResponse(userID string) (string, error) {
    cm.mu.Lock()
    session := cm.getOrCreateSession(userID)
    messages := session.Messages
    cm.mu.Unlock()

    response, err := cm.openAI.GenerateCompletion(context.Background(), CompletionRequest{
        UserID:   userID,
        Messages: messages,
    })

    if err == nil {
        cm.AddMessage(userID, "assistant", response)
    }
    return response, err
}

func (cm *ChatManager) GetChatHistory(userID string) []Message {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    session := cm.getOrCreateSession(userID)
    return session.Messages
}

func (cm *ChatManager) ClearChat(userID string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    prompts := cm.promptManager.BuildPromptList(userID)
    cm.sessions[userID] = &ChatSession{
        Messages:     prompts,
        LastActivity: time.Now(),
    }
}

func (cm *ChatManager) RemoveLastMessage(userID string) bool {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    session := cm.getOrCreateSession(userID)
    if len(session.Messages) > 0 {
        session.Messages = session.Messages[:len(session.Messages)-1]
        session.LastActivity = time.Now()
        return true
    }
    return false
}

func (cm *ChatManager) MessageExists(messageID string) bool {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    for _, session := range cm.sessions {
        for _, msg := range session.Messages {
            if msg.ID == messageID {
                return true
            }
        }
    }
    return false
}

func (cm *ChatManager) UpdateMessage(messageID, content string) bool {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    
    for _, session := range cm.sessions {
        for i, msg := range session.Messages {
            if msg.ID == messageID {
                session.Messages[i].Content = content
                return true
            }
        }
    }
    return false
}

func (cm *ChatManager) DeleteMessage(messageID string) bool {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    
    for _, session := range cm.sessions {
        for i, msg := range session.Messages {
            if msg.ID == messageID {
                session.Messages = append(session.Messages[:i], session.Messages[i+1:]...)
                return true
            }
        }
    }
    return false
}

func (cm *ChatManager) GetActiveUserCount() int {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    return len(cm.sessions)
}

func (cm *ChatManager) CleanupInactiveSessions(timeout time.Duration) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    now := time.Now()
    for userID, session := range cm.sessions {
        if now.Sub(session.LastActivity) > timeout {
            delete(cm.sessions, userID)
        }
    }
}

func (cm *ChatManager) getOrCreateSession(userID string) *ChatSession {
    session, exists := cm.sessions[userID]
    if !exists {
        prompts := cm.promptManager.BuildPromptList(userID)
        session = &ChatSession{
            Messages:     prompts,
            LastActivity: time.Now(),
        }
        cm.sessions[userID] = session
    }
    return session
}

func (cm *ChatManager) SaveAllSessions() {
    // Implementation for saving sessions to persistent storage
}

func (cm *ChatManager) GetUptime() time.Duration {
    return time.Since(startTime)
}

func (cm *ChatManager) GetTotalChats() int {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    return len(cm.sessions)
}

func (cm *ChatManager) IsUserInCooldown(userID string) bool {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    session, exists := cm.sessions[userID]
    if !exists {
        return false
    }
    return time.Since(session.LastActivity) < time.Second * 3
}

func (cm *ChatManager) RegenerateLastResponse(userID string) string {
    cm.mu.Lock()
    session := cm.getOrCreateSession(userID)
    if len(session.Messages) < 2 {
        cm.mu.Unlock()
        return "No message to regenerate!"
    }
    
    session.Messages = session.Messages[:len(session.Messages)-1]
    cm.mu.Unlock()
    
    response, _ := cm.GenerateResponse(userID)
    return response
}

func (cm *ChatManager) GetMemoryStats(userID string) struct {
    UsedTokens   int
    MaxTokens    int
    MessageCount int
    ContextSize  float64
} {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    session := cm.getOrCreateSession(userID)
    return struct {
        UsedTokens   int
        MaxTokens    int
        MessageCount int
        ContextSize  float64
    }{
        UsedTokens:   len(session.Messages) * 100,
        MaxTokens:    4096,
        MessageCount: len(session.Messages),
        ContextSize:  float64(len(session.Messages)) * 1.5,
    }
}

func (cm *ChatManager) ContinueChat(userID string) string {
    cm.mu.Lock()
    session := cm.getOrCreateSession(userID)
    lastMessage := session.Messages[len(session.Messages)-1].Content
    cm.mu.Unlock()
    
    response, _ := cm.GenerateResponse(userID)
    return "Continuing from: " + lastMessage + "\n\n" + response
}

func (cm *ChatManager) SaveChat(userID string) string {
    chatID := GenerateID()
    session := cm.getOrCreateSession(userID)
    
    cm.sessions[chatID] = &ChatSession{
        Messages:     session.Messages,
        LastActivity: time.Now(),
    }
    
    return chatID
}

func (cm *ChatManager) LoadChat(userID, chatID string) error {
    return nil
}

func (cm *ChatManager) ExportChat(userID, format string) string {
    cm.mu.RLock()
    session := cm.getOrCreateSession(userID)
    cm.mu.RUnlock()
    
    var export string
    if format == "json" {
        data, _ := json.Marshal(session.Messages)
        export = string(data)
    } else {
        for _, msg := range session.Messages {
            export += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
        }
    }
    
    return export
}

func GenerateID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
