package main

import (
	"context"
	"dfa-grader/config"
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
var configPath = pflag.StringP("config", "c", "", "configuration file path without extension")

func main() {
	pflag.Parse()
	if *help {
		pflag.PrintDefaults()
		return
	}

	if *runServer {
		err := config.Read(*configPath)
		if err != nil {
			fmt.Printf("Could not read config file: %s", err.Error())
			return
		}
		stop := make(chan os.Signal, 1)
		defer close(stop)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		reload := make(chan os.Signal, 5)
		defer close(reload)
		signal.Notify(reload, syscall.SIGHUP)
		go func() {
			for {
				_, more := <-reload
				if !more {
					return
				}
				fmt.Println("Reloading configuration")

				err := config.Read(*configPath)
				if err != nil {
					fmt.Printf("Could not reload configuration: %s", err.Error())
				}
			}
		}()

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
		webServer.Shutdown(context.Background()) // nolint: gas, errcheck
	}
}
