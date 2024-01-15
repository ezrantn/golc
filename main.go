package main

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"log"
	"net/http"
	"os"
	"text/template"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	r.Get("/", index)
	r.Post("/run", run)

	port := os.Getenv("PORT")

	log.Printf("\033[93mSerpent is started. Press CTRL+C to quit on port %s.\033[0m\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/index.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func run(w http.ResponseWriter, r *http.Request) {
	var prompt struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&prompt); err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	llm, err := openai.NewChat(openai.WithModel(os.Getenv("OPENAI_MODEL")))
	if err != nil {
		log.Printf("Error creating LLM: %v", err)
		http.Error(w, "Error creating LLM: "+err.Error(), http.StatusInternalServerError)
		return
	}

	chatMsg := []schema.ChatMessage{
		schema.SystemChatMessage{Content: "Hello, I am a friendly AI assistant."},
		schema.HumanChatMessage{Content: prompt.Input},
	}

	aimsg, err := llm.Call(context.Background(), chatMsg)
	if err != nil {
		log.Printf("Error calling LLM: %v", err)
		http.Error(w, "Error calling LLM: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if aimsg == nil {
		http.Error(w, "Received nil response from LLM", http.StatusInternalServerError)
		return
	}

	response := struct {
		Input    string `json:"input"`
		Response string `json:"response"`
	}{
		Input:    prompt.Input,
		Response: aimsg.GetContent(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}