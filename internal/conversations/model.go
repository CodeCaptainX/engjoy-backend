package conversations

import (
	"fmt"
	"strings"
	"time"

	"sentenceminer/pkg/model"
)

// UserConversation represents a record in the tbl_users_conversations table.
type UserConversation struct {
	ID           int64      `db:"id" json:"id"`
	UUID         string     `db:"uuid" json:"uuid"`
	Title        *string    `db:"title" json:"title,omitempty"`
	Source       *string    `db:"source" json:"source,omitempty"`
	Category     *string    `db:"category" json:"category,omitempty"`
	UserID       int64      `db:"user_id" json:"userId"` // Owner of the conversation
	IsPublic     bool       `db:"is_public" json:"isPublic"` // New field for public visibility
	StatusID     int        `db:"status_id" json:"statusId"`
	Order        int        `db:"order" json:"order"`
	CreatedBy    int64      `db:"created_by" json:"createdBy"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedBy    int64      `db:"updated_by" json:"updatedBy"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedBy    *int64     `db:"deleted_by" json:"deletedBy,omitempty"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
}

// UserConversationMessage represents a record in the tbl_users_conversations_messages table.
type UserConversationMessage struct {
	ID            int64      `db:"id" json:"id"`
	UUID          string     `db:"uuid" json:"uuid"`
	ConversationID int64      `db:"conversation_id" json:"conversationId"`
	Speaker       string     `db:"speaker" json:"speaker"`
	MessageText   string     `db:"message_text" json:"messageText"`
	MessageOrder  int        `db:"message_order" json:"messageOrder"`
	StatusID      int        `db:"status_id" json:"statusId"`
	CreatedBy     int64      `db:"created_by" json:"createdBy"`
	CreatedAt     time.Time  `db:"created_at" json:"createdAt"`
	UpdatedBy     int64      `db:"updated_by" json:"updatedBy"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedBy     *int64     `db:"deleted_by" json:"deletedBy,omitempty"`
	DeletedAt     *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
}

// CreateUserConversationRequest defines the structure for creating a new user conversation.
type CreateUserConversationRequest struct {
	Title    string `json:"title,omitempty"`
	Source   string `json:"source,omitempty"`
	Category string `json:"category,omitempty"`
	IsPublic *bool  `json:"isPublic,omitempty"` // New field for public visibility
}

// Bind performs validation for CreateUserConversationRequest.
func (r *CreateUserConversationRequest) Bind() error {
	if strings.TrimSpace(r.Title) == "" && strings.TrimSpace(r.Source) == "" && strings.TrimSpace(r.Category) == "" {
		return fmt.Errorf("at least one of title, source, or category must be provided")
	}
	return nil
}

// UpdateUserConversationRequest defines the structure for updating an existing user conversation.
type UpdateUserConversationRequest struct {
	Title    *string `json:"title,omitempty"`
	Source   *string `json:"source,omitempty"`
	Category *string `json:"category,omitempty"`
	IsPublic *bool  `json:"isPublic,omitempty"` // New field for public visibility
}

// Bind performs validation for UpdateUserConversationRequest.
func (r *UpdateUserConversationRequest) Bind() error {
	if r.Title == nil && r.Source == nil && r.Category == nil && r.IsPublic == nil { // Added IsPublic to check
		return fmt.Errorf("at least one field (title, source, category, or isPublic) must be provided for update")
	}
	return nil
}

// UserConversationsShowRequest defines the structure for filtering and paging user conversations.
type UserConversationsShowRequest struct {
	PagingOptions model.Paging   `json:"pagingOptions" query:"paging_options"`
	Filters       []model.Filter `json:"filters" query:"filters"`
	Sorts         []model.Sort   `json:"sorts" query:"sorts"`
	Offset        int            `json:"offset" query:"offset"`
	IsPublic      *bool          `query:"is_public"` // New filter for public visibility
}

// CreateUserConversationMessageRequest defines the structure for creating a new message within a conversation.
type CreateUserConversationMessageRequest struct {
	Speaker     string `json:"speaker"`
	MessageText string `json:"messageText"`
	MessageOrder int   `json:"messageOrder"`
}

// Bind performs validation for CreateUserConversationMessageRequest.
func (r *CreateUserConversationMessageRequest) Bind() error {
	if strings.TrimSpace(r.Speaker) == "" {
		return fmt.Errorf("speaker cannot be empty")
	}
	if strings.TrimSpace(r.MessageText) == "" {
		return fmt.Errorf("message text cannot be empty")
	}
	if r.MessageOrder <= 0 {
		return fmt.Errorf("message order must be a positive integer")
	}
	return nil
}

// CreateMultipleUserConversationMessagesRequest defines the structure for creating multiple messages at once.
type CreateMultipleUserConversationMessagesRequest struct {
	Messages []CreateUserConversationMessageRequest `json:"messages"`
}

// Bind performs validation for CreateMultipleUserConversationMessagesRequest.
func (r *CreateMultipleUserConversationMessagesRequest) Bind() error {
	if len(r.Messages) == 0 {
		return fmt.Errorf("no messages provided")
	}
	for i, msg := range r.Messages {
		if err := msg.Bind(); err != nil {
			return fmt.Errorf("message at index %d: %w", i, err)
		}
	}
	return nil
}

// UpdateUserConversationMessageRequest defines the structure for updating an existing message.
type UpdateUserConversationMessageRequest struct {
	Speaker     *string `json:"speaker,omitempty"`
	MessageText *string `json:"messageText,omitempty"`
	MessageOrder *int   `json:"messageOrder,omitempty"`
}

// Bind performs validation for UpdateUserConversationMessageRequest.
func (r *UpdateUserConversationMessageRequest) Bind() error {
	if r.Speaker == nil && r.MessageText == nil && r.MessageOrder == nil {
		return fmt.Errorf("at least one field (speaker, messageText, or messageOrder) must be provided for update")
	}
	if r.MessageOrder != nil && *r.MessageOrder <= 0 {
		return fmt.Errorf("message order must be a positive integer")
	}
	return nil
}
