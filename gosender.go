package gosender

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	strip "github.com/grokify/html-strip-tags-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Payload represents the request payload structure.
type Payload struct {
	Credentials json.RawMessage `json:"credentials"`
	Token       json.RawMessage `json:"token"`
	MessageBody string          `json:"messageBody"`
}

// ErrorResponse represents an error response structure.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SendResponse represents a successful send response structure.
type SendResponse struct {
	Token  string         `json:"token"`
	Output *gmail.Message `json:"output"`
}

// handleRequest handles the HTTP request to send an email.
func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request. Failed to parse form.", http.StatusBadRequest)
		return
	}

	payloadStr := r.FormValue("payload")
	if payloadStr == "" {
		http.Error(w, "Bad request. Payload not provided.", http.StatusBadRequest)
		return
	}

	payload, err := decodePayload(payloadStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bad request. %s", err.Error()), http.StatusBadRequest)
		return
	}

	client, err := getClient(payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error. %s", err.Error()), http.StatusInternalServerError)
		return
	}
	ctx := context.Background()
	service, err := gmail.NewService(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error. %s", err.Error()), http.StatusInternalServerError)
		return
	}

	message := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(payload.MessageBody)),
	}

	sendResponse, err := service.Users.Messages.Send("me", message).Do()
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error. %s", err.Error()), http.StatusInternalServerError)
		return
	}

	err = trashExistingMessages(service, "INBOX")
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error. %s", err.Error()), http.StatusInternalServerError)
		return
	}

	err = trashExistingMessages(service, "SPAM")
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error. %s", err.Error()), http.StatusInternalServerError)
		return
	}

	token, err := getToken(client)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error. %s", err.Error()), http.StatusInternalServerError)
		return
	}

	response := SendResponse{
		Token:  token,
		Output: sendResponse,
	}

	json.NewEncoder(w).Encode(response)
}

// decodePayload decodes the payload string and returns a Payload object.
func decodePayload(payloadStr string) (*Payload, error) {
	decoded, err := base64.StdEncoding.DecodeString(payloadStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %v", err)
	}

	var payload Payload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	return &payload, nil
}

// getClient returns an authenticated HTTP client using the provided payload.
func getClient(payload *Payload) (*http.Client, error) {
	credentials := strip.StripTags(string(payload.Credentials))
	token := strip.StripTags(string(payload.Token))

	config, err := google.ConfigFromJSON([]byte(credentials), gmail.MailGoogleComScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %v", err)
	}

	tokenSource := config.TokenSource(context.TODO(), &oauth2.Token{
		AccessToken: strip.StripTags(token),
	})

	return oauth2.NewClient(context.Background(), tokenSource), nil
}

// getToken returns the access token as a string from the HTTP client.
func getToken(client *http.Client) (string, error) {
	token, err := client.Transport.(*oauth2.Transport).Source.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %v", err)
	}

	return string(tokenJSON), nil
}

// trashExistingMessages moves existing messages in the specified label to the trash.
func trashExistingMessages(service *gmail.Service, labelID string) error {
	messages, err := service.Users.Messages.List("me").LabelIds(labelID).Do()
	if err != nil {
		return fmt.Errorf("failed to list messages: %v", err)
	}

	for _, message := range messages.Messages {
		_, err := service.Users.Messages.Trash("me", message.Id).Do()
		if err != nil {
			return fmt.Errorf("failed to trash message: %v", err)
		}
	}

	return nil
}

// gosender starts the web server and handles the "/send" endpoint.
func gosender() {
	http.HandleFunc("/send", handleRequest)
	http.ListenAndServe(":8080", nil)
}
