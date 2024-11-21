package services

import "time"

type Message struct {
    ID        string    `json:"id"`
    Role      string    `json:"role"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
