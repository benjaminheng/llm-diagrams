package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/benjaminheng/llm-diagrams/anthropic"
)

type PageData struct {
	Input        string
	DiagramURL   string
	PlantUMLCode string
}

const promptTemplate = `Generate a PlantUML diagram based on the following description. 
Only return the PlantUML code without any explanation or additional text.
The code should start with @startuml and end with @enduml.

Description: %s`

func main() {
	// Ensure the temp directory exists
	os.MkdirAll("temp", 0755)

	http.HandleFunc("/", handleIndex)
	http.Handle("/temp/", http.StripPrefix("/temp/", http.FileServer(http.Dir("temp"))))

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		data := PageData{}
		tmpl.Execute(w, data)
	case "POST":
		handleGenerate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	input := r.FormValue("input")
	if input == "" {
		http.Error(w, "Input is required", http.StatusBadRequest)
		return
	}

	// Generate PlantUML code using Anthropic API
	plantUMLCode, err := generatePlantUMLWithAnthropic(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate diagram
	diagramPath, err := generateDiagram(plantUMLCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Input:        input,
		DiagramURL:   diagramPath,
		PlantUMLCode: plantUMLCode,
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, data)
}

func generatePlantUMLWithAnthropic(input string) (string, error) {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	// Create new Anthropic client
	client := anthropic.New(apiKey)

	// Construct the message request
	req := anthropic.MessageRequest{
		Model:     anthropic.ModelClaude35Sonnet, // Using the latest Claude model
		MaxTokens: 1000,
		Messages: []anthropic.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf(promptTemplate, input),
			},
		},
	}

	// Send the request
	resp, err := client.CreateMessage(req)
	if err != nil {
		return "", fmt.Errorf("failed to create message: %w", err)
	}

	// Extract the response text
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return resp.Content[0].Text, nil
}

func generateDiagram(plantUMLCode string) (string, error) {
	dir := "temp"
	// Create a temporary file for the PlantUML code
	tmpFile, err := os.CreateTemp(dir, "diagram-*.puml")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	// Write the PlantUML code to the temporary file
	if _, err := tmpFile.WriteString(plantUMLCode); err != nil {
		return "", err
	}
	tmpFile.Close()

	// Generate output path
	outputFile := strings.ReplaceAll(tmpFile.Name(), ".puml", ".png")

	// Execute PlantUML
	cmd := exec.Command("plantuml", "-tpng", tmpFile.Name())
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to generate diagram: %v", err)
	}

	return "/" + outputFile, nil
}
