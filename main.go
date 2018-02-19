package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/mattermost/mattermost-server/model"
	"gopkg.in/go-playground/validator.v9"
)

type ARGs struct {
	Email    string `validate:"required"`
	Password string `validate:"required"`
	TeamName string `validate:"required"`
	Hostname string `validate:"required"`
	Socket   string `validate:"required"`
	Debug    bool
	Token    string `validate:"required"`
	Lang     string `validate:"required"`
}

func main() {
	args := ARGs{}
	flag.StringVar(&args.Email, "me", "", "mattermost email")
	flag.StringVar(&args.Hostname, "mh", "", "mattermost hostname")
	flag.StringVar(&args.Password, "mp", "", "mattermost password")
	flag.StringVar(&args.TeamName, "mt", "", "mattermost team name")
	flag.StringVar(&args.Socket, "ms", "", "websocket address")
	flag.BoolVar(&args.Debug, "d", false, "debug mode")
	flag.StringVar(&args.Token, "dt", "", "dialogflow token")
	flag.StringVar(&args.Lang, "dl", "", "dialogflow lang")
	flag.Parse()

	if err := validator.New().Struct(args); err != nil {
		log.Fatalln(err)
	}

	client := model.NewAPIv4Client(args.Hostname)
	usr, resp := client.Login(args.Email, args.Password)
	if resp.Error != nil {
		log.Fatalln(resp.Error.Error())
	}
	if args.Debug {
		log.Printf("%+v\n", usr)
	}

	webSocketClient, err := model.NewWebSocketClient4(args.Socket, client.AuthToken)
	if err != nil {
		log.Fatalln(err.Error())
	}

	webSocketClient.Listen()

	go func() {
		for {
			select {
			case resp := <-webSocketClient.EventChannel:
				if args.Debug {
					log.Printf("%+v\n", resp)
				}
				if err := parseEvent(resp, usr, client, &args); err != nil {
					log.Println(err)
				}
			}
		}
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		for range c {
			if webSocketClient != nil {
				webSocketClient.Close()
			}
			log.Println("exit")
			os.Exit(0)
		}
	}()

	select {}
}
