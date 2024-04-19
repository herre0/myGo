package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"regexp"
	//"sync"
	"time"
	"strings"
	//"github.com/go-chi/chi"
	"github.com/swaggo/http-swagger/v2"
	_ "github.com/swaggo/http-swagger/example/go-chi/docs"
)

type Task struct {
	id int
	title string
	description string
	status string
}

var db *sql.DB
var smallPool chan func()

var (
    WarningLogger *log.Logger
    InfoLogger    *log.Logger
    ErrorLogger   *log.Logger
)

func init() {
    file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }

    InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
    ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}


//	@title			Task API
//	@version		2.0
//	@description	This is a sample api app 

//	@host		localhost:8000
//	@BasePath	
func main() {
	//log.SetOutput(file)
	InfoLogger.Println("Starting the application...")    
	connectToDatabase()
	smallPool = make(chan func(), 20)
	for i := 0; i < 20; i++ {
		go func() {
				for f := range smallPool {
						f()
				}
		}()
	}

	http.HandleFunc("/tasks", getHandler)
	http.HandleFunc("/tasksByPage", getHandlerPagination)
	http.HandleFunc("/addTask", postHandler)
	http.HandleFunc("/updateTask", putHandler)
	http.HandleFunc("/deleteTask", deleteHandler)
	http.HandleFunc("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8000/swagger/doc.json")))
	//r := chi.NewRouter()
	//r := chi.NewRouter()
	// r.Get("/swagger/*", httpSwagger.Handler(
	// 	httpSwagger.URL("http://localhost:1323/swagger/doc.json"), //The url pointing to API definition
	// ))
	//r.Get("", )
    ErrorLogger.Fatal(http.ListenAndServe(":8000", nil))
}



//	@Tags			GET
//	@Summary		Get List of Tasks 
//	@Description	returns task list in json format
//	@Produce		json
//	@Success		200
//	@Router			/tasks [get]
func getHandler(rw http.ResponseWriter, req *http.Request) {
	//wg := sync.WaitGroup{}
	var tasks []Task
	var err error
	//wg.Add(1)
	go func() {
		smallPool <- func() {
			tasks, err = getTaskList()
			if err != nil {
				ErrorLogger.Fatal(err)
			}
			// json.NewEncoder(rw).Encode(tasks)	
			fmt.Println(tasks)			
		}
	}()
	
	time.Sleep(2*time.Second)
	//wg.Wait()
	rw.Write([]byte(fmt.Sprint(tasks)))			
}

func getHandlerPagination(rw http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))

	var tasks []Task
	var err error

	go func() {
		smallPool <- func() {
			tasks, err = getTaskListPagination(page)
			if err != nil {
				ErrorLogger.Fatal(err)
			}
			// json.NewEncoder(rw).Encode(tasks)	
			fmt.Println(tasks)			
		}
	}()
	
	time.Sleep(2*time.Second)
	//wg.Wait()
	rw.Write([]byte(fmt.Sprint(tasks)))			
}

func validatePostRequest(s string)(bool, string){
	// TODO include all special characters with RegexP
	if(strings.Contains(s, "<") || strings.Contains(s, ">") || strings.Contains(s, "!") || strings.Contains(s, "=")|| strings.Contains(s, "#") || strings.Contains(s, "--")){
		return false, "ERROR the json file cannot contain special characters"
	}

	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	taskArr := re.FindAllString(s, -1)
	if(len(taskArr) != 6){
		return false, "ERROR the json file cannot be read"
	}
	title := taskArr[1]	
	description := taskArr[3]	
	status := taskArr[5]	

	if(len(title) > 50 || len(description) > 50 || len(status) > 50){
		return false, "ERROR the fields cannot exceed 50 characters"
	}

	return true, ""
}

//	@Tags			POST
//	@Summary		add a task
//	@Description	adds a task, a task object is required in the body. 
//	@Accept			json
//	@Produce		text/plain
//	@Success		200
//	@Router			/addTask [POST]
func postHandler(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	s := string(body)
	if err != nil {
		http.Error(rw, "ERROR occured while reading the request body", 400)
		return;
	}
	
	passed, message := validatePostRequest(s)
	if !passed {
		http.Error(rw, message, 400)
		return;
	}

	

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
				ErrorLogger.Fatal(err)
			}
		}
	}()	

	time.Sleep(2*time.Second)	
	json.NewEncoder(rw).Encode(taskId)	
}

