package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "your-module/internal/bot"
    "your-module/internal/config"
    "your-module/internal/services"
    "github.com/bwmarrin/discordgo"
)

func main() {
    // Load configuration
    cfg := config.Load()

    // Initialize services
    openAI := services.NewOpenAIService(cfg.OpenAIKey)
    promptManager := services.NewPromptManager()
    chatManager := services.NewChatManager(openAI, promptManager)
    proxyClient := services.NewProxyClient(cfg.ProxyURL, cfg.ProxyPassword)

    // Create Discord session
    discord, err := discordgo.New("Bot " + cfg.DiscordToken)
    if err != nil {
        log.Fatal("Error creating Discord session:", err)
    }

    // Add required intents
    discord.Identify.Intents = discordgo.IntentsGuildMessages | 
                              discordgo.IntentsGuildMessageReactions | 
                              discordgo.IntentsDirectMessages |
                              discordgo.IntentsMessageContent

    // Open connection to Discord first
    err = discord.Open()
    if err != nil {
        log.Fatal("Error opening connection:", err)
    }
    defer discord.Close()

    // Initialize bot server
    botServer := bot.NewServer(discord, promptManager, chatManager, proxyClient)

    // Add the interaction handler
    discord.AddHandler(botServer.HandleInteractionCreate)

    // Start the server (which will register commands when ready)
    err = botServer.Start()
    if err != nil {
        log.Fatal("Error starting bot server:", err)
    }

    // Register commands
    botServer.RegisterCommands()
    
    // Register event handlers
    discord.AddHandler(botServer.HandleMessageCreate)
    discord.AddHandler(botServer.HandleInteractionCreate)

    // Open connection to Discord
    err = discord.Open()
    if err != nil {
        log.Fatal("Error opening connection:", err)
    }
    defer discord.Close()

    // Wait for interrupt signal
    log.Println("Bot is running. Press CTRL-C to exit.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
    <-sc

    log.Println("Gracefully shutting down...")
}
