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

// initialise to load environment variable from .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	r.Get("/", index)
	r.Post("/run", run)
	log.Println("\033[93mBreeze started. Press CTRL+C to quit.\033[0m")
	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}

// index
func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/index.html")
	t.Execute(w, nil)
}

// call the LLM and return the response
func run(w http.ResponseWriter, r *http.Request) {
	prompt := struct {
		Input string `json:"input"`
	}{}
	// decode JSON from client
	err := json.NewDecoder(r.Body).Decode(&prompt)
	if err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	// create the LLM
	llm, err := openai.NewChat(openai.WithModel(os.Getenv("OPENAI_MODEL")))
	if err != nil {
		log.Printf("Error creating LLM: %v", err)
		http.Error(w, "Error creating LLM: "+err.Error(), http.StatusInternalServerError)
		return
	}

	chatmsg := []schema.ChatMessage{
		schema.SystemChatMessage{Content: "Hello, I am a friendly AI assistant."},
		schema.HumanChatMessage{Content: prompt.Input},
	}
	aimsg, err := llm.Call(context.Background(), chatmsg)
	if err != nil {
		log.Printf("Error calling LLM: %v", err)
		http.Error(w, "Error calling LLM: "+err.Error(), http.StatusInternalServerError)
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
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
