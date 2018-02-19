package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	protocolVersion = "20150910"
	baseURL         = "https://api.dialogflow.com/v1"
)

type dialogFlowRequest struct {
	Query     string `json:"query"`
	SessionID string `json:"sessionId"`
	Lang      string `json:"lang"`
	Version   string `json:"v"`
}

type dialogFlowResponse struct {
	ID        string           `json:"id"`
	Timestamp time.Time        `json:"timestamp"`
	Lang      string           `json:"lang"`
	SessionID string           `json:"session_id"`
	Result    dialogFlowResult `json:"result"`
	Status    dialogFlowStatus `json:"status"`
}

type dialogFlowResult struct {
	Source        string             `json:"source"`
	ResolvedQuery string             `json:"resolvedQuery"`
	Speech        string             `json:"speech"`
	Action        string             `json:"action"`
	Parameters    interface{}        `json:"parameters"`
	MetaData      dialogFlowMetaData `json:"metadata"`
}

type dialogFlowMetaData struct {
	InputContext              []interface{} `json:"inputContexts"`
	OutputContexts            []interface{} `json:"outputContexts"`
	IntentName                string        `json:"intentName"`
	IntentID                  string        `json:"intentId"`
	WebHookUsed               string        `json:"webhookUsed"`
	WebHookForSlotFillingUsed string        `json:"webhookForSlotFillingUsed"`
	Contexts                  []interface{} `json:"contexts"`
}

type dialogFlowStatus struct {
	Code            int    `json:"code"`
	ErrorType       string `json:"errorType"`
	WebHookTimedOut bool   `json:"webhookTimedOut"`
}

func sendDialogFlow(userID string, query, token, lang string) (*dialogFlowResponse, error) {
	httpClient := http.Client{Timeout: time.Second * 15}

	bts, err := json.Marshal(dialogFlowRequest{
		Query:     query,
		Lang:      lang,
		SessionID: userID,
		Version:   protocolVersion,
	})
	if err != nil {
		return nil, err
	}

	rq, err := http.NewRequest("POST", fmt.Sprintf(`%s/query`, baseURL),
		bytes.NewReader(bts))
	if err != nil {
		return nil, err
	}
	rq.Header.Add("Content-Type", "application/json; charset=UTF-8")
	rq.Header.Add("Accept", "application/json")
	rq.Header.Add("Authorization", "Bearer "+token)

	rsp, err := httpClient.Do(rq)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		bts, _ := ioutil.ReadAll(rsp.Body)
		return nil, errors.New(string(bts))
	}

	var dfRsp dialogFlowResponse
	if err := json.NewDecoder(rsp.Body).Decode(&dfRsp); err != nil {
		return nil, err
	}

	return &dfRsp, nil
}
