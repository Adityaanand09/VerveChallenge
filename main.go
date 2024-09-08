package main

import (
	"VerveChallenge/internal/config"
	"VerveChallenge/internal/dispatcher"
	"errors"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"VerveChallenge/FileWriter"
	"VerveChallenge/VerveRequestHandler"
	"VerveChallenge/VerveTrackHandler"

	"github.com/gin-gonic/gin"
)

func main() {
	app := gin.New()

	err := config.Initialize()
	if err != nil {
		slog.Error("Terminating Server, error in initializing configs", "error", err)
		return
	}

	var (
		numWorkers      = viper.GetInt("NUMBER_OF_WORKERS")
		buffChannelSize = viper.GetInt("BUFFERED_CHANNEL_SIZE")
	)

	fw := FileWriter.New(FileWriter.Configs{FileName: viper.GetString("FILENAME"), WriteInterval: viper.GetInt("WRITE_INTERVAL_MIN")})
	d := dispatcher.NewAsyncDispatcher(numWorkers, buffChannelSize, fw)
	verveHandler := VerveRequestHandler.New(fw, d)
	trackHandler := VerveTrackHandler.New()

	//FileWriter.New(FileWriter.Configs{WriteInterval: 1})
	app.GET("/api/verve/accept", verveHandler.HandleJson)
	app.GET("/api/verve/track", trackHandler.HandleJson)

	server := &http.Server{
		Addr:    ":8080",
		Handler: app,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)

	//if there is an error in starting the server, we pass the os.Kill signal to the quit channel,
	//here we are checking if the error is other than http: Server closed since this error is returned
	//whenever the running server will be closed , and in that case we don't want a log "Failed to start
	//server" to be logged
	go func() {
		slog.Info("Starting Server at 8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start server", "error", err)
			quit <- os.Kill
		}
	}()

	<-quit
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
