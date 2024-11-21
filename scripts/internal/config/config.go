package config

import (
    "os"
    "strconv"
    "time"
    "github.com/joho/godotenv"
)

type Config struct {
    // Discord Configuration
    DiscordToken  string
    GuildID       string
    
    // OpenAI Proxy Configuration
    ProxyURL      string
    ProxyPassword string
    OpenAIKey     string
    
    // Bot Settings
    DefaultPrefix    string
    DefaultModel    string
    DefaultTemp     float64
    MaxTokens      int
    
    // Timeouts and Limits
    RequestTimeout  time.Duration
    SessionTimeout time.Duration
    RateLimit      int
    
    // Development Mode
    Debug bool
}

func Load() *Config {
    godotenv.Load()

    return &Config{
        // Discord
        DiscordToken: getEnv("DISCORD_TOKEN"),
        GuildID:      getEnv("GUILD_ID"),
        
        // Proxy
        ProxyURL:      getEnv("PROXY_URL"),
        ProxyPassword: getEnv("PROXY_PASSWORD"),
        
        // Bot Settings
        DefaultPrefix: getEnv("DEFAULT_PREFIX", "/"),
        DefaultModel: getEnv("DEFAULT_MODEL", "chatgpt-4o-latest"),
        DefaultTemp:  getEnvFloat("DEFAULT_TEMPERATURE", 0.83),
        MaxTokens:    getEnvInt("MAX_TOKENS", 1096),
        
        // Timeouts
        RequestTimeout:  time.Duration(getEnvInt("REQUEST_TIMEOUT", 30)) * time.Second,
        SessionTimeout: time.Duration(getEnvInt("SESSION_TIMEOUT", 3600)) * time.Second,
        RateLimit:      getEnvInt("RATE_LIMIT", 60),
        
        // Debug Mode
        Debug: getEnvBool("DEBUG"),
    }
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}

func getEnvInt(key string, fallback int) int {
    if value, exists := os.LookupEnv(key); exists {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
    if value, exists := os.LookupEnv(key); exists {
        if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
            return floatVal
        }
    }
    return fallback
}

func getEnvBool(key string, fallback bool) bool {
    if value, exists := os.LookupEnv(key); exists {
        if boolVal, err := strconv.ParseBool(value); err == nil {
            return boolVal
        }
    }
    return fallback
}
