package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv" //for .env file
	_ "github.com/lib/pq"      // Import pq driver
)

var db *sql.DB

func init() {
	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
func main() {
	// Connect to the PostgreSQL database
	connStr := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer db.Close()

	// Ensure the database connection works
	if err := db.Ping(); err != nil {
		fmt.Println("Error pinging the database:", err)
		return
	}
	fmt.Println("Connected to the database")

	http.HandleFunc("/todo", postTodo)
	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func postTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed reading request body", http.StatusInternalServerError)
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

		// Insert todo into the database
		query := `INSERT INTO todos (id, name, age, data) VALUES ($1, $2, $3, $4)`
		_, err = db.Exec(query, list.Id, list.Name, list.Age, list.Data)
		if err != nil {
			http.Error(w, "Failed to insert todo", http.StatusInternalServerError)
			return
		}

		response := map[string]int64{
			"id": id,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else if r.Method == http.MethodGet {
		query := r.URL.Query()
		id := query.Get("id")

		if len(id) == 0 {
			// Fetch all todos from the database
			rows, err := db.Query("SELECT id, name, age, data FROM todos")
			if err != nil {
				http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var todos []todo
			for rows.Next() {
				var t todo
				if err := rows.Scan(&t.Id, &t.Name, &t.Age, &t.Data); err != nil {
					http.Error(w, "Error scanning todos", http.StatusInternalServerError)
					return
				}
				todos = append(todos, t)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(todos)
		} else {
			intID, _ := strconv.Atoi(id)
			var t todo
			err := db.QueryRow("SELECT id, name, age, data FROM todos WHERE id = $1", int64(intID)).Scan(&t.Id, &t.Name, &t.Age, &t.Data)
			if err != nil {
				http.Error(w, "Todo not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(t)
		}
	} else if r.Method == http.MethodPut {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var listdata todo
		if err := json.Unmarshal(body, &listdata); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Update todo in the database
		query := `UPDATE todos SET name = $2, age = $3, data = $4 WHERE id = $1`
		_, err = db.Exec(query, listdata.Id, listdata.Name, listdata.Age, listdata.Data)
		if err != nil {
			http.Error(w, "Failed to update todo", http.StatusInternalServerError)
			return
		}
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

		// Delete todo from the database
		deleteQuery := `DELETE FROM todos WHERE id = $1`
		_, err = db.Exec(deleteQuery, int64(intID))
		if err != nil {
			http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

type todo struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Age  int64  `json:"age"`
	Data string `json:"data"`
}
