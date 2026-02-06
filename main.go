package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"gopkg.in/yaml.v2"
)

// Event represents a single event with ixwts name and date.
type Event struct {
	Name string    `yaml:"name"`
	Date time.Time `yaml:"date"`
}

// EventData holds all events read from the YAML file.
type EventData struct {
	Events []Event `yaml:"events"`
}

// PageData is the data structure passed to the HTML template.
type PageData struct {
	NextEvent       Event
	FollowingEvents []Event
}

// loadEvents reads and parses the events.yaml file.
func loadEvents(path string) (*EventData, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", path, err)
	}

	var eventData EventData
	err = yaml.Unmarshal(data, &eventData)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal yaml from %s: %w", path, err)
	}

	return &eventData, nil
}

// countdownHandler handles the main page logic.
func countdownHandler(w http.ResponseWriter, r *http.Request) {
	// Load events from YAML file.
	eventData, err := loadEvents("events.yaml")
	if err != nil {
		log.Printf("Error loading events: %v", err)
		http.Error(w, "Could not load events", http.StatusInternalServerError)
		return
	}

	// Filter for future events.
	now := time.Now()
	var futureEvents []Event
	for _, event := range eventData.Events {
		if event.Date.After(now) {
			futureEvents = append(futureEvents, event)
		}
	}

	// Sort future events by date.
	sort.Slice(futureEvents, func(i, j int) bool {
		return futureEvents[i].Date.Before(futureEvents[j].Date)
	})

	// Prepare data for the template.
	pageData := PageData{}
	if len(futureEvents) > 0 {
		pageData.NextEvent = futureEvents[0]
	}
	if len(futureEvents) > 1 {
		pageData.FollowingEvents = futureEvents[1:]
	}

	// Parse and execute the template.
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Could not render page", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, pageData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Could not render page", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", countdownHandler)

	log.Println("Starting server on :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
