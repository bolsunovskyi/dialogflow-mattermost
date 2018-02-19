package main

import (
	"encoding/json"
	"errors"
	"log"
	"regexp"

	"github.com/mattermost/mattermost-server/model"
)

func isMentioned(e *model.WebSocketEvent, id string) bool {
	reg, _ := regexp.Compile(`\w+`)

	if v, ok := e.Data["mentions"]; ok {
		if vs, ok := v.(string); ok {
			ids := reg.FindAllString(vs, -1)
			for _, mentionID := range ids {
				if mentionID == id {
					return true
				}
			}
		}
	}

	return false
}

type Post struct {
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	ChannelID string `json:"channel_id"`
}

func getPostData(e *model.WebSocketEvent) (*Post, error) {
	if v, ok := e.Data["post"]; ok {
		if str, ok := v.(string); ok {
			var post Post
			if err := json.Unmarshal([]byte(str), &post); err == nil {
				return &post, nil
			}
		}
	}

	return nil, errors.New("wrong post data")
}

func parseEvent(e *model.WebSocketEvent, user *model.User, client *model.Client4, args *ARGs) error {
	if e.Event == model.WEBSOCKET_EVENT_POSTED {
		if isMentioned(e, user.Id) {
			if args.Debug {
				log.Println("bot mentioned")
			}

			post, err := getPostData(e)
			if err != nil {
				return err
			}

			if args.Debug {
				log.Printf("post: %+v", *post)
			}

			rsp, err := sendDialogFlow(post.UserID, post.Message, args.Token, args.Lang)
			if err != nil {
				return err
			}

			if _, matterRsp := client.CreatePost(&model.Post{
				Message:   rsp.Result.Speech,
				ChannelId: post.ChannelID,
			}); matterRsp.Error != nil {
				return errors.New(matterRsp.Error.Error())
			}
		}
	}

	return nil
}
