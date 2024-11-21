package bot

import (
    "strings"
    "github.com/bwmarrin/discordgo"
    "your-module/internal/services"
)

type EventHandler struct {
    discord       *discordgo.Session
    promptManager *services.PromptManager
    chatManager   *services.ChatManager
    proxyClient   *services.ProxyClient
}

func NewEventHandler(d *discordgo.Session, pm *services.PromptManager, cm *services.ChatManager, pc *services.ProxyClient) *EventHandler {
    return &EventHandler{
        discord:       d,
        promptManager: pm,
        chatManager:   cm,
        proxyClient:   pc,
    }
}

func (h *EventHandler) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if m.Author.ID == s.State.User.ID {
        return
    }

    if m.ReferencedMessage != nil && m.ReferencedMessage.Author.ID == s.State.User.ID {
        h.handleBotReply(s, m)
        return
    }

    for _, mention := range m.Mentions {
        if mention.ID == s.State.User.ID {
            h.handleBotMention(s, m)
            return
        }
    }
}

func (h *EventHandler) HandleMessageEdit(s *discordgo.Session, m *discordgo.MessageUpdate) {
    if m.Author == nil || m.Author.ID == s.State.User.ID {
        return
    }

    if h.chatManager.MessageExists(m.ID) {
        h.chatManager.UpdateMessage(m.ID, m.Content)
        
        s.ChannelTyping(m.ChannelID)
        
        response, err := h.chatManager.GenerateResponse(m.Author.ID)
        if err != nil {
            h.sendErrorResponse(s, m.ChannelID, err)
            return
        }
        h.sendResponse(s, m.ChannelID, response)
    }
}

func (h *EventHandler) HandleMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
    h.chatManager.DeleteMessage(m.ID)
}

func (h *EventHandler) handleBotReply(s *discordgo.Session, m *discordgo.MessageCreate) {
    s.ChannelTyping(m.ChannelID)
    h.chatManager.AddMessage(m.Author.ID, "user", m.Content)
    
    response, err := h.chatManager.GenerateResponse(m.Author.ID)
    if err != nil {
        h.sendErrorResponse(s, m.ChannelID, err)
        return
    }
    h.sendResponse(s, m.ChannelID, response)
}

func (h *EventHandler) handleBotMention(s *discordgo.Session, m *discordgo.MessageCreate) {
    content := strings.TrimSpace(strings.ReplaceAll(
        m.Content, 
        "<@"+s.State.User.ID+">", 
        "",
    ))

    s.ChannelTyping(m.ChannelID)
    h.chatManager.AddMessage(m.Author.ID, "user", content)
    
    response, err := h.chatManager.GenerateResponse(m.Author.ID)
    if err != nil {
        h.sendErrorResponse(s, m.ChannelID, err)
        return
    }
    h.sendResponse(s, m.ChannelID, response)
}

func (h *EventHandler) sendResponse(s *discordgo.Session, channelID string, content string) {
    msg, err := s.ChannelMessageSend(channelID, content)
    if err != nil {
        return
    }
    
    h.chatManager.AddMessage(msg.Author.ID, "assistant", content)
}

func (h *EventHandler) sendErrorResponse(s *discordgo.Session, channelID string, err error) {
    s.ChannelMessageSend(channelID, "An error occurred while processing your request.")
}

func (h *EventHandler) RegisterHandlers() {
    h.discord.AddHandler(h.HandleMessageCreate)
    h.discord.AddHandler(h.HandleMessageEdit)
    h.discord.AddHandler(h.HandleMessageDelete)
}
