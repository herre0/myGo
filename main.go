package main

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
    "os"
	"github.com/go-sql-driver/mysql"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"regexp"
)

type Task struct {
	id int
	title string
	description string
	status string
}
var db *sql.DB

func main() {
	fmt.Println("server is running")
	connectToDatabase()
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
        switch req.Method {
			case "GET":
				tasks, err := getTaskList()
				if err != nil {
					log.Fatal(err)
				}
				// json.NewEncoder(rw).Encode(tasks)	
				rw.Write([]byte(fmt.Sprint(tasks)))				
			case "POST":
				body, err := ioutil.ReadAll(req.Body)
				s := string(body)
				re := regexp.MustCompile(`[a-zA-Z0-9]+`)
				taskArr := re.FindAllString(s, -1)
				taskId, err := addTask(Task{
					 	title:  taskArr[1],
					 	description: taskArr[3],
					 	status:  taskArr[5],
					 })
				if err != nil {
					log.Fatal(err)
				}
				json.NewEncoder(rw).Encode(taskId)
			case "DELETE":				
				// http://localhost:8000?id=1
				query := req.URL.Query()
				// convert string into int
				id, _ := strconv.Atoi(query.Get("id"))
				deleteTask(id)
				
				rw.Write([]byte("successfully deleted!"))
			case "PUT":		
				body, err := ioutil.ReadAll(req.Body)
				s := string(body)
				re := regexp.MustCompile(`[a-zA-Z0-9]+`)
				taskArr := re.FindAllString(s, -1)
				id, _ := strconv.Atoi(taskArr[1])
				taskId, err := updateTask(Task{
					id:  id,
					title:  taskArr[3],
					description: taskArr[5],
					status:  taskArr[7],
				})
				if err != nil {
					log.Fatal(err)
				}
				json.NewEncoder(rw).Encode(taskId)
		}
    })

	task, err := getTaskById(2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Task:\n", task)

	// taskId, err := addTask(Task{
	// 	title:  "test1",
	// 	description: "test",
	// 	status:  "todo",
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("ID of added task: %v\n", taskId)

	rows2, err := updateTask(Task{
		id:  6,
		title:  "test33333",
		description: "test233333",
		status:  "todooo",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Rows affected for Update: %v\n", rows2)

	rows, err := deleteTask(5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Rows affected for Deletion: %v\n", rows)

	
    log.Fatal(http.ListenAndServe(":8000", nil))
}

func connectToDatabase() {
	// Capture connection properties.
    cfg := mysql.Config{
        User:   os.Getenv("DBUSER"),
        Passwd: os.Getenv("DBPASS"),
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        DBName: "task",
    }
    // Get a database handle.
    var err error
    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
        log.Fatal(err)
    }

    pingErr := db.Ping()
    if pingErr != nil {
        log.Fatal(pingErr)
    }

    fmt.Println("Connected!")
}

func getTaskList() ([]Task, error) {
    // tasks to hold data from returned rows.
    var tasks []Task

    rows, err := db.Query("SELECT * FROM task")
    if err != nil {
        return nil, fmt.Errorf("An error occured: %v", err)
    }
    defer rows.Close()
    // Loop through rows, using Scan to assign column data to struct fields.
    for rows.Next() {
        var task Task
        if err := rows.Scan(&task.id, &task.title, &task.description, &task.status); err != nil {
            return nil, fmt.Errorf("An error occured: %v", err)
        }
        tasks = append(tasks, task)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("An error occured: %v", err)
    }
	
    return tasks, nil
}

func getTaskById(id int64) (Task, error) {
    var task Task

    row := db.QueryRow("SELECT * FROM task WHERE id = ?", id)
    if err := row.Scan(&task.id, &task.title, &task.description, &task.status); err != nil {
        if err == sql.ErrNoRows {
            return task, fmt.Errorf("getTaskById %d: no such a task", id)
        }
        return task, fmt.Errorf("getTaskById %d: %v", id, err)
    }
    return task, nil
}

func addTask(task Task) (int64, error) {
    result, err := db.Exec("INSERT INTO task (title, description, status) VALUES (?, ?, ?)", task.title, task.description, task.status)
    if err != nil {
        return 0, fmt.Errorf("addTask: %v", err)
    }
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("addTask: %v", err)
    }
    return id, nil
}

func updateTask(task Task) (int64, error) {
    result, err := db.Exec("Update task set title = ?, description = ?, status = ? where id = ?", task.title, task.description, task.status, task.id)
    if err != nil {
        return 0, fmt.Errorf("updateTask: %v", err)
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return 0, fmt.Errorf("updateTask: %v", err)
    }
    return rows, nil    
}

func deleteTask(id int) (int64, error) {
    result, err := db.Exec("delete from task where id = ?", id)
    if err != nil {
        return 0, fmt.Errorf("deleteTask: %v", err)
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
        return 0, fmt.Errorf("deleteTask: %v", err)
    }
    return rows, nil
}
