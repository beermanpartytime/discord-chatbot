package bot

import (
    "fmt"
    "log"
    "time"
    "strings"
    "github.com/bwmarrin/discordgo"
    "your-module/internal/services"
)

type CommandHandler struct {
    discord       *discordgo.Session
    promptManager *services.PromptManager
    chatManager   *services.ChatManager
    proxyClient   *services.ProxyClient
}

func NewCommandHandler(d *discordgo.Session, pm *services.PromptManager, cm *services.ChatManager, pc *services.ProxyClient) *CommandHandler {
    return &CommandHandler{
        discord:       d,
        promptManager: pm,
        chatManager:   cm,
        proxyClient:   pc,
    }
}

var commands = []*discordgo.ApplicationCommand{
    {
        Name: "new-chat",
        Description: "Start a new chat session",
    },
    {
        Name: "regenerate",
        Description: "Regenerate the last bot response",
    },
    {
        Name: "continue",
        Description: "Continue from the last message",
    },
    {
        Name: "set-definitions",
        Description: "Set chatbot personality definitions",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "description",
                Description: "Bot's description",
                Required:    true,
            },
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "personality",
                Description: "Bot's personality traits",
                Required:    true,
            },
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "scenario",
                Description: "Current scenario/setting",
                Required:    false,
            },
        },
    },
    {
        Name: "set-userpersona",
        Description: "Set your character persona",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "persona",
                Description: "Your character description",
                Required:    false,
            },
        },
    },
    {
        Name: "set-usertoken",
        Description: "Set your user token for the chat",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "token",
                Description: "Your unique token",
                Required:    true,
            },
        },
    },
    {
        Name: "save-chat",
        Description: "Save current chat history",
    },
    {
        Name: "load-chat",
        Description: "Load a saved chat history",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "chat-id",
                Description: "ID of saved chat",
                Required:    true,
            },
        },
    },
    {
        Name: "toggle-stream",
        Description: "Toggle streaming mode for responses",
    },
    {
        Name: "set-temperature",
        Description: "Set AI response randomness (0.0-2.0)",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionNumber,
                Name:        "value",
                Description: "Temperature value",
                Required:    true,
                MinValue:    &[]float64{0.0}[0],
                MaxValue:    2.0,
            },
        },
    },
    {
        Name: "help",
        Description: "Show available commands and their usage",
    },
    {
        Name: "stats",
        Description: "Show chat statistics and token usage",
    },
    {
        Name: "undo",
        Description: "Remove the last message from chat history",
    },
    {
        Name: "export-chat",
        Description: "Export current chat history as a text file",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "format",
                Description: "Export format (txt/json)",
                Required:    false,
                Choices: []*discordgo.ApplicationCommandOptionChoice{
                    {Name: "Text", Value: "txt"},
                    {Name: "JSON", Value: "json"},
                },
            },
        },
    },
    {
        Name: "set-first-message",
        Description: "Set custom greeting message for new chats",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "message",
                Description: "Custom greeting message",
                Required:    true,
            },
        },
    },
    {
        Name: "set-authors-note",
        Description: "Set additional context notes for the chat",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "note",
                Description: "Author's note content",
                Required:    true,
            },
        },
    },
    {
        Name: "memory",
        Description: "View current context and memory usage statistics",
    },
    {
        Name: "clear-memory",
        Description: "Clear chat context while preserving personality settings",
    },
    {
        Name: "ping",
        Description: "Check bot latency and status",
    },
    {
        Name: "backup",
        Description: "Create backup of all user settings and chat data",
    },
    {
        Name: "restore",
        Description: "Restore settings and chat data from backup",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "backup-id",
                Description: "Backup identifier to restore from",
                Required:    true,
            },
        },
    },
    {
        Name: "switch-model",
        Description: "Switch between different AI models",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "model",
                Description: "AI model to use",
                Required:    true,
                Choices: []*discordgo.ApplicationCommandOptionChoice{
                    {Name: "GPT-4 Turbo", Value: "gpt-4-turbo"},
                    {Name: "GPT-3.5 Turbo", Value: "gpt-3.5-turbo"},
                },
            },
        },
    },
}

