package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/omer1998/chat-app-go.git/chat/app/domain/chatapp"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/mux"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
	"github.com/omer1998/chat-app-go.git/chat/foundation/web"
)

func main() {
	// traceIDFn := func(cxt context.Context) string{

	// }
	cxt := context.Background()
	// log := logger.NewWithHandler(slog.NewJSONHandler(os.Stdout, nil))
	traceIdFun := func(cxt context.Context) string {
		return web.GetTraceId(cxt).String()
	}
	log := logger.New(os.Stdout, logger.LevelInfo, "CAP", traceIdFun)
	log.Info(cxt, "intialize the CAP service", "name", slog.StringValue("omer"))
	// doneChan := make(chan bool, 1)
	if err := run(cxt, log); err != nil {
		log.Error(cxt, "startup", "err", err)
		os.Exit(1)
	}

	fmt.Println("continue running the main function ..")
	// <-doneChan
}

func run(cxt context.Context, log *logger.Logger) error {
	// here we need to open the server we need to reach listen and serve
	// configure nats connection and jetstream
	log.Info(cxt, "Start up", "GOMAXPROS", runtime.GOMAXPROCS(0))
	defer log.Info(cxt, "shutdown complete")

	// >==================================================================
	nc, err := nats.Connect("demo.nats.io")
	if err != nil {
		return fmt.Errorf("error nats connection: %w", err)
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		return fmt.Errorf("error creating jetstream: %w", err)
	}
	sub := "omercap"
	// creating stream
	s, err := js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
		Name:     sub,
		Subjects: []string{sub},
	})
	if err != nil {
		return fmt.Errorf("error creating stream: %w", err)
	}
	//create or update consummer
	_, err = s.CreateOrUpdateConsumer(cxt, jetstream.ConsumerConfig{
		Name:      "omerconsumer",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("error creating consumer: %w", err)
	}

	a, err := chatapp.NewApp(log, js, sub, s)
	if err != nil {
		return fmt.Errorf("error creating app: %w", err)
	}
	webApi := mux.WebAPI(mux.Config{Log: log, Js: js, Subject: sub, Stream: s, Api: a})
	if webApi == nil {
		log.Error(cxt, "webApi is nil")
		return fmt.Errorf("webApi is nil")
	}

	app := http.Server{
		Addr:    "localhost:3000",
		Handler: webApi,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info(cxt, "Server starting now")

		serverErr <- app.ListenAndServe()
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error(cxt, "error from sever ", "Error", err.Error())
		return fmt.Errorf("err from server starting : %s", err.Error())

	case signal := <-shutdownChan:
		log.Info(cxt, "shutdown", "status", "shutdwon started", "signal", signal)
		defer log.Info(cxt, "shutdown", "status", "shutdown complete", "signal", signal)
		cxt, cancel := context.WithTimeout(cxt, time.Duration(time.Second*4))
		defer cancel()
		if err := app.Shutdown(cxt); err != nil {
			app.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

	}

	return nil

}
