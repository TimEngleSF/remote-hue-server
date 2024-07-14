package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/TimEngleSF/remote-hue-server/internal/service"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

// GPTStatusRequest represents the structure of the status request from GPT.
type GPTStatusRequest struct {
	Data struct {
		Rooms service.GroupNames `json:"room"`
	} `json:"data"`
}

// twilioWebHookHandler handles incoming requests from Twilio's webhook.
func (app *application) twilioWebHookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive Text Handler")

	// Read the body of the request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	// Parse incoming data from Twilio
	formData, err := url.ParseQuery(string(body))
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	from := formData.Get("From")
	bodyText := formData.Get("Body")

	if from != app.config.userPhoneNumber {
		app.logger.Error("received a text message from an unauthorized number", "unauthorized_number", from)
		w.WriteHeader(http.StatusOK)
		return
	}

	systemRoleMessage := service.SystemRoleMessage(*app.groupsState, app.groupNames)
	var jsonMessage JSONMessage

	// Call the OpenAI API
	gptResponse, err := app.openai.TranformTextBodyToJSON(systemRoleMessage, bodyText)
	if err != nil {
		app.logError(r, err)
		return
	}

	// Unmarshal the response from OpenAI into JSONMessage
	//  err = json.Unmarshal([]byte(gptResponse), &jsonMessage)
	err = jsonMessage.UnmarshalJSON([]byte(gptResponse))
	if err != nil {
		app.logError(r, err)
		return
	}

	// Process request based on type
	switch jsonMessage.Type {
	case "status":
		app.handleStatusRequest(jsonMessage)
	case "update":
		fmt.Println("Update!")
	default:
		app.logger.Error("received a text message with an unknown type", "type", jsonMessage.Type)
	}

	// Respond with a status 200 OK
	w.WriteHeader(http.StatusOK)
}

// handleStatusRequest processes the status request from the JSON message.
func (app *application) handleStatusRequest(jsonMsg JSONMessage) {
	// Update group state field
	err := app.SetGroupsStateField()
	if err != nil {
		app.logger.Error("error getting groups state", "error", err)
		return
	}

	// Unmarshal JSON message data into GPTStatusRequest struct
	var statusRequest GPTStatusRequest
	data, err := json.Marshal(jsonMsg.Data)
	if err != nil {
		app.logger.Error("error marshalling json message")
		return
	}

	err = json.Unmarshal(data, &statusRequest.Data)
	if err != nil {
		app.logger.Error("error unmarshalling json message")
		return
	}

	fmt.Println("DATA!!!!", string(data))

	// Prepare and send the Twilio message
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(app.config.userPhoneNumber)
	params.SetFrom(app.config.twilioPhoneNumber)
	params.SetBody(app.groupsState.GroupStatusMessage(service.GroupNames(statusRequest.Data.Rooms)))

	_, err = app.twilio.Api.CreateMessage(params)
	if err != nil {
		app.logger.Error("error sending Twilio message", "error", err)
	}
}
