package VerveRequestHandler

import (
	dispatcher2 "VerveChallenge/internal/dispatcher"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strconv"
)

type RequestHandler struct {
	Id         string `json:"id"`
	Endpoint   string `json:"endpoint"`
	dispatcher dispatcher
	producer   Producer
}

type RequestData struct {
	Count int    `json:"count"`
	ID    string `json:"id"`
}

type dispatcher interface {
	Dispatch(m dispatcher2.Message)
}

type Producer interface {
	GetValue() int
	Produce(ctx context.Context, key string, payload []byte) error
}

func New(d dispatcher, p Producer) RequestHandler {
	r1 := RequestHandler{dispatcher: d, producer: p}
	return r1
}

func (r RequestHandler) HandleJson(ctx *gin.Context) {
	var req = &RequestHandler{}
	req.Id = ctx.Request.URL.Query().Get("id")
	req.Endpoint = ctx.Request.URL.Query().Get("endpoint")

	err := r.helper(req.Id, req.Endpoint)
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		ctx.Writer.Write([]byte("failed"))
		return
	}
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Writer.Write([]byte("ok"))
	return
}

func (r RequestHandler) helper(id string, endpoint string) error {
	val := r.producer.GetValue()
	requestData := RequestData{
		Count: val,
		ID:    id,
	}

	// Convert the data structure to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return err
	}
	if endpoint != "" {
		resp, err := http.Post("http://localhost:8080/api/verve/"+endpoint, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error sending POST request: %v", err)
		}
		defer resp.Body.Close()
		slog.Info("Response code", "code=", resp.StatusCode)
	}
	newID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	r.dispatcher.Dispatch(dispatcher2.Message{Id: newID})

	return nil
}
