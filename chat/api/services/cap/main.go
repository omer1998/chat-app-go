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

	"github.com/omer1998/chat-app-go.git/chat/app/sdk/mux"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
)

func main() {
	// traceIDFn := func(cxt context.Context) string{

	// }
	cxt := context.Background()
	// log := logger.NewWithHandler(slog.NewJSONHandler(os.Stdout, nil))
	log := logger.New(os.Stdout, logger.LevelDebug, "CAP", nil)
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

	log.Info(cxt, "Start up", "GOMAXPROS", runtime.GOMAXPROCS(0))
	defer log.Info(cxt, "shutdown complete")

	// here we need to open the server we need to reach listen and serve

	webApi := mux.WebAPI(mux.Config{Log: log})
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
