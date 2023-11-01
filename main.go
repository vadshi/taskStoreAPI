/*
// Пример REST сервера с несколькими маршрутами(используем только стандартную библиотеку)

// POST   /task/              :  создаёт задачу и возвращает её ID
// GET    /task/<taskid>      :  возвращает одну задачу по её ID
// GET    /task/              :  возвращает все задачи
// DELETE /task/<taskid>      :  удаляет задачу по ID
// DELETE /task/              :  удаляет все задачи
// GET    /tag/<tagname>      :  возвращает список задач с заданным тегом
// GET    /due/<yy>/<mm>/<dd> :  возвращает список задач, запланированных на указанную дату

Структура проекта
https://github.com/golang-standards/project-layout/blob/master/README_ru.md
*/

package main

import (
	"TaskStoreAPI/internal/taskstore"
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var (
	port  string = "8080"
	store *taskstore.TaskStore
)

type ErrorMessage struct {
	Message string `json:"message"`
}

/*
type taskServer struct {
	store *taskstore.TaskStore
}

func NewTaskServer() *taskServer {
	store := taskstore.New()
	return &taskServer{store: store}
}

func (ts *taskServer) taskHandler(w http.ResponseWriter, r *http.Request) {
	//Request is only '/task/' URL without ID
	if r.URL.Path == "/task/" {
		if r.Method == http.MethodPost {
			ts.createTaskHandler(w, r)
		} else if r.Method == http.MethodGet {
			ts.getAllTaskHandler(w, r)
		} else if r.Method == http.MethodDelete {
			ts.deleteAllTaskHandler(w, r)
		} else {
			http.Error(w, fmt.Sprintf("expect method GET, POST, DELETE at '/task', got %v", r.Method), http.StatusMethodNotAllowed)
			return
		}

	} else {
		// Request has an ID as '/task/<id>' URL
		path := strings.Trim(r.URL.Path, "/")
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			http.Error(w, "expect 'task/<id>' in task handler", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(pathParts[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodGet {
			ts.getTaskHandler(w, r, int(id))
		} else if r.Method == http.MethodDelete {
			ts.deleteTaskHandler(w, r, int(id))
		} else {
			http.Error(w, fmt.Sprintf("expect method GET, DELETE at '/task<id>', got %v", r.Method), http.StatusMethodNotAllowed)
			return
		}
	}
}
*/

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling task create at %s\n", r.URL.Path)

	// Структура для создания задачи
	type RequestTask struct {
		Text string    `json:"text"`
		Tags []string  `json:"tags"`
		Due  time.Time `json:"due"`
	}

	// Для ответа в виде JSON
	type ResponseId struct {
		Id int `json:"id"`
	}

	// JSON в качестве Content-Type
	contentType := r.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	// Обработка тела запроса
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var rt RequestTask
	if err := dec.Decode(&rt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Создаем новую задачу в хранилище и получаем ее <id>
	id := store.CreateTask(rt.Text, rt.Tags, rt.Due)

	// Создаем json для ответа
	js, err := json.Marshal(ResponseId{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // код ошибки 500
		return
	}

	// Обязательно вносим изменения в Header до вызова метода Write()!
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func getAllTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling get all tasks at %s\n", r.URL.Path)

	allTasks := store.GetAllTasks()

	js, err := json.Marshal(allTasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // код ошибки 500
		return
	}

	// Обязательно вносим изменения в Header до вызова метода Write()!
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling get task at %s\n", r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	//Считаем id из строки запроса и конвертируем его в int
	vars := mux.Vars(r) // {"id" : "12"}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("client trying to use invalid id param:", err)
		msg := ErrorMessage{Message: "do not use ID not supported int casting"}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(msg)
		return
	}
	log.Println("Task id #:", id)

	task, err := store.GetTask(id)
	if err != nil {
		msg := ErrorMessage{Message: "id not found"}
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(msg)
		return
	}
	js, err := json.Marshal(task)
	if err != nil {
		msg := ErrorMessage{Message: "Internal Server Error"}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(msg)
		return
	}

	w.Write(js)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling delete task at %s\n", r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	//Считаем id из строки запроса и конвертируем его в int
	vars := mux.Vars(r) // {"id" : "12"}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("client trying to use invalid id param:", err)
		msg := ErrorMessage{Message: "do not use ID not supported int casting"}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(msg)
		return
	}
	log.Println("Task id #:", id)

	err = store.DeleteTask(id)
	if err != nil {
		msg := ErrorMessage{Message: "id not found"}
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(msg)
		return
	}
}

func deleteAllTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling delete all tasks at %s\n", r.URL.Path)

	store.DeleteAllTasks()
}

func getTagkHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling get task at %s\n", r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	//Считаем id из строки запроса и конвертируем его в int
	vars := mux.Vars(r) // {"id" : "12"}
	name := vars["name"]

	log.Println("Task tag is #:", name)

	task, err := store.GetTag(name)
	if err != nil {
		msg := ErrorMessage{Message: "id not found"}
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(msg)
		return
	}
	js, err := json.Marshal(task)
	if err != nil {
		msg := ErrorMessage{Message: "Internal Server Error"}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(msg)
		return
	}

	w.Write(js)
}

func getDueHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling get task at %s\n", r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	//Считаем id из строки запроса и конвертируем его в int
	vars := mux.Vars(r) // {"id" : "12"}
	_, err := strconv.Atoi(vars["yy"])
	if err != nil {
		log.Println("client trying to use invalid id param:", err)
		msg := ErrorMessage{Message: "do not use ID not supported int casting"}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(msg)
		return
	}
	_, err = strconv.Atoi(vars["mm"])
	if err != nil {
		log.Println("client trying to use invalid id param:", err)
		msg := ErrorMessage{Message: "do not use ID not supported int casting"}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(msg)
		return
	}
	_, err = strconv.Atoi(vars["dd"])
	if err != nil {
		log.Println("client trying to use invalid id param:", err)
		msg := ErrorMessage{Message: "do not use ID not supported int casting"}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(msg)
		return
	}

	date := "20" + vars["yy"] + "-" + vars["mm"] + "-" + vars["dd"]

	log.Println("Task date #:", date)

	task, err := store.GetDue(date)
	if err != nil {
		msg := ErrorMessage{Message: "id not found"}
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(msg)
		return
	}
	js, err := json.Marshal(task)
	if err != nil {
		msg := ErrorMessage{Message: "Internal Server Error"}
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(msg)
		return
	}

	w.Write(js)
}

func main() {
	log.Println("Trying to start REST API Taskstore!")

	store = taskstore.New()
	router := mux.NewRouter()
	// Строгий конечный '/'
	router.StrictSlash(true)

	router.HandleFunc("/task", createTaskHandler).Methods("POST")
	router.HandleFunc("/task/{id:[0-9]+}", getTaskHandler).Methods("GET")
	router.HandleFunc("/task", getAllTaskHandler).Methods("GET")
	router.HandleFunc("/task/{id:[0-9]+}", deleteTaskHandler).Methods("DELETE")
	router.HandleFunc("/task", deleteAllTaskHandler).Methods("DELETE")
	router.HandleFunc("/tag/{name}", getTagkHandler).Methods("GET")
	router.HandleFunc("/due/{yy:[0-9]{2}}/{mm:[0-9]{2}}/{dd:[0-9]{2}}", getDueHandler).Methods("GET")

	// GET    /tag/<tagname>      :  возвращает список задач с заданным тегом
	// GET    /due/<yy>/<mm>/<dd> :  возвращает список задач, запланированных на указанную дату

	log.Println("Router configured successfully! Let's go!")
	log.Fatal(http.ListenAndServe(":"+port, router))
}
