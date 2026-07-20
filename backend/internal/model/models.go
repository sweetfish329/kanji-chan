package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 幹事・管理者
type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	OAuthProvider string    `gorm:"column:oauth_provider;type:varchar(50);not null;default:'';uniqueIndex:idx_provider_id" json:"oauth_provider"`
	OAuthID       string    `gorm:"column:oauth_id;type:varchar(255);not null;default:'';uniqueIndex:idx_provider_id" json:"oauth_id"`
	Email         string    `gorm:"type:varchar(255);not null" json:"email"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Role          string    `gorm:"type:varchar(50);default:'user'" json:"role"`
	GeminiAPIKey  string    `gorm:"type:varchar(255)" json:"gemini_api_key,omitempty"` // 暗号化して保存するか、まずは平文（デモ用）で
	CreatedAt     time.Time `json:"created_at"`
	Events        []Event   `gorm:"foreignKey:CreatedBy" json:"events,omitempty"`
	ApiKeys       []ApiKey  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"api_keys,omitempty"`
}

// ApiKey 幹事ちゃん API キー (MCPおよび外部API連携用)
type ApiKey struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	Name       string     `gorm:"type:varchar(255);not null" json:"name"`
	Role       string     `gorm:"type:varchar(50);default:'user'" json:"role"`
	KeyPrefix  string     `gorm:"type:varchar(20);not null" json:"key_prefix"`    // 例: "kc_8f3a9b..." (UI表示用)
	KeyHash    string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"-"` // SHA-256
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// Event 調整イベント
type Event struct {
	ID                   string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Title                string    `gorm:"type:varchar(255);not null" json:"title"`
	Description          string    `gorm:"type:text" json:"description"`
	CreatedBy            *uint     `json:"created_by,omitempty"`
	Status               string    `gorm:"type:varchar(50);default:'scheduling'" json:"status"` // 'scheduling' or 'confirmed'
	ConfirmedCandidateID *uint     `json:"confirmed_candidate_id,omitempty"`
	CreatedAt            time.Time `json:"created_at"`

	// リレーション
	Candidates         []EventCandidate `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE" json:"candidates"`
	Responses          []Response       `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE" json:"responses,omitempty"`
	ConfirmedCandidate *EventCandidate  `gorm:"foreignKey:ConfirmedCandidateID" json:"confirmed_candidate,omitempty"`
}

// BeforeCreate GORMフック: レコード作成前にUUIDを自動生成する
func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// EventCandidate イベントの候補日時
type EventCandidate struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	EventID   string `gorm:"type:varchar(36);not null;index" json:"event_id"`
	EventDate string `gorm:"type:date;not null" json:"event_date"` // YYYY-MM-DD
	StartTime string `gorm:"type:time;not null" json:"start_time"` // HH:MM
	EndTime   string `gorm:"type:time;not null" json:"end_time"`   // HH:MM
}

// Response 回答
type Response struct {
	ID             uint              `gorm:"primaryKey" json:"id"`
	EventID        string            `gorm:"type:varchar(36);not null;index" json:"event_id"`
	RespondentName string            `gorm:"type:varchar(255);not null" json:"respondent_name"`
	Comment        string            `gorm:"type:text" json:"comment"`
	EditToken      string            `gorm:"type:varchar(255);not null" json:"edit_token"`
	CreatedAt      time.Time         `json:"created_at"`
	Answers        []CandidateAnswer `gorm:"foreignKey:ResponseID;constraint:OnDelete:CASCADE" json:"answers"`
}

// CandidateAnswer 候補日に対する回答 (〇=ok, △=maybe, ×=ng)
type CandidateAnswer struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	ResponseID   uint   `gorm:"not null;uniqueIndex:idx_response_candidate" json:"response_id"`
	CandidateID  uint   `gorm:"not null;uniqueIndex:idx_response_candidate;index" json:"candidate_id"`
	AnswerStatus string `gorm:"type:varchar(10);not null" json:"answer_status"` // 'ok', 'maybe', 'ng'
}
