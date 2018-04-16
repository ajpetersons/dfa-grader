package main

import (
	"context"
	"dfa-grader/server"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

var port = pflag.IntP("port", "p", 80, "define port to listen to")
var help = pflag.BoolP("help", "h", false, "show usage")
var runServer = pflag.Bool("serve", false, "run http server")

func main() {
	pflag.Parse()
	if *help {
		pflag.PrintDefaults()
		return
	}

	if *runServer {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		router := server.NewRouter()

		webServer := &http.Server{
			Addr:         fmt.Sprintf(":%d", *port),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  10 * time.Second,
		}
		go func() {
			err := webServer.ListenAndServe()
			if err == http.ErrServerClosed {
				return
			}
			log.Println(err.Error())
		}()

		<-stop
		webServer.Shutdown(context.Background())
	}
}
