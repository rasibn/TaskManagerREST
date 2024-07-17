package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/rasibn/TaskManagerREST/taskstore"
)

type Server struct {
	store *taskstore.TaskStore
}

func NewServer() *Server {
	store := taskstore.New()
	return &Server{store: store}
}

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Contenet-Type", "application/json")
	w.Write(js)
}

func (s *Server) createTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling task create at %s\n", req.URL.Path)

	type RequestTask struct {
		Text string    `json:"text"`
		Tags []string  `json:"tags"`
		Due  time.Time `json:"due"`
	}

	type ResponseId struct {
		Id int `json:"id"`
	}

	// Enforce a JSON Content-Type
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	// DECODING
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var rt RequestTask
	if err := dec.Decode(&rt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// VALID

	id := s.store.CreateTask(rt.Text, rt.Tags, rt.Due)
	renderJSON(w, ResponseId{Id: id})
}

func (s *Server) getAllTasksHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get all tasks at %s\n", req.URL.Path)
	allTasks := s.store.GetAllTasks()
	renderJSON(w, allTasks)
}

func (s *Server) deleteAllTasksHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling tasks by tag at %s\n", req.URL.Path)
	s.store.DeleteAllTasks()
}

// mux.HandleFunc("GET /task/{id}/", server.getTaskHandler)
func (s *Server) getTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get task at %s\n", req.URL.Path)

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	task, err := s.store.GetTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	renderJSON(w, task)
}

// mux.HandleFunc("DELETE /task/{id}/", server.deleteTaskHandler)
func (s *Server) deleteTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling delete task at %s\n", req.URL.Path)

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = s.store.DeleteTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

// mux.HandleFunc("GET /task/{tag}/", server.tagHandler)
func (s *Server) tagHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get tasks by tag at %s\n", req.URL.Path)

	tag := req.PathValue("tag")
	tasks := s.store.GetTasksByTag(tag)
	renderJSON(w, tasks)
}

// mux.HandleFunc("GET /due/{year}/{month}/{day}/", server.dueHandler)
func (s *Server) dueHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get task by due at %s\n", req.URL.Path)

	badRequestError := func() {
		http.Error(w, fmt.Sprintf("expected /due/<year>/<month>/<day>, got %v", req.URL.Path), http.StatusBadRequest)
	}

	year, errYear := strconv.Atoi(req.PathValue("year"))
	month, errMonth := strconv.Atoi(req.PathValue("month"))
	day, errDay := strconv.Atoi(req.PathValue("day"))
	if errYear != nil || errMonth != nil || errDay != nil || month < int(time.January) || month > int(time.December) {
		badRequestError()
		return
	}

	tasks := s.store.GetTasksByDueDate(year, time.Month(month), day)
	renderJSON(w, tasks)
}

func main() {
	mux := http.NewServeMux()
	server := NewServer()

	mux.HandleFunc("POST /task/", server.createTaskHandler)
	mux.HandleFunc("GET /task/", server.getAllTasksHandler)
	mux.HandleFunc("DELETE /task/", server.deleteAllTasksHandler)
	mux.HandleFunc("GET /task/{id}/", server.getTaskHandler)
	mux.HandleFunc("DELETE /task/{id}/", server.deleteTaskHandler)
	mux.HandleFunc("GET /tag/{tag}/", server.tagHandler)
	mux.HandleFunc("GET /due/{year}/{month}/{day}/", server.dueHandler)

	log.Fatal(http.ListenAndServe("localhost:6969", mux))
}
