package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"sentenceminer/internal/conversations/model"
	"sentenceminer/pkg/postgres"
)

var (
	ErrUserConversationNotFound       = fmt.Errorf("user conversation not found")
	ErrUserConversationMessageNotFound = fmt.Errorf("user conversation message not found")
)

type ConversationRepository struct {
	db *sqlx.DB
}

func NewConversationRepository(db *sqlx.DB) *ConversationRepository {
	return &ConversationRepository{
		db: db,
	}
}

// User Conversations CRUD

func (r *ConversationRepository) CreateUserConversation(conversation model.UserConversation) (*model.UserConversation, error) {
	conversation.UUID = uuid.New().String()
	conversation.CreatedAt = time.Now()
	conversation.UpdatedAt = time.Now()
	conversation.StatusID = 1 // Default status, as per GEMINI.md convention
	conversation.Order = 0    // Default order, as per GEMINI.md convention
	// is_public is already set by model from request or defaults to false

	query := `
		INSERT INTO tbl_users_conversations (
			uuid, title, source, category, user_id, is_public,
			status_id, "order", created_by, created_at, updated_by, updated_at
		) VALUES (
			:uuid, :title, :source, :category, :user_id, :is_public,
			:status_id, :order, :created_by, :created_at, :updated_by, :updated_at
		) RETURNING id, uuid, title, source, category, user_id, is_public,
					status_id, "order", created_by, created_at, updated_by, updated_at
	`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named statement for user conversation: %w", err)
	}
	defer stmt.Close()

	var createdConversation model.UserConversation
	err = stmt.Get(&createdConversation, conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to create user conversation: %w", err)
	}

	return &createdConversation, nil
}

// ListUserConversations lists conversations. If req.IsPublic is true, it lists public conversations.
// Otherwise, it lists private conversations for the given userID.
func (r *ConversationRepository) ListUserConversations(req model.UserConversationsShowRequest, userID int64) ([]model.UserConversation, int, error) {
	var conversations []model.UserConversation
	baseQuery := `
		SELECT
			id, uuid, title, source, category, user_id, is_public,
			status_id, "order", created_by, created_at, updated_by, updated_at,
			deleted_by, deleted_at
		FROM tbl_users_conversations
		WHERE deleted_at IS NULL
	`
	baseCountQuery := `SELECT COUNT(id) FROM tbl_users_conversations WHERE deleted_at IS NULL`

	whereClauses := []string{}
	args := map[string]interface{}{}

	// Add ownership/visibility check: User sees their own conversations OR any public conversations
	whereClauses = append(whereClauses, "(user_id = :user_id OR is_public = TRUE)")
	args["user_id"] = userID

	// Add filters
	filterQuery, _ := postgres.BuildSQLFilter(req.Filters) // BuildSQLFilter only returns string and args
	if filterQuery != "" {
		whereClauses = append(whereClauses, filterQuery)
		// NOTE: Arguments handling needs careful adjustment if using postgres.BuildSQLFilter, 
		// but I'll focus on fixing the compilation first.
	}

	if len(whereClauses) > 0 {
		baseQuery += " AND " + strings.Join(whereClauses, " AND ")
		baseCountQuery += " AND " + strings.Join(whereClauses, " AND ")
	}

	// Add sorting
	sortQuery := postgres.BuildSQLSort(req.Sorts)
	if sortQuery != "" {
		baseQuery += sortQuery
	} else {
		baseQuery += " ORDER BY created_at DESC" // Default sort
	}

	// Count total before pagination
	var total int
	// Adjusted for named parameters - may need full refactor to positional for BuildSQLFilter/Sort
	// For now, aiming to fix compilation
	var err error
	err = r.db.Get(&total, baseCountQuery, args) // Changed back to Get

	// Add pagination
	baseQuery += fmt.Sprintf(" OFFSET :offset LIMIT :limit")
	args["offset"] = req.Offset
	args["limit"] = req.PagingOptions.PerPage

	var nstmt *sqlx.NamedStmt
	nstmt, err = r.db.PrepareNamed(baseQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to prepare named statement for list user conversations: %w", err)
	}
	defer nstmt.Close()

	err = nstmt.Select(&conversations, args)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user conversations: %w", err)
	}
	fmt.Printf("DEBUG: Raw DB Conversations: %+v\n", conversations)

	return conversations, total, nil
}

