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

type Event struct {
	Name string    `yaml:"name"`
	Date time.Time `yaml:"date"`
}

type EventData struct {
	Events []Event `yaml:"events"`
}

type EventView struct {
	Name     string
	DateISO  string
	DateText string
}

type PageData struct {
	NextEvent       *EventView
	FollowingEvents []EventView
}

func loadEvents(path string) (*EventData, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", path, err)
	}

	var eventData EventData
	if err := yaml.Unmarshal(data, &eventData); err != nil {
		return nil, fmt.Errorf("could not unmarshal yaml: %w", err)
	}

	return &eventData, nil
}

func toView(e Event) EventView {
	return EventView{
		Name:     e.Name,
		DateISO:  e.Date.Format(time.RFC3339),
		DateText: e.Date.Format("2006-01-02 15:04"),
	}
}

func countdownHandler(w http.ResponseWriter, r *http.Request) {
	eventData, err := loadEvents("events.yaml")
	if err != nil {
		log.Printf("Error loading events: %v", err)
		http.Error(w, "Could not load events", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	var futureEvents []Event
	for _, event := range eventData.Events {
		if event.Date.After(now) {
			futureEvents = append(futureEvents, event)
		}
	}

	sort.Slice(futureEvents, func(i, j int) bool {
		return futureEvents[i].Date.Before(futureEvents[j].Date)
	})

	var pageData PageData

	if len(futureEvents) > 0 {
		ev := toView(futureEvents[0])
		pageData.NextEvent = &ev
	}

	for _, e := range futureEvents[1:] {
		pageData.FollowingEvents = append(pageData.FollowingEvents, toView(e))
	}

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Could not render page", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, pageData); err != nil {
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
