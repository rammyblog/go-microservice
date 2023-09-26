package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rammyblog/broker/event"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := JsonResponse{
		Error:   false,
		Message: "Hit the Broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)

	case "log":
		app.logEventViaRabbit(w, requestPayload.Log)

	case "mail":
		app.sendMail(w, requestPayload.Mail)

	default:
		app.errorJSON(w, errors.New("unknown action"))

	}

}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {

	jsonData, _ := json.Marshal(entry)
	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer((jsonData)))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, err)
		return
	}

	var payload JsonResponse

	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)

}
func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json to send to the auth ms

	jsonData, _ := json.MarshalIndent(a, "", "\t")

	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer((jsonData)))

	if err != nil {
		app.errorJSON(w, err)

		return
	}

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)

		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid crendentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService JsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)

	if err != nil {
		app.errorJSON(w, err)

		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)

	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data
	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.Marshal(msg)

	mailServiceURL := "http://mail-service/send"

	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer((jsonData)))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling mail service"))
		return
	}
	var payload JsonResponse
	payload.Error = false
	payload.Message = "Sent mail to " + msg.To
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, entry LogPayload) {
	err := app.pushToQueue(entry.Name, entry.Data)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Logged via rabbitMQ"
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)

	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.Marshal(&payload)

	err = emitter.Push(string(j), "log.INFO")

	if err != nil {
		return err
	}

	return nil
}
