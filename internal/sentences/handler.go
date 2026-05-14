package sentences

import (
	"errors"
	"log"
	"sentenceminer/internal/sentences/model"
	"sentenceminer/internal/sentences/repository"
	"sentenceminer/internal/sentences/service"
	"sentenceminer/pkg/http/response"
	"sentenceminer/pkg/postgres"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type SentenceHandler struct {
	service *service.SentenceService
}

func NewSentenceHandler(db *sqlx.DB) *SentenceHandler {
	return &SentenceHandler{
		service: service.NewSentenceService(db),
	}
}

func (h *SentenceHandler) analyzeSentence(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	sentence, analysis, err := h.service.AnalyzeExisting(c.Context(), id)
	if err != nil {
		log.Printf("[SentenceMiner] analyze_sentence_failed sentence_id=%d error=%s", id, sanitizeLogValue(err.Error()))
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("sentence analyzed", fiber.StatusOK, fiber.Map{
		"sentence": sentence,
		"analysis": analysis,
	}))
}

func (h *SentenceHandler) createSentence(c *fiber.Ctx) error {
	var req model.CreateSentenceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return fiber.NewError(fiber.StatusBadRequest, "text is required")
	}

	sentence, err := h.service.Create(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(response.NewResponse("sentence created", fiber.StatusCreated, fiber.Map{
		"sentence": sentence,
	}))
}

func (h *SentenceHandler) importEnvironmentPack(c *fiber.Ctx) error {
	imported, err := h.service.ImportEnvironmentPack()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("environment pack imported", fiber.StatusOK, fiber.Map{
		"status":   "ok",
		"imported": imported,
	}))
}

func (h *SentenceHandler) show(c *fiber.Ctx) error {
	req, err := postgres.ExtractQueryParamsRequest(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	items, total, respErr := h.service.Show(*req)
	if respErr != nil {
		return c.Status(respErr.StatusCode).JSON(response.ApiResponseError(false, respErr.MessageID, respErr.StatusCode, respErr.Err))
	}
	return response.JSONWithPaging(c, fiber.StatusOK, "sentences fetched", items, req.PagingOptions.Page, req.PagingOptions.PerPage, total)
}

func (h *SentenceHandler) getSentence(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	sentence, err := h.service.GetSentence(id)
	if err != nil {
		if errors.Is(err, repository.ErrSentenceNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "sentence not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("sentence fetched", fiber.StatusOK, fiber.Map{
		"sentence": sentence,
	}))
}

func (h *SentenceHandler) categoryFeed(c *fiber.Ctx) error {
	var req model.CategoryFeedRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	if strings.TrimSpace(req.Category) == "" {
		req.Category = "general"
	}
	if req.Limit <= 0 {
		req.Limit = 12
	}

	items, generated, err := h.service.GetCategoryFeed(c.Context(), req.Category, req.Focus, req.ExcludeIDs, req.Limit)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(response.NewResponse("category feed fetched with error", fiber.StatusOK, fiber.Map{
			"sentences": []any{},
			"generated": 0,
			"hasMore":   false,
			"error":     err.Error(),
		}))
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("category feed fetched", fiber.StatusOK, fiber.Map{
		"sentences": items,
		"generated": generated,
		"hasMore":   len(items) == req.Limit || generated > 0,
	}))
}

func (h *SentenceHandler) rateSentenceReview(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	var req model.ReviewRatingRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	rating := strings.TrimSpace(strings.ToLower(req.Rating))
	if rating == "" {
		return fiber.NewError(fiber.StatusBadRequest, "rating is required")
	}

	sentence, err := h.service.RateSentenceReview(id, rating)
	if err != nil {
		if errors.Is(err, repository.ErrSentenceNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "sentence not found")
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("sentence review updated", fiber.StatusOK, fiber.Map{
		"sentence": sentence,
	}))
}

func (h *SentenceHandler) generateSpeech(c *fiber.Ctx) error {
	var req model.TTSRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}

	text := strings.TrimSpace(req.Text)
	if text == "" {
		return fiber.NewError(fiber.StatusBadRequest, "text is required")
	}

	mode := strings.TrimSpace(strings.ToLower(req.Mode))
	if mode == "" {
		mode = "normal"
	}

	audio, err := h.service.GenerateSpeech(c.Context(), text, mode)
	if err != nil {
		log.Printf("[SentenceMiner] tts_failed mode=%s text_len=%d error=%s", mode, len(text), sanitizeLogValue(err.Error()))
		return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
	}

	c.Set("Content-Type", "audio/wav")
	c.Set("Content-Disposition", `inline; filename="sentence.wav"`)
	return c.Send(audio)
}

func (h *SentenceHandler) deleteSentence(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	deletedAt, err := h.service.SoftDeleteSentence(id)
	if err != nil {
		if errors.Is(err, repository.ErrSentenceNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "sentence not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.NewResponse("sentence deleted", fiber.StatusOK, fiber.Map{
		"status":     "deleted",
		"sentenceId": id,
		"deletedAt":  deletedAt,
	}))
}

func sanitizeLogValue(value string) string {
	cleaned := strings.ReplaceAll(value, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\r", " ")
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}
	return strings.TrimSpace(cleaned)
}
