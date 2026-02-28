package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Event struct {
	ID      int
	Type    string `json:"type"`
	Message string `json:"message"`
	Time    time.Time
}

var mtx = sync.Mutex{}
var eventList = make([]Event, 0)
var id atomic.Int64

func jsonCorrect(e Event) bool {
	if e.Message == "" {
		return false
	}
	if e.Type == "" {
		return false
	}
	return true
}

func addEventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		fmt.Println("Error with decode JSON body!")
		return
	}
	if jsonCorrect(event) {
		mtx.Lock()
		id.Add(1)
		event.ID = int(id.Load())
		event.Time = time.Now()

		eventList = append(eventList, event)

		fmt.Println("Event added!")

		if _, err := w.Write([]byte("Event saved!")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			fmt.Println("Response error:", err)
			return
		}
		mtx.Unlock()
	} else {
		w.WriteHeader(http.StatusBadRequest)

		fmt.Println("Not enough info in JSON!")
		return
	}
}

func eventListHandler(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(eventList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		fmt.Println("Error convertation to JSON:", err)
		return
	}

	if _, err := w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		fmt.Println("Error with response JSON:", err)
		return
	}
}

func typeFilterHandler(w http.ResponseWriter, r *http.Request) {
	filteredEvents := make([]Event, 0)

	typeEvent := r.URL.Query().Get("type")

	mtx.Lock()
	for _, v := range eventList {
		if v.Type == typeEvent {
			filteredEvents = append(filteredEvents, v)
		}
	}
	mtx.Unlock()

	response, err := json.Marshal(filteredEvents)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		fmt.Println("Error convertation to JSON:", err)
		return
	}

	if _, err := w.Write(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		fmt.Println("Response error:", err)
		return
	}
	fmt.Println("Filtered data sent!")
}

func eventClearHandler(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	eventList = []Event{}
	id.Store(0)

	msg := "All events cleared!"
	if _, err := w.Write([]byte(msg)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		fmt.Println("Response error:", err)
		return
	}
	fmt.Println(msg)
	mtx.Unlock()
}

func main() {
	http.HandleFunc("/event/add", addEventHandler)
	http.HandleFunc("/event/list", eventListHandler)
	http.HandleFunc("/event/typeFilter", typeFilterHandler)
	http.HandleFunc("/event/clear", eventClearHandler)

	if err := http.ListenAndServe(":9091", nil); err != nil {
		fmt.Println("Server error!")
		return
	}
}
