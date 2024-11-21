package services

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
    "net/http"
)

type OpenAIService struct {
    proxyClient *ProxyClient
    cache       map[string][]Message
    mu          sync.RWMutex
    timeout     time.Duration
    apiKey string
    client *http.Client
}

type CompletionRequest struct {
    UserID      string
    Messages    []Message
    MaxTokens   int
    Temperature float64
}

func NewOpenAIService(apiKey string) *OpenAIService {
    return &OpenAIService{
        apiKey: apiKey,
        client: &http.Client{},
    }
}


func (s *OpenAIService) GenerateCompletion(ctx context.Context, req CompletionRequest) (string, error) {
    s.mu.Lock()
    messages := s.getMessages(req.UserID)
    messages = append(messages, req.Messages...)
    s.cache[req.UserID] = messages
    s.mu.Unlock()

    response, err := s.proxyClient.SendRequest(req.UserID, messages)
    if err != nil {
        return "", fmt.Errorf("completion generation failed: %v", err)
    }

    s.mu.Lock()
    s.cache[req.UserID] = append(s.cache[req.UserID], Message{
        Role:    "assistant",
        Content: response,
    })
    s.mu.Unlock()

    return response, nil
}

func (s *OpenAIService) RegenerateLastResponse(userID string) (string, error) {
    s.mu.Lock()
    messages := s.cache[userID]
    if len(messages) < 2 {
        s.mu.Unlock()
        return "", fmt.Errorf("no previous messages to regenerate")
    }

    // Remove last assistant message
    messages = messages[:len(messages)-1]
    s.cache[userID] = messages
    s.mu.Unlock()

    return s.GenerateCompletion(context.Background(), CompletionRequest{
        UserID:   userID,
        Messages: messages,
    })
}

func (s *OpenAIService) ClearHistory(userID string) {
    s.mu.Lock()
    delete(s.cache, userID)
    s.mu.Unlock()
}

func (s *OpenAIService) GetHistory(userID string) []Message {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.getMessages(userID)
}

func (s *OpenAIService) AddSystemPrompt(userID, content string) {
    s.mu.Lock()
    messages := s.getMessages(userID)
    messages = append([]Message{{Role: "system", Content: content}}, messages...)
    s.cache[userID] = messages
    s.mu.Unlock()
}

func (s *OpenAIService) AddUserMessage(userID, content string) {
    s.mu.Lock()
    s.cache[userID] = append(s.getMessages(userID), Message{
        Role:    "user",
        Content: content,
    })
    s.mu.Unlock()
}

func (s *OpenAIService) getMessages(userID string) []Message {
    messages, exists := s.cache[userID]
    if !exists {
        return []Message{}
    }
    return messages
}

func (s *OpenAIService) ExportChat(userID string) ([]byte, error) {
    s.mu.RLock()
    messages := s.getMessages(userID)
    s.mu.RUnlock()

    return json.MarshalIndent(messages, "", "  ")
}

func (s *OpenAIService) ImportChat(userID string, data []byte) error {
    var messages []Message
    if err := json.Unmarshal(data, &messages); err != nil {
        return err
    }

    s.mu.Lock()
    s.cache[userID] = messages
    s.mu.Unlock()

    return nil
}
