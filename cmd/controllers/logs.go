package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"solback/internal/models"

	"github.com/gin-gonic/gin"
)

const defaultLogsLimit = 20

type LogProvider interface {
	GetLogs(ctx context.Context, limit int, eventID string) ([]models.Log, error)
	TruncateLogs(ctx context.Context) (int, error)
}

type LogsController struct {
	service LogProvider
}

type DeleteLogsResponse struct {
	Deleted int `json:"deleted"`
}

func NewLogsController(service LogProvider) (*LogsController, error) {
	if service == nil {
		return nil, errors.New("log service is nil")
	}

	return &LogsController{service: service}, nil
}

func (c *LogsController) RegisterRoutes(router *gin.Engine) error {
	if c == nil {
		return errors.New("logs controller is nil")
	}
	if router == nil {
		return errors.New("router is nil")
	}

	router.GET("/logs", c.getLogs)
	router.DELETE("/logs", c.deleteLogs)
	return nil
}

func (c *LogsController) getLogs(ctx *gin.Context) {
	limit, err := parseLogsLimit(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid logs limit"})
		return
	}

	eventID := parseLogsEventID(ctx)
	logs, err := c.service.GetLogs(ctx.Request.Context(), limit, eventID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load logs"})
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func (c *LogsController) deleteLogs(ctx *gin.Context) {
	deleted, err := c.service.TruncateLogs(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete logs"})
		return
	}

	ctx.JSON(http.StatusOK, DeleteLogsResponse{Deleted: deleted})
}

func parseLogsLimit(ctx *gin.Context) (int, error) {
	value := ctx.Query("n")
	if value == "" {
		return defaultLogsLimit, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if limit <= 0 {
		return 0, errors.New("limit must be positive")
	}

	return limit, nil
}

func parseLogsEventID(ctx *gin.Context) string {
	eventID := ctx.Query("eventId")
	if eventID == "" {
		eventID = ctx.Query("event_id")
	}
	return eventID
}
