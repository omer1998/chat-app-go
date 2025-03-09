package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/google/uuid"
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
	// retrieve or set cap id

	filePath := filepath.Join("../../../zarf/")
	_, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error making dir path for cap file: %w", err)
		}
		f, err := os.Create(filePath + "/" + "cap.id")
		if err != nil {
			return fmt.Errorf("error creating cap file: %w", err)
		}
		_, err = f.WriteString(uuid.NewString())
		if err != nil {
			return fmt.Errorf("error writing cap file: %w", err)
		}
		f.Close()
	}
	f, err := os.Open(filePath + "/" + "cap.id")
	if err != nil {
		return fmt.Errorf("error opening cap file: %w", err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading cap file: %w", err)
	}
	capId := string(data)

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
		MaxAge:   time.Hour * 24,
	})

	if err != nil {
		return fmt.Errorf("error creating stream: %w", err)
	}
	// consName := uuid.NewString()
	//create or update consummer

	cons, err := s.CreateOrUpdateConsumer(cxt, jetstream.ConsumerConfig{
		Durable:       capId,
		Name:          capId,
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverNewPolicy,
	})
	if err != nil {
		return fmt.Errorf("error creating consumer: %w", err)
	}
	// defer s.DeleteConsumer(cxt, "omerconsumer")

	capIdUUUID := uuid.MustParse(capId)
	fmt.Println(">>>>> capIdUUID: ", capIdUUUID)
	a, err := chatapp.NewApp(log, js, sub, s, cons, capIdUUUID)
	if err != nil {
		return fmt.Errorf("error creating app: %w", err)
	}
	webApi := mux.WebAPI(mux.Config{Log: log, Api: a})
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