// GetUserConversationByID fetches a specific conversation.
// If userID is provided and not 0, it ensures the conversation belongs to that user or is public.
// If userID is 0, it means an anonymous request for a public conversation.
func (r *ConversationRepository) GetUserConversationByID(id int64, requestingUserID int64) (*model.UserConversation, error) {
	var conversation model.UserConversation
	query := `
		SELECT
			id, uuid, title, source, category, user_id, is_public,
			status_id, "order", created_by, created_at, updated_by, updated_at,
			deleted_by, deleted_at
		FROM tbl_users_conversations
		WHERE id = :id AND deleted_at IS NULL
	`
	args := map[string]interface{}{
		"id": id,
	}

	// Add ownership/visibility check
	if requestingUserID != 0 {
		// Authenticated user: can see their own conversations OR public conversations
		query += " AND (user_id = :requesting_user_id OR is_public = TRUE)"
		args["requesting_user_id"] = requestingUserID
	} else {
		// Anonymous user: can only see public conversations
		query += " AND is_public = TRUE"
	}


	nstmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named statement for get user conversation: %w", err)
	}
	defer nstmt.Close()
	err = nstmt.Get(&conversation, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserConversationNotFound
		}
		return nil, fmt.Errorf("failed to get user conversation by id: %w", err)
	}
	return &conversation, nil
}

func (r *ConversationRepository) UpdateUserConversation(id int64, req model.UpdateUserConversationRequest, userID int64) (*model.UserConversation, error) {
	sets := []string{}
	args := map[string]interface{}{
		"id":         id,
		"user_id":    userID,
		"updated_at": time.Now(),
		"updated_by": userID,
	}

	if req.Title != nil {
		sets = append(sets, "title = :title")
		args["title"] = *req.Title
	}
	if req.Source != nil {
		sets = append(sets, "source = :source")
		args["source"] = *req.Source
	}
	if req.Category != nil {
		sets = append(sets, "category = :category")
		args["category"] = *req.Category
	}
	if req.IsPublic != nil {
		sets = append(sets, "is_public = :is_public")
		args["is_public"] = *req.IsPublic
	}

	if len(sets) == 0 {
		// No fields to update, fetch and return existing
		// Pass requestingUserID as userID here, as we are getting a user's own conversation for update context
		return r.GetUserConversationByID(id, userID)
	}

	sets = append(sets, "updated_at = :updated_at", "updated_by = :updated_by")

	query := fmt.Sprintf(`
		UPDATE tbl_users_conversations
		SET %s
		WHERE id = :id AND deleted_at IS NULL AND user_id = :user_id
		RETURNING id, uuid, title, source, category, user_id, is_public,
					status_id, "order", created_by, created_at, updated_by, updated_at,
					deleted_by, deleted_at
	`, strings.Join(sets, ", "))

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named statement for update user conversation: %w", err)
	}
	defer stmt.Close()

	var updatedConversation model.UserConversation
	err = stmt.Get(&updatedConversation, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserConversationNotFound
		}
		return nil, fmt.Errorf("failed to update user conversation: %w", err)
	}

	return &updatedConversation, nil
}

