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
	"strings"
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

func validatePostRequest(req *http.Request)(bool, string){
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
		return false, "ERROR occured while reading the request body"
	}	
	
	// TODO include all special characters with RegexP
	s := string(body)
	if(strings.Contains(s, "<") || strings.Contains(s, ">") || strings.Contains(s, "!") || strings.Contains(s, "=")|| strings.Contains(s, "#") || strings.Contains(s, "--")){
		log.Fatal(err)
		return false, "ERROR the json file cannot contain special characters"
	}

	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	taskArr := re.FindAllString(s, -1)
	if(len(taskArr) != 6){
		log.Fatal(err)
		return false, "ERROR the json file cannot be read"
	}
	title := taskArr[1]	
	description := taskArr[3]	
	status := taskArr[5]	

	if(len(title) > 50 || len(description) > 50 || len(status) > 50){
		log.Fatal(err)
		return false, "ERROR the fields cannot exceed 50 characters"
	}

	return true, ""
}

func postHandler(rw http.ResponseWriter, req *http.Request) {
	passed, message := validatePostRequest(req)
	if !passed {
		http.Error(rw, message, 400)
		return;
	}

	body, err := ioutil.ReadAll(req.Body)
	
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

func validatePutRequest(req *http.Request)(bool, string){
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
		return false, "ERROR occured while reading the request body"
	}	
	
	// TODO include all special characters with RegexP
	s := string(body)
	if(strings.Contains(s, "<") || strings.Contains(s, ">") || strings.Contains(s, "!") || strings.Contains(s, "=")|| strings.Contains(s, "#") || strings.Contains(s, "--")){
		log.Fatal(err)
		return false, "ERROR the json file cannot contain special characters"
	}

	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	taskArr := re.FindAllString(s, -1)
	if(len(taskArr) != 8){
		log.Fatal(err)
		return false, "ERROR the json file cannot be read"
	}
	id := taskArr[1]	
	title := taskArr[3]	
	description := taskArr[5]	
	status := taskArr[7]	

	if _, err := strconv.Atoi(id); err != nil {
		log.Fatal(err)
		return false, "Id must be a valid number"
	}

	if(len(title) > 50 || len(description) > 50 || len(status) > 50){
		log.Fatal(err)
		return false, "ERROR the fields cannot exceed 50 characters"
	}

	return true, ""
}

func putHandler(rw http.ResponseWriter, req *http.Request) {
	passed, message := validatePutRequest(req)
	if !passed {
		http.Error(rw, message, 400)
		return;
	}

	body, err := ioutil.ReadAll(req.Body)
	s := string(body)
	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	taskArr := re.FindAllString(s, -1)
	id, _ := strconv.Atoi(taskArr[1])

	var rows int64
	go func() {
		smallPool <- func() {
			rows, err = updateTask(Task{
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
	if(rows < 1) {
		http.Error(rw, "Id doesn't exist", 400)
		return;
	}
	json.NewEncoder(rw).Encode(rows)
}

func deleteHandler(rw http.ResponseWriter, req *http.Request) {
	// http://localhost:8000?id=1
	query := req.URL.Query()
	// convert string into int
	id, _ := strconv.Atoi(query.Get("id"))
	if id < 0 {
		http.Error(rw, "Id must be a valid number", 400)
		return;
	}
	var rows int64

	go func() {
		rows, _ = deleteTask(id)
	}()

	time.Sleep(2*time.Second)
	if(rows < 1) {
		http.Error(rw, "Id doesn't exist", 400)
		return;
	}
		
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

    fmt.Println("Connected To DB!")
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
