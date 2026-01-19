package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RefreshService interface {
	Refresh(ctx context.Context) error
}

type RefreshController struct {
	service RefreshService
}

type RefreshResponse struct {
	Status string `json:"status"`
}

func NewRefreshController(service RefreshService) (*RefreshController, error) {
	if service == nil {
		return nil, errors.New("refresh service is nil")
	}

	return &RefreshController{service: service}, nil
}

func (c *RefreshController) RegisterRoutes(router *gin.Engine) error {
	if c == nil {
		return errors.New("refresh controller is nil")
	}
	if router == nil {
		return errors.New("router is nil")
	}

	router.GET("/refresh", c.refresh)
	return nil
}

func (c *RefreshController) refresh(ctx *gin.Context) {
	go func() {
		if err := c.service.Refresh(context.Background()); err != nil {
			log.Printf("refresh sources: %v", err)
		}
	}()

	ctx.JSON(http.StatusAccepted, RefreshResponse{Status: "started"})
}
