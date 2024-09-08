package VerveRequestHandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"VerveChallenge/internal"

	"github.com/gin-gonic/gin"
)

type RequestHandler struct {
	id         string `json:"id"`
	endpoint   string `json:"endpoint"`
	dispatcher dispatcher
	fileWriter FileWriter
}

type RequestData struct {
	Count int `json:"count"`
}
type FileWriter interface {
	GetValue() int
	Write()
}

type dispatcher interface {
	Dispatch(m internal.Message)
}

func New(fw FileWriter, d dispatcher) RequestHandler {
	r1 := RequestHandler{dispatcher: d, fileWriter: fw}
	return r1
}

func (r RequestHandler) HandleJson(ctx *gin.Context) {
	var req = &RequestHandler{}
	req.id = ctx.Request.URL.Query().Get("id")
	req.endpoint = ctx.Request.URL.Query().Get("endpoint")

	err := r.helper(req.id, req.endpoint)
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		ctx.Writer.Write([]byte("failed"))
		return
	}
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Writer.Write([]byte("ok"))
	return
}

func (r RequestHandler) helper(id string, endpoint string) error {
	val := r.fileWriter.GetValue()
	requestData := RequestData{
		Count: val,
	}

	// Convert the data structure to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return err
	}
	if endpoint != "" {
		resp, err := http.Post("http://localhost:8080/api/verve/"+endpoint, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error sending GET request: %v", err)
		}
		defer resp.Body.Close()
		slog.Info("Response code = ", "response code", resp.StatusCode)
	}
	newID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	r.dispatcher.Dispatch(internal.Message{Id: newID})

	return nil
}
