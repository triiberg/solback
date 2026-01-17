package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func RegisterHealthRoutes(router *gin.Engine) error {
	if router == nil {
		return errors.New("router is nil")
	}

	router.GET("/health", HealthHandler())
	return nil
}

func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
	}
}
