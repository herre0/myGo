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
	//"sync"
	"time"
)

type Task struct {
	id int
	title string
	description string
	status string
}
var db *sql.DB
var smallPool chan func()

func main() {
	fmt.Println("server is running")
	connectToDatabase()
	smallPool = make(chan func(), 20)
	for i := 0; i < 20; i++ {
		go func() {
				for f := range smallPool {
						f()
				}
		}()
	}


	http.HandleFunc("/", getHandler)
	http.HandleFunc("/addTask", postHandler)
	http.HandleFunc("/updateTask", putHandler)
	http.HandleFunc("/deleteTask", deleteHandler)

    log.Fatal(http.ListenAndServe(":8000", nil))
}

func getHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("get Handler is working")
	//wg := sync.WaitGroup{}
	var tasks []Task
	var err error
	//wg.Add(1)
	go func() {
		smallPool <- func() {
			tasks, err = getTaskList()
			if err != nil {
				log.Fatal(err)
			}
			// json.NewEncoder(rw).Encode(tasks)	
			fmt.Println(tasks)			
		}
	}()
	
	time.Sleep(2*time.Second)
	//wg.Wait()
	rw.Write([]byte(fmt.Sprint(tasks)))			
}

func postHandler(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	s := string(body)
	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	taskArr := re.FindAllString(s, -1)
	var taskId int64
	go func() {
		smallPool <- func() {
			taskId, err = addTask(Task{
				title:  taskArr[1],
				description: taskArr[3],
				status:  taskArr[5],
				})
			if err != nil {
				log.Fatal(err)
			}
		}
	}()	

	time.Sleep(2*time.Second)	
	json.NewEncoder(rw).Encode(taskId)	
}

func putHandler(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	s := string(body)
	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	taskArr := re.FindAllString(s, -1)
	id, _ := strconv.Atoi(taskArr[1])
	var taskId int64
	go func() {
		smallPool <- func() {
			taskId, err = updateTask(Task{
				id:  id,
				title:  taskArr[3],
				description: taskArr[5],
				status:  taskArr[7],
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	time.Sleep(2*time.Second)
	json.NewEncoder(rw).Encode(taskId)
}

func deleteHandler(rw http.ResponseWriter, req *http.Request) {
	// http://localhost:8000?id=1
	query := req.URL.Query()
	// convert string into int
	id, _ := strconv.Atoi(query.Get("id"))
	
	go func() {
		deleteTask(id)
	}()

	time.Sleep(2*time.Second)
	rw.Write([]byte("successfully deleted!"))
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
