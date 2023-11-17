package server

import (
	"concurrency-practice/pkg/transmuxer"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
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

			log.Println("START: Setup File Server")
			{
				_, err := os.Stat(cfg.StaticDirName)
				if err != nil {
					if exists := os.IsNotExist(err); !exists {
						log.Println(errors.New("static directory does not exist"))
					} else {
						log.Println(err.Error())
					}
					return nil
				}
				tmpDir := uuid.New().String()
				defer func() {
					if _, err := os.Stat(tmpDir); err != nil {
						fmt.Println(err.Error())
					}
					if err := os.RemoveAll(tmpDir); err != nil {
						fmt.Println(err.Error())
					}
				}()
				if err = os.MkdirAll(tmpDir, 0755); err != nil {
					log.Println(err.Error())
					return nil
				}
				if err != nil {
					log.Println(err.Error())
					return nil
				}
				transmuxer.TheServer = transmuxer.NewFileServer(tmpDir, cfg.StaticDirName)
				log.Println("DONE: Setup File Server")
			}

			log.Println("START: Setup Server")
			{
				r := mux.NewRouter()
				if err := registerHandlers(r); err != nil {
					log.Println(err.Error())
					return nil
				}
				http.Handle("/", r)
				log.Println("DONE: Setup Server")
				log.Println("Listening on port", cfg.Port)
				if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
					log.Println("FAILED: Setup Server")
					log.Println("ERROR:", err)
					return nil
				}
			}
			return nil
		},
	}
}
