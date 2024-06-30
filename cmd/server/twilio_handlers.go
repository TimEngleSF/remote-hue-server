package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (app *application) twilioWebHookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive Text Handler")

	// Read the body of the request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	// Parse the form data manually
	formData, err := url.ParseQuery(string(body))
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	// Extract the necessary fields
	from := formData.Get("From")
	bodyText := formData.Get("Body")

	// Print the extracted fields
	fmt.Printf("From: %s\n", from)
	fmt.Printf("Body: %s\n", bodyText)

	if from == app.config.userPhoneNumber {
		// Call the OpenAI API
		gptResponse, err := app.openai.TranformTextBodyToJSON(bodyText)
		if err != nil {
			http.Error(w, "Unable to process text", http.StatusBadRequest)
		}
		fmt.Println(gptResponse)
	} else {
		app.logger.Error("received a text message from an unauthorized number", "unauthorized_number", from)
	}

	// Respond with a status 200 OK
	w.WriteHeader(http.StatusOK)
}