func (h *CommandHandler) RegisterCommands() {
    for _, cmd := range commands {
        _, err := h.discord.ApplicationCommandCreate(h.discord.State.User.ID, "", cmd)
        if err != nil {
            log.Printf("Error creating command %v: %v", cmd.Name, err)
        }
    }
}

func (h *CommandHandler) HandleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
    if i.Type != discordgo.InteractionApplicationCommand {
        return
    }

    commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
        "new-chat":          h.handleNewChat,
        "regenerate":        h.handleRegenerate,
        "continue":          h.handleContinue,
        "set-definitions":   h.handleSetDefinitions,
        "set-userpersona":   h.handleSetUserPersona,
        "set-usertoken":     h.handleSetUserToken,
        "save-chat":         h.handleSaveChat,
        "load-chat":         h.handleLoadChat,
        "toggle-stream":     h.handleToggleStream,
        "set-temperature":   h.handleSetTemperature,
        "help":             h.handleHelp,
        "stats":            h.handleStats,
        "undo":             h.handleUndo,
        "export-chat":      h.handleExportChat,
        "set-first-message": h.handleSetFirstMessage,
        "set-authors-note":  h.handleSetAuthorsNote,
        "memory":           h.handleMemory,
        "clear-memory":     h.handleClearMemory,
        "ping":             h.handlePing,
        "backup":           h.handleBackup,
        "restore":          h.handleRestore,
        "switch-model":     h.handleSwitchModel,
    }

    if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
        handler(s, i)
    }
}

func (h *CommandHandler) handleNewChat(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    h.chatManager.CreateNewChat(userID)
    
    response := "New chat session started! 🌟"
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: response,
        },
    })
}

func (h *CommandHandler) handleRegenerate(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    
    // Respond with "thinking" message
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "🔄 Regenerating last response...",
        },
    })

    response := h.chatManager.RegenerateLastResponse(userID)
    
    // Edit the response with the new content
    s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Content: &response,
    })
}

func (h *CommandHandler) handleSetDefinitions(s *discordgo.Session, i *discordgo.InteractionCreate) {
    options := i.ApplicationCommandData().Options
    definitions := make(map[string]string)
    
    for _, opt := range options {
        definitions[opt.Name] = opt.StringValue()
    }
    
    userID := i.Member.User.ID
    h.promptManager.UpdateDefinitions(userID, definitions)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "✅ Bot definitions updated successfully!",
        },
    })
}

func (h *CommandHandler) handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
    latency := s.HeartbeatLatency().Milliseconds()
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("🏓 Pong! Latency: %dms", latency),
        },
    })
}

func (h *CommandHandler) handleMemory(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    stats := h.chatManager.GetMemoryStats(userID)
    
    response := fmt.Sprintf("📊 Memory Usage:\nTokens: %d/%d\nMessages: %d\nContext Size: %.2f KB", 
        stats.UsedTokens, stats.MaxTokens, stats.MessageCount, stats.ContextSize)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: response,
        },
    })
}

func (h *CommandHandler) handleSwitchModel(s *discordgo.Session, i *discordgo.InteractionCreate) {
    model := i.ApplicationCommandData().Options[0].StringValue()
    userID := i.Member.User.ID
    
    h.proxyClient.SwitchModel(userID, model)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("🔄 Switched to %s model", model),
        },
    })
}

func (h *CommandHandler) handleContinue(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
    })

    response := h.chatManager.ContinueChat(userID)
    s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Content: &response,
    })
}

func (h *CommandHandler) handleSetUserPersona(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    persona := ""
    if len(i.ApplicationCommandData().Options) > 0 {
        persona = i.ApplicationCommandData().Options[0].StringValue()
    }
    
    h.promptManager.SetUserPersona(userID, persona)
    
    response := "User persona updated! ✨"
    if persona == "" {
        response = "User persona cleared! 🔄"
    }
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: response,
        },
    })
}

func (h *CommandHandler) handleSetUserToken(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    token := i.ApplicationCommandData().Options[0].StringValue()
    
    h.promptManager.SetUserToken(userID, token)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "🔑 User token set successfully!",
        },
    })
}

