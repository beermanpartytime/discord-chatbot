package services

import (
    "encoding/json"
    "sync"
)

type PromptManager struct {
    prompts    map[string]*UserPrompts
    mu         sync.RWMutex
}

type UserPrompts struct {
    Description    string
    Personality   string
    Scenario      string
    FirstMessage  string
    AuthorsNote   string
    UserPersona   string
    UserToken     string
    SystemPrompts []string
}

func NewPromptManager() *PromptManager {
    return &PromptManager{
        prompts: make(map[string]*UserPrompts),
    }
}

func (pm *PromptManager) GetUserPrompts(userID string) *UserPrompts {
    pm.mu.RLock()
    defer pm.mu.RUnlock()

    prompts, exists := pm.prompts[userID]
    if !exists {
        return pm.createDefaultPrompts()
    }
    return prompts
}

func (pm *PromptManager) UpdateDefinitions(userID string, definitions map[string]string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    prompts := pm.getOrCreatePrompts(userID)
    
    if desc, ok := definitions["description"]; ok {
        prompts.Description = desc
    }
    if pers, ok := definitions["personality"]; ok {
        prompts.Personality = pers
    }
    if scen, ok := definitions["scenario"]; ok {
        prompts.Scenario = scen
    }
}

func (pm *PromptManager) SetUserPersona(userID, persona string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    prompts := pm.getOrCreatePrompts(userID)
    prompts.UserPersona = persona
}

func (pm *PromptManager) SetUserToken(userID, token string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    prompts := pm.getOrCreatePrompts(userID)
    prompts.UserToken = token
}

func (pm *PromptManager) SetFirstMessage(userID, message string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    prompts := pm.getOrCreatePrompts(userID)
    prompts.FirstMessage = message
}

func (pm *PromptManager) SetAuthorsNote(userID, note string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    prompts := pm.getOrCreatePrompts(userID)
    prompts.AuthorsNote = note
}

func (pm *PromptManager) AddSystemPrompt(userID, prompt string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    prompts := pm.getOrCreatePrompts(userID)
    prompts.SystemPrompts = append(prompts.SystemPrompts, prompt)
}

func (pm *PromptManager) BuildPromptList(userID string) []Message {
    pm.mu.RLock()
    prompts := pm.getOrCreatePrompts(userID)
    pm.mu.RUnlock()

    var messages []Message

    // Add system prompts in order
    if prompts.Description != "" {
        messages = append(messages, Message{
            Role:    "system",
            Content: "Character Description: " + prompts.Description,
        })
    }

    if prompts.Personality != "" {
        messages = append(messages, Message{
            Role:    "system",
            Content: "Personality: " + prompts.Personality,
        })
    }

    if prompts.Scenario != "" {
        messages = append(messages, Message{
            Role:    "system",
            Content: "Scenario: " + prompts.Scenario,
        })
    }

    if prompts.UserPersona != "" {
        messages = append(messages, Message{
            Role:    "system",
            Content: "User Persona: " + prompts.UserPersona,
        })
    }

    if prompts.FirstMessage != "" {
        messages = append(messages, Message{
            Role:    "assistant",
            Content: prompts.FirstMessage,
        })
    }

    return messages
}

func (pm *PromptManager) ExportPrompts(userID string) ([]byte, error) {
    pm.mu.RLock()
    prompts := pm.getOrCreatePrompts(userID)
    pm.mu.RUnlock()

    return json.MarshalIndent(prompts, "", "  ")
}

func (pm *PromptManager) ImportPrompts(userID string, data []byte) error {
    var prompts UserPrompts
    if err := json.Unmarshal(data, &prompts); err != nil {
        return err
    }

    pm.mu.Lock()
    pm.prompts[userID] = &prompts
    pm.mu.Unlock()

    return nil
}

func (pm *PromptManager) getOrCreatePrompts(userID string) *UserPrompts {
    prompts, exists := pm.prompts[userID]
    if !exists {
        prompts = pm.createDefaultPrompts()
        pm.prompts[userID] = prompts
    }
    return prompts
}

func (pm *PromptManager) createDefaultPrompts() *UserPrompts {
    return &UserPrompts{
        SystemPrompts: make([]string, 0),
    }
}
