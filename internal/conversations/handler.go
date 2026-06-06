package conversations

import (
	"errors"
	"sentenceminer/internal/conversations/model"
	"sentenceminer/internal/conversations/repository"
	"sentenceminer/pkg/http/response"
	"sentenceminer/pkg/logs"
	"sentenceminer/pkg/postgres"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type ConversationHandler struct {
	service *ConversationService
}

func NewConversationHandler(db *sqlx.DB) *ConversationHandler {
	return &ConversationHandler{
		service: NewConversationService(db),
	}
}

// User Conversation Handlers

func (h *ConversationHandler) createUserConversation(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64) // Must be authenticated
	var req model.CreateUserConversationRequest
	if err := c.BodyParser(&req); err != nil {
		logs.Error("create_user_conversation_body_parse_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := req.Bind(); err != nil {
		logs.Error("create_user_conversation_bind_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	conversation, err := h.service.CreateUserConversation(req, userID)
	if err != nil {
		logs.Error("create_user_conversation_service_error", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(response.NewResponse(c, "user conversation created", fiber.StatusCreated, fiber.Map{
		"conversation": conversation,
	}))
}

func (h *ConversationHandler) showUserConversations(c *fiber.Ctx) error {
	// Determine if the request is for public conversations or user's own conversations
	var requestingUserID int64
	if c.Locals("userID") != nil {
		requestingUserID = c.Locals("userID").(int64) // Authenticated user
	} else {
		requestingUserID = 0 // Anonymous request
	}

	var conversationReq model.UserConversationsShowRequest
	// Extract query params for filtering, sorting, pagination
	queryParams, err := postgres.ExtractQueryParamsRequest(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	}
	conversationReq.PagingOptions = queryParams.PagingOptions
	conversationReq.Filters = queryParams.Filters
	conversationReq.Sorts = queryParams.Sorts
	conversationReq.Offset = queryParams.Offset

	// Extract is_public query parameter if present
	isPublicStr := c.Query("is_public")
	if isPublicStr != "" {
		isPublicBool, err := strconv.ParseBool(isPublicStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid value for is_public parameter")
		}
		conversationReq.IsPublic = &isPublicBool
	}
	
	conversations, total, respErr := h.service.ListUserConversations(conversationReq, requestingUserID)
	if respErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ApiResponseError(false, "failed to fetch user conversations", fiber.StatusInternalServerError, respErr))
	}
	return response.JSONWithPaging(c, fiber.StatusOK, "user conversations fetched", conversations, queryParams.PagingOptions.Page, queryParams.PagingOptions.PerPage, total)
}

func (h *ConversationHandler) getUserConversation(c *fiber.Ctx) error {
	// Determine if the request is for public conversations or user's own conversations
	var requestingUserID int64
	if c.Locals("userID") != nil {
		requestingUserID = c.Locals("userID").(int64) // Authenticated user
	} else {
		requestingUserID = 0 // Anonymous request
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid conversation ID")
	}

	conversation, err := h.service.GetUserConversation(id, requestingUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserConversationNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user conversation not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "user conversation fetched", fiber.StatusOK, fiber.Map{
		"conversation": conversation,
	}))
}

func (h *ConversationHandler) updateUserConversation(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64) // Must be authenticated and owner
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid conversation ID")
	}

	var req model.UpdateUserConversationRequest
	if err := c.BodyParser(&req); err != nil {
		logs.Error("create_user_conversation_body_parse_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := req.Bind(); err != nil {
		logs.Error("create_user_conversation_bind_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	updatedConversation, err := h.service.UpdateUserConversation(id, req, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserConversationNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user conversation not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "user conversation updated", fiber.StatusOK, fiber.Map{
		"conversation": updatedConversation,
	}))
}

func (h *ConversationHandler) deleteUserConversation(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64) // Must be authenticated and owner
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid conversation ID")
	}

	err = h.service.SoftDeleteUserConversation(id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserConversationNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user conversation not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "user conversation deleted", fiber.StatusOK, nil))
}

// User Conversation Message Handlers

