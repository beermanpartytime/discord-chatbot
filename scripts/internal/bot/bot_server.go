package bot

import (
    "log"
    "sync"
    "github.com/bwmarrin/discordgo"
    "your-module/internal/services"
)

type Server struct {
    discord       *discordgo.Session
    promptManager *services.PromptManager
    chatManager   *services.ChatManager
    proxyClient   *services.ProxyClient
    commands      *CommandHandler
    events        *EventHandler
    mu            sync.RWMutex
}

func NewServer(
    discord *discordgo.Session,
    pm *services.PromptManager,
    cm *services.ChatManager,
    pc *services.ProxyClient,
) *Server {
    server := &Server{
        discord:       discord,
        promptManager: pm,
        chatManager:   cm,
        proxyClient:   pc,
    }

    // Initialize handlers
    server.commands = NewCommandHandler(discord, pm, cm, pc)
    server.events = NewEventHandler(discord, pm, cm, pc)

    return server
}

func (s *Server) Start() error {
    // Wait for Discord session to be ready before registering commands
    s.discord.AddHandler(func(session *discordgo.Session, r *discordgo.Ready) {
        // Register commands after bot is ready
        s.commands.RegisterCommands()
    })

    // Set initial presence
    err := s.discord.UpdateGameStatus(0, "Ready to chat! 💬")
    if err != nil {
        return err
    }

    // Register event handlers
    s.events.RegisterHandlers()

    // Add ready event handler
    s.discord.AddHandler(s.handleReady)

    log.Println("Bot server initialized and ready")
    return nil
}


func (s *Server) Stop() error {
    // Cleanup and shutdown logic
    s.mu.Lock()
    defer s.mu.Unlock()

    // Save any pending data
    s.chatManager.SaveAllSessions()

    // Remove commands on shutdown
    for _, cmd := range commands {
        err := s.discord.ApplicationCommandDelete(s.discord.State.User.ID, "", cmd.ID)
        if err != nil {
            log.Printf("Error removing command %v: %v", cmd.Name, err)
        }
    }

    return nil
}

func (s *Server) handleReady(session *discordgo.Session, event *discordgo.Ready) {
    log.Printf("Logged in as: %v#%v", event.User.Username, event.User.Discriminator)
    log.Printf("Connected to %d guilds", len(event.Guilds))
}

func (s *Server) HandleInteractionCreate(session *discordgo.Session, i *discordgo.InteractionCreate) {
    s.commands.HandleInteractionCreate(session, i)
}

func (s *Server) HandleMessageCreate(session *discordgo.Session, m *discordgo.MessageCreate) {
    s.events.HandleMessageCreate(session, m)
}

// Utility methods for server management
func (s *Server) GetGuildCount() int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return len(s.discord.State.Guilds)
}

func (s *Server) GetActiveUsers() int {
    return s.chatManager.GetActiveUserCount()
}

func (s *Server) IsUserInCooldown(userID string) bool {
    return s.chatManager.IsUserInCooldown(userID)
}

func (s *Server) GetServerStats() map[string]interface{} {
    s.mu.RLock()
    defer s.mu.RUnlock()

    return map[string]interface{}{
        "guilds":       len(s.discord.State.Guilds),
        "active_users": s.GetActiveUsers(),
        "uptime":      s.chatManager.GetUptime(),
        "total_chats": s.chatManager.GetTotalChats(),
    }
}

func (s *Server) RegisterCommands() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Register commands through the command handler
    s.commands.RegisterCommands()
    return nil
}
