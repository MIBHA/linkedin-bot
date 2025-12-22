package storage

import (
	"time"

	"gorm.io/gorm"
)

// ConnectionRequest represents a sent connection request
type ConnectionRequest struct {
	ID          uint      `gorm:"primaryKey"`
	ProfileURL  string    `gorm:"uniqueIndex;not null"`
	ProfileName string    `gorm:"index"`
	Message     string    `gorm:"type:text"`
	Status      string    `gorm:"default:'pending';index"` // pending, accepted, rejected, withdrawn
	SentAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// MessageHistory represents a sent message
type MessageHistory struct {
	ID             uint      `gorm:"primaryKey"`
	RecipientURL   string    `gorm:"index;not null"`
	RecipientName  string    `gorm:"index"`
	MessageText    string    `gorm:"type:text;not null"`
	ConversationID string    `gorm:"index"`
	SentAt         time.Time `gorm:"autoCreateTime"`
}

// SearchHistory tracks search queries performed
type SearchHistory struct {
	ID          uint   `gorm:"primaryKey"`
	Query       string `gorm:"not null"`
	Filters     string `gorm:"type:text"` // JSON string of filters
	ResultCount int
	SearchedAt  time.Time `gorm:"autoCreateTime"`
}

// ProfileVisit tracks profile visits to avoid suspicious patterns
type ProfileVisit struct {
	ID         uint      `gorm:"primaryKey"`
	ProfileURL string    `gorm:"uniqueIndex;not null"`
	VisitCount int       `gorm:"default:1"`
	LastVisit  time.Time `gorm:"autoUpdateTime"`
	FirstVisit time.Time `gorm:"autoCreateTime"`
}

// TableName overrides for custom table names (optional)
func (ConnectionRequest) TableName() string {
	return "connection_requests"
}

func (MessageHistory) TableName() string {
	return "message_history"
}

func (SearchHistory) TableName() string {
	return "search_history"
}

func (ProfileVisit) TableName() string {
	return "profile_visits"
}

// BeforeCreate hook for ConnectionRequest
func (cr *ConnectionRequest) BeforeCreate(tx *gorm.DB) error {
	// Set default status if not provided
	if cr.Status == "" {
		cr.Status = "pending"
	}
	return nil
}