func (h *ConversationHandler) addMessageToUserConversation(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64) // Must be authenticated and owner of conversation
	conversationID, err := strconv.ParseInt(c.Params("conversationId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid conversation ID")
	}

	var req model.CreateUserConversationMessageRequest
	if err := c.BodyParser(&req); err != nil {
		logs.Error("create_user_conversation_body_parse_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := req.Bind(); err != nil {
		logs.Error("create_user_conversation_bind_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	message, err := h.service.AddMessageToUserConversation(conversationID, req, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserConversationNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "conversation not found or does not belong to user")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(response.NewResponse(c, "message added to user conversation", fiber.StatusCreated, fiber.Map{
		"message": message,
	}))
}

func (h *ConversationHandler) addMultipleMessagesToUserConversation(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64) // Must be authenticated and owner of conversation
	conversationID, err := strconv.ParseInt(c.Params("conversationId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid conversation ID")
	}

	var req model.CreateMultipleUserConversationMessagesRequest
	if err := c.BodyParser(&req); err != nil {
		logs.Error("create_user_conversation_body_parse_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := req.Bind(); err != nil {
		logs.Error("create_user_conversation_bind_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	messages, err := h.service.AddMultipleMessagesToUserConversation(conversationID, req, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserConversationNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "conversation not found or does not belong to user")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(response.NewResponse(c, "multiple messages added to user conversation", fiber.StatusCreated, fiber.Map{
		"messages": messages,
	}))
}

func (h *ConversationHandler) listUserConversationMessages(c *fiber.Ctx) error {
	// Determine if the request is for public conversations or user's own conversations
	var requestingUserID int64
	if c.Locals("userID") != nil {
		requestingUserID = c.Locals("userID").(int64) // Authenticated user
	} else {
		requestingUserID = 0 // Anonymous request
	}
	
	conversationID, err := strconv.ParseInt(c.Params("conversationId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid conversation ID")
	}

	messages, err := h.service.ListUserConversationMessages(conversationID, requestingUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "user conversation messages fetched", fiber.StatusOK, fiber.Map{
		"messages": messages,
	}))
}

func (h *ConversationHandler) getUserConversationMessage(c *fiber.Ctx) error {
	// Determine if the request is for public conversations or user's own conversations
	var requestingUserID int64
	if c.Locals("userID") != nil {
		requestingUserID = c.Locals("userID").(int64) // Authenticated user
	} else {
		requestingUserID = 0 // Anonymous request
	}

	conversationID, err := strconv.ParseInt(c.Params("conversationId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid conversation ID")
	}
	messageID, err := strconv.ParseInt(c.Params("messageId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid message ID")
	}

	message, err := h.service.GetUserConversationMessage(messageID, conversationID, requestingUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserConversationMessageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user conversation message not found")
		}
		if errors.Is(err, repository.ErrUserConversationNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "conversation not found or does not belong to user")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "user conversation message fetched", fiber.StatusOK, fiber.Map{
		"message": message,
	}))
}

func (h *ConversationHandler) updateUserConversationMessage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64) // Must be authenticated and owner of conversation
	conversationID, err := strconv.ParseInt(c.Params("conversationId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid conversation ID")
	}
	messageID, err := strconv.ParseInt(c.Params("messageId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid message ID")
	}

	var req model.UpdateUserConversationMessageRequest
	if err := c.BodyParser(&req); err != nil {
		logs.Error("create_user_conversation_body_parse_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := req.Bind(); err != nil {
		logs.Error("create_user_conversation_bind_error", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	updatedMessage, err := h.service.UpdateUserConversationMessage(messageID, conversationID, req, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserConversationMessageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user conversation message not found")
		}
		if errors.Is(err, repository.ErrUserConversationNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "conversation not found or does not belong to user")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "user conversation message updated", fiber.StatusOK, fiber.Map{
		"message": updatedMessage,
	}))
}

func (h *ConversationHandler) deleteUserConversationMessage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64) // Must be authenticated and owner of conversation
	conversationID, err := strconv.ParseInt(c.Params("conversationId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid conversation ID")
	}
	messageID, err := strconv.ParseInt(c.Params("messageId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid message ID")
	}

	err = h.service.SoftDeleteUserConversationMessage(messageID, conversationID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserConversationMessageNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user conversation message not found")
		}
		if errors.Is(err, repository.ErrUserConversationNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "conversation not found or does not belong to user")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(c, "user conversation message deleted", fiber.StatusOK, nil))
}