func validatePutRequest(s string)(bool, string){
	
	// TODO include all special characters with RegexP
	if(strings.Contains(s, "<") || strings.Contains(s, ">") || strings.Contains(s, "!") || strings.Contains(s, "=")|| strings.Contains(s, "#") || strings.Contains(s, "	-")){
		return false, "ERROR the json file cannot contain special characters"
	}

	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	taskArr := re.FindAllString(s, -1)
	if(len(taskArr) != 8){
		return false, "ERROR the json file cannot be read"
	}
	id := taskArr[1]	
	title := taskArr[3]	
	description := taskArr[5]	
	status := taskArr[7]	

	if _, err := strconv.Atoi(id); err != nil {
		return false, "Id must be a valid number"
	}

	if(len(title) > 50 || len(description) > 50 || len(status) > 50){
		return false, "ERROR the fields cannot exceed 50 characters"
	}

	return true, ""
}

//	@Tags			PUT
//	@Summary		Update a task by Id 
//	@Description	Update any task by providing a new Task in the body and an id in the parameters
//	@Produce		text/plain
//	@Success		200
//	@Router			/updateTask [PUT]
func putHandler(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "ERROR occured while reading the request body", 400)
		return; 
	}	

	s := string(body)	
	passed, message := validatePutRequest(s)
	if !passed {
		http.Error(rw, message, 400)
		return;
	}
	
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
				ErrorLogger.Fatal(err)
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


//	@Tags			DELETE
//	@Summary		Delete a task by Id 
//	@Description	Delete any task by an id parameter
//	@Produce		text/plain
//	@Success		200
//	@Router			/deleteTask [DELETE]
func deleteHandler(rw http.ResponseWriter, req *http.Request) {
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
        ErrorLogger.Fatal(err)
    }

    pingErr := db.Ping()
    if pingErr != nil {
        ErrorLogger.Fatal(pingErr)
    }

    InfoLogger.Println("Connected to DB")
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

func getTaskListPagination(page int) ([]Task, error) {
    var tasks []Task
	limit := 5
	offset := limit * (page-1) // 0*5=0, 1*5=5 ..
	// avoids SQL Injection
	sqlstmt := fmt.Sprintf("SELECT * FROM task limit %d offset %d", limit, offset)
	rows, err := db.Query(sqlstmt)
    if err != nil {
        return nil, fmt.Errorf("An error occured: %v", err)
    }
    defer rows.Close()
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

// func getTaskById(id int64) (Task, error) {
//     var task Task

//     row := db.QueryRow("SELECT * FROM task WHERE id = ?", id)
//     if err := row.Scan(&task.id, &task.title, &task.description, &task.status); err != nil {
//         if err == sql.ErrNoRows {
//             return task, fmt.Errorf("getTaskById %d: no such a task", id)
//         }
//         return task, fmt.Errorf("getTaskById %d: %v", id, err)
//     }
//     return task, nil
// }

func addTask(task Task) (int64, error) {
    result, err := db.Exec("INSERT INTO task (title, description, status) VALUES (?, ?, ?)", task.title, task.description, task.status)
    if err != nil {
		ErrorLogger.Println("addTask: %v", err)
        return 0, fmt.Errorf("addTask: %v", err)
    }
    id, err := result.LastInsertId()
    if err != nil {
		ErrorLogger.Println("addTask: %v", err)
        return 0, fmt.Errorf("addTask: %v", err)
    }
    return id, nil
}

func updateTask(task Task) (int64, error) {
    result, err := db.Exec("Update task set title = ?, description = ?, status = ? where id = ?", task.title, task.description, task.status, task.id)
    if err != nil {
		ErrorLogger.Println("updateTask: %v", err)
        return 0, fmt.Errorf("updateTask: %v", err)
    }
    rows, err := result.RowsAffected()
    if err != nil {
		ErrorLogger.Println("updateTask: %v", err)
        return 0, fmt.Errorf("updateTask: %v", err)
    }
    return rows, nil    
}

func deleteTask(id int) (int64, error) {	

	if id < 0 {
		fmt.Println("ID must be bigger than 0")
		return 0, nil;
	}
    result, err := db.Exec("delete from task where id = ?", id)
    if err != nil {
		ErrorLogger.Println("deleteTask: %v", err)
        return 0, fmt.Errorf("deleteTask: %v", err)
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
		ErrorLogger.Println("deleteTask: %v", err)
        return 0, fmt.Errorf("deleteTask: %v", err)
    }
    return rows, nil
}