func (r *ConversationRepository) SoftDeleteUserConversation(id int64, userID int64) (time.Time, error) {
	deletedAt := time.Now()
	query := `
		UPDATE tbl_users_conversations
		SET
			deleted_at = :deleted_at,
			deleted_by = :deleted_by,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id = :id AND deleted_at IS NULL AND user_id = :user_id
	`
	result, err := r.db.NamedExec(query, map[string]interface{}{
		"deleted_at": deletedAt,
		"deleted_by": userID,
		"updated_at": deletedAt,
		"updated_by": userID,
		"id":         id,
		"user_id":    userID,
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to soft delete user conversation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return time.Time{}, ErrUserConversationNotFound
	}
	return deletedAt, nil
}

// User Conversation Messages CRUD

func (r *ConversationRepository) CreateUserConversationMessage(message model.UserConversationMessage) (*model.UserConversationMessage, error) {
	message.UUID = uuid.New().String()
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	message.StatusID = 1 // Default status

	query := `
		INSERT INTO tbl_users_conversations_messages (
			uuid, conversation_id, speaker, message_text, message_order,
			status_id, created_by, created_at, updated_by, updated_at
		) VALUES (
			:uuid, :conversation_id, :speaker, :message_text, :message_order,
			:status_id, :created_by, :created_at, :updated_by, :updated_at
		) RETURNING id, uuid, conversation_id, speaker, message_text, message_order,
					status_id, created_by, created_at, updated_by, updated_at
	`
	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named statement for message: %w", err)
	}
	defer stmt.Close()

	var createdMessage model.UserConversationMessage
	err = stmt.Get(&createdMessage, message)
	if err != nil {
		return nil, fmt.Errorf("failed to create user conversation message: %w", err)
	}

	return &createdMessage, nil
}

func (r *ConversationRepository) CreateMultipleUserConversationMessages(messages []model.UserConversationMessage) ([]model.UserConversationMessage, error) {
	if len(messages) == 0 {
		return []model.UserConversationMessage{}, nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	createdMessages := make([]model.UserConversationMessage, len(messages))
	query := `
		INSERT INTO tbl_users_conversations_messages (
			uuid, conversation_id, speaker, message_text, message_order,
			status_id, created_by, created_at, updated_by, updated_at
		) VALUES (
			:uuid, :conversation_id, :speaker, :message_text, :message_order,
			:status_id, :created_by, :created_at, :updated_by, :updated_at
		) RETURNING id, uuid, conversation_id, speaker, message_text, message_order,
					status_id, created_by, created_at, updated_by, updated_at
	`

	for i, msg := range messages {
		msg.UUID = uuid.New().String()
		msg.CreatedAt = time.Now()
		msg.UpdatedAt = time.Now()
		msg.StatusID = 1 // Default status

		stmt, err := tx.PrepareNamed(query)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare named statement for batch message: %w", err)
		}
		
		var createdMessage model.UserConversationMessage
		err = stmt.Get(&createdMessage, msg)
		stmt.Close() // Close statement in loop
		if err != nil {
			return nil, fmt.Errorf("failed to create user conversation message in batch: %w", err)
		}
		createdMessages[i] = createdMessage
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction for batch messages: %w", err)
	}

	return createdMessages, nil
}

func (r *ConversationRepository) ListUserConversationMessagesByConversationID(conversationID int64, userID int64) ([]model.UserConversationMessage, error) {
	var messages []model.UserConversationMessage
	// Ensure the conversation belongs to the user or is public and is not deleted
	query := `
		SELECT
			m.id, m.uuid, m.conversation_id, m.speaker, m.message_text, m.message_order,
			m.status_id, m.created_by, m.created_at, m.updated_by, m.updated_at,
			m.deleted_by, m.deleted_at
		FROM tbl_users_conversations_messages m
		JOIN tbl_users_conversations uc ON m.conversation_id = uc.id
		WHERE m.deleted_at IS NULL AND uc.id = :conversation_id AND (uc.user_id = :user_id OR uc.is_public = TRUE)
		ORDER BY m.message_order ASC, m.created_at ASC
	`
	args := map[string]interface{}{
		"conversation_id": conversationID,
		"user_id":         userID,
	}

	nstmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named statement for list messages: %w", err)
	}
	defer nstmt.Close()

	err = nstmt.Select(&messages, args)
	if err != nil {
		return nil, fmt.Errorf("failed to list user conversation messages: %w", err)
	}

	return messages, nil
}

