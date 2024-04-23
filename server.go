package main

import (
	"context"
	"encoding/json"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"html/template"
	"net/http"
	"os"
	"time"
)

var openaiClient *openai.Client

func main() {
	// Initialize OpenAI client
	openaiAPIKey := "YOUR_OPEN_API_KEY"
	if openaiAPIKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is not set")
		return
	}
	openaiClient = openai.NewClient(openaiAPIKey)

	// Define HTTP routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/message", messageHandler)

	// Start the HTTP server
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(w, "home.html", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	t, err := template.ParseFiles("frontend/templates/" + tmpl)
	if err != nil {
		return err
	}
	err = t.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON message from the request body
	var requestBody struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Failed to decode JSON message", http.StatusBadRequest)
		return
	}

	// Process the user's message using ChatGPT
	response, err := processUserMessage(requestBody.Message)
	if err != nil {
		http.Error(w, "Failed to process user message", http.StatusInternalServerError)
		return
	}

	// Send the response back to the user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func processUserMessage(userMessage string) (string, error) {
	logMessageToFile(time.Now().Format("2006-01-02 15:04:05") + ". User: " + userMessage)
	// Call OpenAI API to generate response using ChatGPT
	resp, err := openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	logMessageToFile(time.Now().Format("2006-01-02 15:04:05") + ". Server: " + resp.Choices[0].Message.Content)
	// Extract and return the response from ChatGPT
	return resp.Choices[0].Message.Content, nil
}

func logMessageToFile(message string) {
	file, err := os.OpenFile("history.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening history file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(message)
	if err != nil {
		fmt.Println("Error writing to history file:", err)
		return
	}
}