func (h *CommandHandler) handleSaveChat(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    chatID := h.chatManager.SaveChat(userID)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("💾 Chat saved! ID: `%s`", chatID),
        },
    })
}

func (h *CommandHandler) handleLoadChat(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    chatID := i.ApplicationCommandData().Options[0].StringValue()
    
    if err := h.chatManager.LoadChat(userID, chatID); err != nil {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "❌ Chat not found or error loading chat",
            },
        })
        return
    }
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "📂 Chat loaded successfully!",
        },
    })
}

func (h *CommandHandler) handleToggleStream(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    isEnabled := h.proxyClient.ToggleStream(userID)
    
    response := "Stream mode enabled! 📺"
    if !isEnabled {
        response = "Stream mode disabled! 📴"
    }
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: response,
        },
    })
}

func (h *CommandHandler) handleSetTemperature(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    temp := i.ApplicationCommandData().Options[0].FloatValue()
    
    h.proxyClient.SetTemperature(userID, temp)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("🌡️ Temperature set to %.2f", temp),
        },
    })
}

func (h *CommandHandler) handleUndo(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    if removed := h.chatManager.RemoveLastMessage(userID); removed {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "↩️ Last message removed from history",
            },
        })
    } else {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "❌ No messages to remove",
            },
        })
    }
}

func (h *CommandHandler) handleExportChat(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    format := "txt"
    if len(i.ApplicationCommandData().Options) > 0 {
        format = i.ApplicationCommandData().Options[0].StringValue()
    }
    
    exportData := h.chatManager.ExportChat(userID, format)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "📤 Here's your chat export:",
            Files: []*discordgo.File{
                {
                    Name:        fmt.Sprintf("chat_export.%s", format),
                    ContentType: "text/plain",
                    Reader:      strings.NewReader(exportData),
                },
            },
        },
    })
}

func (h *CommandHandler) handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
    helpText := "Available Commands:\n" +
        "`/new-chat` - Start a new chat session\n" +
        "`/regenerate` - Regenerate last response\n" +
        "`/continue` - Continue from last message\n" +
        "`/set-definitions` - Set bot personality\n" +
        "`/set-userpersona` - Set your character\n" +
        "`/toggle-stream` - Toggle streaming mode\n" +
        "`/save-chat` - Save current chat\n" +
        "`/load-chat` - Load saved chat"

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: helpText,
        },
    })
}

func (h *CommandHandler) handleStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    stats := h.chatManager.GetChatHistory(userID)
    
    response := fmt.Sprintf("Chat Statistics:\nMessages: %d\n", len(stats))
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: response,
        },
    })
}

func (h *CommandHandler) handleSetFirstMessage(s *discordgo.Session, i *discordgo.InteractionCreate) {
    message := i.ApplicationCommandData().Options[0].StringValue()
    userID := i.Member.User.ID
    
    h.promptManager.SetFirstMessage(userID, message)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "First message updated! ✨",
        },
    })
}

func (h *CommandHandler) handleSetAuthorsNote(s *discordgo.Session, i *discordgo.InteractionCreate) {
    note := i.ApplicationCommandData().Options[0].StringValue()
    userID := i.Member.User.ID
    
    h.promptManager.SetAuthorsNote(userID, note)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "Author's note updated! 📝",
        },
    })
}

func (h *CommandHandler) handleClearMemory(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID := i.Member.User.ID
    h.chatManager.ClearChat(userID)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "Chat memory cleared! 🧹",
        },
    })
}

func (h *CommandHandler) handleBackup(s *discordgo.Session, i *discordgo.InteractionCreate) {
    backupID := time.Now().Format("20060102150405")
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("Backup created! ID: %s 💾", backupID),
        },
    })
}


func (h *CommandHandler) handleRestore(s *discordgo.Session, i *discordgo.InteractionCreate) {
    backupID := i.ApplicationCommandData().Options[0].StringValue()
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("Restored from backup: %s 📥", backupID),
        },
    })
}



