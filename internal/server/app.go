package server

import (
	"concurrency-practice/internal/store"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func App() *cli.App {
	return &cli.App{
		Name:        "transmuxer",
		Description: "hls playlists to mkv",
		Usage:       "input a url ",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "port",
				Aliases:  []string{"p"},
				Usage:    "the port",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "static-dir-name",
				Aliases:  []string{"s"},
				Usage:    "name of long term mediafile storage directory",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			cfg := Config{
				StaticDirName: c.String("static-dir-name"),
				Port:          c.String("port"),
			}
			var err error

			log.Println("START: Setup File Server")
			{
				if store.TheServer, err = store.NewFileServer(cfg.StaticDirName); err != nil {
					log.Println(err.Error())
					return err
				}
			}
			log.Println("DONE: Setup File Server")

			log.Println("START: Setup HTTP Server")
			{
				r := mux.NewRouter()
				if err = registerHandlers(r); err != nil {
					log.Println(err.Error())
					return err
				}
				http.Handle("/", r)
				shutdown := make(chan os.Signal, 1)
				signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
				go func() {
					if err = http.ListenAndServe(":"+cfg.Port, nil); err != nil {
						log.Println("FAILED: Setup Server")
						log.Println("ERROR:", err)
						panic(err)
					}
				}()
				log.Println("DONE: Setup Server")
				log.Println("Listening on port", cfg.Port)

				// SHUTDOWN
				<-shutdown
				log.Println("START: Shutting down...")
				store.TheServer.Cleanup()
				log.Println("DONE: Shut down the server")
			}
			return nil
		},
	}
}
