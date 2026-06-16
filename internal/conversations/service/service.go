package service

import (
	"fmt"
	"sentenceminer/internal/conversations/model"
	"sentenceminer/internal/conversations/repository"

	"github.com/jmoiron/sqlx"
)

type ConversationService struct {
	repo *repository.ConversationRepository
}

func NewConversationService(db *sqlx.DB) *ConversationService {
	return &ConversationService{
		repo: repository.NewConversationRepository(db),
	}
}

// User Conversations Service Methods

func (s *ConversationService) CreateUserConversation(req model.CreateUserConversationRequest, userID int64) (*model.UserConversation, error) {
	conversation := model.UserConversation{
		UserID:    userID,
		CreatedBy: userID,
		UpdatedBy: userID,
		IsPublic:  false, // Default to private unless explicitly set
	}

	if req.Title != "" {
		conversation.Title = &req.Title
	}
	if req.Source != "" {
		conversation.Source = &req.Source
	}
	if req.Category != "" {
		conversation.Category = &req.Category
	}
	if req.IsPublic != nil {
		conversation.IsPublic = *req.IsPublic
	}

	createdConversation, err := s.repo.CreateUserConversation(conversation)
	if err != nil {
		return nil, fmt.Errorf("service: failed to create user conversation: %w", err)
	}

	return createdConversation, nil
}

func (s *ConversationService) ListUserConversations(req model.UserConversationsShowRequest, requestingUserID int64) ([]model.UserConversation, int, error) {
	// If req.IsPublic is true, we are requesting public conversations.
	// In this case, the requestingUserID might be 0 (anonymous) or just browsing public content.
	// The repository handles filtering by is_public=true or user_id=:user_id based on req.IsPublic.
	conversations, total, err := s.repo.ListUserConversations(req, requestingUserID)
	if err != nil {
		return nil, 0, fmt.Errorf("service: failed to list user conversations: %w", err)
	}
	return conversations, total, nil
}

func (s *ConversationService) GetUserConversation(id int64, requestingUserID int64) (*model.UserConversation, error) {
	// The repository handles the logic of allowing access to own conversations or public ones.
	conversation, err := s.repo.GetUserConversationByID(id, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get user conversation: %w", err)
	}
	return conversation, nil
}

func (s *ConversationService) UpdateUserConversation(id int64, req model.UpdateUserConversationRequest, userID int64) (*model.UserConversation, error) {
	// Ownership check is done implicitly by the repository's WHERE clause (user_id = :user_id)
	conversation, err := s.repo.UpdateUserConversation(id, req, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to update user conversation: %w", err)
	}
	return conversation, nil
}

func (s *ConversationService) SoftDeleteUserConversation(id int64, userID int64) error {
	// Ownership check is done implicitly by the repository's WHERE clause (user_id = :user_id)
	_, err := s.repo.SoftDeleteUserConversation(id, userID)
	if err != nil {
		return fmt.Errorf("service: failed to soft delete user conversation: %w", err)
	}
	return nil
}

// User Conversation Messages Service Methods

func (s *ConversationService) AddMessageToUserConversation(conversationID int64, req model.CreateUserConversationMessageRequest, userID int64) (*model.UserConversationMessage, error) {
	// Verify that the conversation exists and belongs to the user (only owner can add messages)
	_, err := s.repo.GetUserConversationByID(conversationID, userID) // Pass userID to ensure ownership
	if err != nil {
		return nil, fmt.Errorf("service: conversation not found or does not belong to user: %w", err)
	}

	message := model.UserConversationMessage{
		ConversationID: conversationID,
		Speaker:        req.Speaker,
		MessageText:    req.MessageText,
		MessageOrder:   req.MessageOrder,
		CreatedBy:      userID,
		UpdatedBy:      userID,
	}

	createdMessage, err := s.repo.CreateUserConversationMessage(message)
	if err != nil {
		return nil, fmt.Errorf("service: failed to add message to user conversation: %w", err)
	}
	return createdMessage, nil
}

func (s *ConversationService) AddMultipleMessagesToUserConversation(conversationID int64, req model.CreateMultipleUserConversationMessagesRequest, userID int64) ([]model.UserConversationMessage, error) {
	// Verify that the conversation exists and belongs to the user (only owner can add messages)
	_, err := s.repo.GetUserConversationByID(conversationID, userID) // Pass userID to ensure ownership
	if err != nil {
		return nil, fmt.Errorf("service: conversation not found or does not belong to user: %w", err)
	}

	messagesToCreate := make([]model.UserConversationMessage, len(req.Messages))
	for i, msgReq := range req.Messages {
		messagesToCreate[i] = model.UserConversationMessage{
			ConversationID: conversationID,
			Speaker:        msgReq.Speaker,
			MessageText:    msgReq.MessageText,
			MessageOrder:   msgReq.MessageOrder,
			CreatedBy:      userID,
			UpdatedBy:      userID,
		}
	}

	createdMessages, err := s.repo.CreateMultipleUserConversationMessages(messagesToCreate)
	if err != nil {
		return nil, fmt.Errorf("service: failed to add multiple messages to user conversation: %w", err)
	}
	return createdMessages, nil
}

func (s *ConversationService) ListUserConversationMessages(conversationID int64, requestingUserID int64) ([]model.UserConversationMessage, error) {
	// The repository handles the logic of allowing access if conversation belongs to user or is public.
	messages, err := s.repo.ListUserConversationMessagesByConversationID(conversationID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to list user conversation messages: %w", err)
	}
	return messages, nil
}

func (s *ConversationService) GetUserConversationMessage(id int64, conversationID int64, requestingUserID int64) (*model.UserConversationMessage, error) {
	// The repository handles the logic of allowing access if message belongs to user's conversation or a public one.
	message, err := s.repo.GetUserConversationMessageByID(id, conversationID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get user conversation message: %w", err)
	}
	return message, nil
}

func (s *ConversationService) UpdateUserConversationMessage(id int64, conversationID int64, req model.UpdateUserConversationMessageRequest, userID int64) (*model.UserConversationMessage, error) {
	// Verify that the conversation exists and belongs to the user (only owner can update messages)
	// The repository's WHERE clause already ensures this for update operations.
	message, err := s.repo.UpdateUserConversationMessage(id, conversationID, req, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to update user conversation message: %w", err)
	}
	return message, nil
}

func (s *ConversationService) SoftDeleteUserConversationMessage(id int64, conversationID int64, userID int64) error {
	// Verify that the conversation exists and belongs to the user (only owner can delete messages)
	// The repository's WHERE clause already ensures this for delete operations.
	_, err := s.repo.SoftDeleteUserConversationMessage(id, conversationID, userID)
	if err != nil {
		return fmt.Errorf("service: failed to soft delete user conversation message: %w", err)
	}
	return nil
}
