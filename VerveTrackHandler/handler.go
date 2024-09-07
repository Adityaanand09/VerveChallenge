package VerveTrackHandler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type RequestHandler struct {
}

func New() RequestHandler {
	return RequestHandler{}
}

func (r RequestHandler) HandleJson(ctx *gin.Context) {
	ctx.Writer.WriteHeader(http.StatusOK)
	return
}
