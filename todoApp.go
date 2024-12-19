package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var todoItems map[int64]todo

func main() {
	http.HandleFunc("/todo", postTodo)
	port := ":8080"
	todoItems = map[int64]todo{}
	fmt.Printf("starting server in http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func postTodo(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed ", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var list todo
		if err := json.Unmarshal(body, &list); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		id := time.Now().UnixNano()
		list.Id = id
		todoItems[id] = list

		response := map[string]int64{
			"id": id,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else if r.Method == http.MethodGet {
		query := r.URL.Query()
		id := query.Get("id")

		if len(id) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(todoItems)
			return
		}

		intID, _ := strconv.Atoi(id)
		todo, ok := todoItems[int64(intID)]
		if ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(todo)
		} else {
			response := map[string]string{
				"error": "invalid todo id",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	} else if r.Method == http.MethodPut {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed ", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var listdata todo
		if err := json.Unmarshal(body, &listdata); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		todoItems[listdata.Id] = listdata

	} else if r.Method == http.MethodDelete {
		query := r.URL.Query()
		id := query.Get("id")

		if len(id) == 0 {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		intID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		todoID := int64(intID)

		if _, ok := todoItems[todoID]; !ok {
			http.Error(w, "Todo item not found", http.StatusNotFound)
			return
		}
		delete(todoItems, todoID)
		w.WriteHeader(http.StatusNoContent)
	}

}

type todo struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Age  int64  `json:"age"`
	Data string `json:"data"`
}

type newtodo struct {
	Id   int64  `json:"id"`
	Data string `json:"data"`
}
