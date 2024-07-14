package main

import (
	"encoding/json" // New import
	"net/http"
)

type envelope map[string]any

// Change the data parameter to have the type envelope instead of any.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// JSONMessage represents the structure of the incoming WebSocket messages.
type JSONMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (m *JSONMessage) UnmarshalJSON(data []byte) error {
	type Alias JSONMessage
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	return json.Unmarshal(data, &aux)
}
