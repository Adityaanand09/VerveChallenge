package main

import (
	"VerveChallenge/Kafka"
	"VerveChallenge/VerveRequestHandler"
	"VerveChallenge/VerveTrackHandler"
	"VerveChallenge/internal/configs"
	"VerveChallenge/internal/dispatcher"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

func main() {
	app := gin.New()

	err := configs.Initialize()
	if err != nil {
		slog.Error("Terminating Server, error in initializing configs", "error", err)
		return
	}

	var (
		numWorkers      = viper.GetInt("NUMBER_OF_WORKERS")
		buffChannelSize = viper.GetInt("BUFFERED_CHANNEL_SIZE")
	)

	fw := Kafka.New(Kafka.Configs{Topic: viper.GetString("KAFKA_TOPIC"), WriteInterval: viper.GetInt("WRITE_INTERVAL_MIN"), Brokers: []string{viper.GetString("KAFKA_BROKER_ADDRESS")}, Username: "", Password: ""})

	d := dispatcher.NewAsyncDispatcher(numWorkers, buffChannelSize, fw)
	verveHandler := VerveRequestHandler.New(d, fw)
	trackHandler := VerveTrackHandler.New()

	//Producer.New(Producer.Configs{WriteInterval: 1})
	app.GET("/api/verve/accept", verveHandler.HandleJson)
	app.POST("/api/verve/track", trackHandler.HandleJson)

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