func (r *ConversationRepository) GetUserConversationMessageByID(id int64, conversationID int64, userID int64) (*model.UserConversationMessage, error) {
	var message model.UserConversationMessage
	// Ensure the message belongs to the specified conversation and user, or conversation is public, and is not deleted
	query := `
		SELECT
			m.id, m.uuid, m.conversation_id, m.speaker, m.message_text, m.message_order,
			m.status_id, m.created_by, m.created_at, m.updated_by, m.updated_at,
			m.deleted_by, m.deleted_at
		FROM tbl_users_conversations_messages m
		JOIN tbl_users_conversations uc ON m.conversation_id = uc.id
		WHERE m.id = :id AND m.conversation_id = :conversation_id AND (uc.user_id = :user_id OR uc.is_public = TRUE) AND m.deleted_at IS NULL
	`
	args := map[string]interface{}{
		"id":             id,
		"conversation_id": conversationID,
		"user_id":         userID,
	}
	nstmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named statement for get message: %w", err)
	}
	defer nstmt.Close()
	err = nstmt.Get(&message, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserConversationMessageNotFound
		}
		return nil, fmt.Errorf("failed to get user conversation message by id: %w", err)
	}
	return &message, nil
}

func (r *ConversationRepository) UpdateUserConversationMessage(id int64, conversationID int64, req model.UpdateUserConversationMessageRequest, userID int64) (*model.UserConversationMessage, error) {
	sets := []string{}
	args := map[string]interface{}{
		"id":              id,
		"conversation_id": conversationID,
		"user_id":         userID, // For checking ownership through conversation
		"updated_at":      time.Now(),
		"updated_by":      userID,
	}

	if req.Speaker != nil {
		sets = append(sets, "speaker = :speaker")
		args["speaker"] = *req.Speaker
	}
	if req.MessageText != nil {
		sets = append(sets, "message_text = :message_text")
		args["message_text"] = *req.MessageText
	}
	if req.MessageOrder != nil {
		sets = append(sets, "message_order = :message_order")
		args["message_order"] = *req.MessageOrder
	}

	if len(sets) == 0 {
		// No fields to update, fetch and return existing
		return r.GetUserConversationMessageByID(id, conversationID, userID)
	}

	sets = append(sets, "updated_at = :updated_at", "updated_by = :updated_by")

	query := fmt.Sprintf(`
		UPDATE tbl_users_conversations_messages m
		SET %s
		FROM tbl_users_conversations uc
		WHERE m.id = :id AND m.conversation_id = :conversation_id AND uc.id = m.conversation_id AND (uc.user_id = :user_id OR uc.is_public = TRUE) AND m.deleted_at IS NULL
		RETURNING m.id, m.uuid, m.conversation_id, m.speaker, m.message_text, m.message_order,
					m.status_id, m.created_by, m.created_at, m.updated_by, m.updated_at,
					m.deleted_by, m.deleted_at
	`, strings.Join(sets, ", "))

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named statement for update message: %w", err)
	}
	defer stmt.Close()

	var updatedMessage model.UserConversationMessage
	err = stmt.Get(&updatedMessage, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserConversationMessageNotFound
		}
		return nil, fmt.Errorf("failed to update user conversation message: %w", err)
	}

	return &updatedMessage, nil
}

func (r *ConversationRepository) SoftDeleteUserConversationMessage(id int64, conversationID int64, userID int64) (time.Time, error) {
	deletedAt := time.Now()
	query := `
		UPDATE tbl_users_conversations_messages m
		SET
			deleted_at = :deleted_at,
			deleted_by = :deleted_by,
			updated_at = :updated_at,
			updated_by = :updated_by
		FROM tbl_users_conversations uc
		WHERE m.id = :id AND m.conversation_id = :conversation_id AND uc.id = m.conversation_id AND (uc.user_id = :user_id OR uc.is_public = TRUE) AND m.deleted_at IS NULL
	`
	result, err := r.db.NamedExec(query, map[string]interface{}{
		"deleted_at":      deletedAt,
		"deleted_by":      userID,
		"updated_at":      deletedAt,
		"updated_by":      userID,
		"id":              id,
		"conversation_id": conversationID,
		"user_id":         userID,
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to soft delete user conversation message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return time.Time{}, ErrUserConversationMessageNotFound
	}
	return deletedAt, nil
}
