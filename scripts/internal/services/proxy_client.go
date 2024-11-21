package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type ProxyClient struct {
    proxyURL    string
    password    string
    client      *http.Client
    userConfigs map[string]*UserConfig
    mu          sync.RWMutex
}

type UserConfig struct {
    Model             string
    Temperature       float64
    Stream           bool
    MaxTokens        int
    PresencePenalty  float64
    FrequencyPenalty float64
    TopP             float64
}

func NewProxyClient(proxyURL, password string) *ProxyClient {
    return &ProxyClient{
        proxyURL: proxyURL,
        password: password,
        client: &http.Client{
            Timeout: time.Second * 60,
        },
        userConfigs: make(map[string]*UserConfig),
    }
}

func (pc *ProxyClient) SendRequest(userID string, messages []Message) (string, error) {
    pc.mu.RLock()
    config := pc.getUserConfig(userID)
    pc.mu.RUnlock()

    payload := map[string]interface{}{
        "messages":          messages,
        "model":            config.Model,
        "temperature":      config.Temperature,
        "max_tokens":       config.MaxTokens,
        "stream":           config.Stream,
        "presence_penalty": config.PresencePenalty,
        "frequency_penalty": config.FrequencyPenalty,
        "top_p":            config.TopP,
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return "", fmt.Errorf("error marshaling request: %v", err)
    }

    req, err := http.NewRequest("POST", pc.proxyURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return "", fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", pc.password)

    resp, err := pc.client.Do(req)
    if err != nil {
        return "", fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("error decoding response: %v", err)
    }

    return extractResponse(result)
}

func (pc *ProxyClient) SwitchModel(userID, model string) {
    pc.mu.Lock()
    defer pc.mu.Unlock()
    
    config := pc.getUserConfig(userID)
    config.Model = model
}

func (pc *ProxyClient) SetTemperature(userID string, temp float64) {
    pc.mu.Lock()
    defer pc.mu.Unlock()
    
    config := pc.getUserConfig(userID)
    config.Temperature = temp
}

func (pc *ProxyClient) ToggleStream(userID string) bool {
    pc.mu.Lock()
    defer pc.mu.Unlock()
    
    config := pc.getUserConfig(userID)
    config.Stream = !config.Stream
    return config.Stream
}

func (pc *ProxyClient) getUserConfig(userID string) *UserConfig {
    config, exists := pc.userConfigs[userID]
    if !exists {
        config = &UserConfig{
            Model:            "gpt-4-turbo",
            Temperature:      0.83,
            MaxTokens:        1096,
            Stream:           true,
            PresencePenalty:  0,
            FrequencyPenalty: 0.6,
            TopP:             0.99,
        }
        pc.userConfigs[userID] = config
    }
    return config
}

func extractResponse(result map[string]interface{}) (string, error) {
    choices, ok := result["choices"].([]interface{})
    if !ok || len(choices) == 0 {
        return "", fmt.Errorf("invalid response format")
    }

    choice, ok := choices[0].(map[string]interface{})
    if !ok {
        return "", fmt.Errorf("invalid choice format")
    }

    message, ok := choice["message"].(map[string]interface{})
    if !ok {
        return "", fmt.Errorf("invalid message format")
    }

    content, ok := message["content"].(string)
    if !ok {
        return "", fmt.Errorf("invalid content format")
    }

    return content, nil
}
