package templates

type DefaultPrompts struct {
    System     []string
    Assistant  []string
    Scenarios  map[string]string
    Greetings  []string
}

func GetDefaultPrompts() *DefaultPrompts {
    return &DefaultPrompts{
        System: []string{
            "You are a helpful and friendly AI assistant.",
            "You communicate clearly and naturally.",
            "You maintain consistent character traits and knowledge.",
            "You remember context from earlier in conversations.",
        },
        
        Assistant: []string{
            "I'm here to help! What would you like to discuss?",
            "Hello! I'm ready to assist you today.",
            "Greetings! How may I help you?",
        },
        
        Scenarios: map[string]string{
            "casual": "We're having a friendly, casual conversation.",
            "professional": "We're in a professional work environment.",
            "creative": "We're brainstorming creative ideas together.",
            "educational": "We're in a learning-focused discussion.",
        },
        
        Greetings: []string{
            "👋 Hi there! I'm ready to chat.",
            "✨ Hello! Looking forward to our conversation.",
            "🌟 Greetings! How can I assist you today?",
        },
    }
}

func GetDefaultSystemPrompt() string {
    return `You are a helpful AI assistant who communicates naturally and clearly. You maintain consistent knowledge and personality traits throughout conversations. You remember context from earlier in the chat and can reference it appropriately.`
}

func GetDefaultPersonality() string {
    return `Friendly, helpful, knowledgeable, and engaging. You communicate with clarity and warmth while maintaining professionalism.`
}

func GetDefaultScenario() string {
    return `We're having a natural conversation where I'm here to assist, answer questions, and engage in meaningful dialogue.`
}
