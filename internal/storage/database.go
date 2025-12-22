package storage

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db *gorm.DB
}

type Config struct {
	DBPath   string
	LogLevel logger.LogLevel
}

func DefaultConfig() *Config {
	return &Config{DBPath: "linkedin_bot.db", LogLevel: logger.Warn}
}

func NewDatabase(config *Config) (*Database, error) {
	if config == nil {
		config = DefaultConfig()
	}
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}
	db, err := gorm.Open(sqlite.Open(config.DBPath), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.AutoMigrate(&ConnectionRequest{}, &MessageHistory{}, &SearchHistory{}, &ProfileVisit{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	fmt.Printf("Database initialized: %s\n", config.DBPath)
	return &Database{db: db}, nil
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

func (d *Database) HasSentConnectionRequest(profileURL string) (bool, error) {
	var count int64
	err := d.db.Model(&ConnectionRequest{}).Where("profile_url = ?", profileURL).Count(&count).Error
	return count > 0, err
}

func (d *Database) CreateConnectionRequest(profileURL, profileName, message string) error {
	req := &ConnectionRequest{ProfileURL: profileURL, ProfileName: profileName, Message: message, Status: "pending"}
	return d.db.Create(req).Error
}

func (d *Database) UpdateConnectionStatus(profileURL, status string) error {
	return d.db.Model(&ConnectionRequest{}).Where("profile_url = ?", profileURL).Update("status", status).Error
}

func (d *Database) GetConnectionRequests(status string, limit int) ([]ConnectionRequest, error) {
	var requests []ConnectionRequest
	query := d.db.Model(&ConnectionRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("sent_at DESC").Find(&requests).Error
	return requests, err
}

func (d *Database) GetConnectionRequestStats() (map[string]int64, error) {
	stats := make(map[string]int64)
	var total, pending, accepted, rejected, today int64
	d.db.Model(&ConnectionRequest{}).Count(&total)
	stats["total"] = total
	d.db.Model(&ConnectionRequest{}).Where("status = ?", "pending").Count(&pending)
	stats["pending"] = pending
	d.db.Model(&ConnectionRequest{}).Where("status = ?", "accepted").Count(&accepted)
	stats["accepted"] = accepted
	d.db.Model(&ConnectionRequest{}).Where("status = ?", "rejected").Count(&rejected)
	stats["rejected"] = rejected
	todayDate := time.Now().UTC().Truncate(24 * time.Hour)
	d.db.Model(&ConnectionRequest{}).Where("sent_at >= ?", todayDate).Count(&today)
	stats["today"] = today
	return stats, nil
}

func (d *Database) HasSentMessage(recipientURL string) (bool, error) {
	var count int64
	err := d.db.Model(&MessageHistory{}).Where("recipient_url = ?", recipientURL).Count(&count).Error
	return count > 0, err
}

func (d *Database) CreateMessage(recipientURL, recipientName, messageText, conversationID string) error {
	msg := &MessageHistory{RecipientURL: recipientURL, RecipientName: recipientName, MessageText: messageText, ConversationID: conversationID}
	return d.db.Create(msg).Error
}

func (d *Database) GetMessages(limit int) ([]MessageHistory, error) {
	var messages []MessageHistory
	query := d.db.Model(&MessageHistory{})
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("sent_at DESC").Find(&messages).Error
	return messages, err
}

func (d *Database) CreateSearchHistory(query, filters string, resultCount int) error {
	search := &SearchHistory{Query: query, Filters: filters, ResultCount: resultCount}
	return d.db.Create(search).Error
}

func (d *Database) RecordProfileVisit(profileURL string) error {
	var visit ProfileVisit
	result := d.db.Where("profile_url = ?", profileURL).First(&visit)
	if result.Error == gorm.ErrRecordNotFound {
		visit = ProfileVisit{ProfileURL: profileURL, VisitCount: 1}
		return d.db.Create(&visit).Error
	} else if result.Error != nil {
		return result.Error
	}
	visit.VisitCount++
	visit.LastVisit = time.Now().UTC()
	return d.db.Save(&visit).Error
}

func (d *Database) GetRecentVisits(hours int) ([]ProfileVisit, error) {
	var visits []ProfileVisit
	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	err := d.db.Where("last_visit >= ?", since).Order("last_visit DESC").Find(&visits).Error
	return visits, err
}

func (d *Database) GetVisitCount(profileURL string) (int, error) {
	var visit ProfileVisit
	err := d.db.Where("profile_url = ?", profileURL).First(&visit).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return visit.VisitCount, err
}
