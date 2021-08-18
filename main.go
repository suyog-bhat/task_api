package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type task struct {
	//gorm.Model
	Key   int    `json:"key" gorm:"primaryKey"`
	Title string `json:"title"`
	Note  string `json:"note"`
}

func changeResponseContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func returnTaskAll(w http.ResponseWriter, r *http.Request) {
	fmt.Println("return all task endpoint")
	//w.Header().Add("content-type", "application/json")
	var t []task
	res := db.Find(&t)
	if res.RowsAffected == 0 {
		fmt.Fprintf(w, "no task found")
	} else {
		json.NewEncoder(w).Encode(&t)
		fmt.Println(t)
	}

}

func returnTaskId(w http.ResponseWriter, r *http.Request) {
	fmt.Println("single task fetching endpoint")
	vars := mux.Vars(r)
	key, _ := strconv.Atoi(vars["id"])

	var t task
	fmt.Printf("%d,%T", key, key)
	res := db.First(&t, key)
	fmt.Println(res)
	//fmt.Println(res.RowsAffected)
	if res.RowsAffected == 0 {
		//fmt.Fprintf(w,"no task found with id : %d",key)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - task not found"))

	} else {
		marshedTask, err := json.Marshal(t)
		if err == nil {
			fmt.Println(t)
			w.Write(marshedTask)
			fmt.Printf("task fetched:%v", t)
		}
	}

	// json.NewEncoder(w).Encode(t)

}

func createTask(w http.ResponseWriter, r *http.Request) {
	//	w.Header().Add("content-type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var newTask task
	err := decoder.Decode(&newTask)
	if err == nil {
		//db insert
		res := db.Create(&newTask)
		if res.Error != nil {
			fmt.Println(res.Error)
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("409 - task id already exist"))
		} else {
			fmt.Println(newTask)
			fmt.Fprintf(w, "new task added with id:%d", newTask.Key)
		}

	} else {
		fmt.Println(err)
	}

}

func deleteTaskId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	key, _ := strconv.Atoi(vars["id"])
	var t task
	res := db.Delete(&t, key)
	if res.RowsAffected == 0 {
		fmt.Println(res.Error)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - record doesnt exist"))

	} else {
		fmt.Println(t)
		fmt.Fprintf(w, "task deleted with id :%s", id)
	}

}

func updateTaskId(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var updateTask, t task
	err := decoder.Decode(&updateTask)
	if err == nil {
		res:=db.First(&t, updateTask.Key)
		if res.RowsAffected==0{
			fmt.Fprintf(w,"task doesnt exist")

		}else{
			t.Key = updateTask.Key
			t.Note = updateTask.Note
			t.Title = updateTask.Title
			db.Save(&t)
			fmt.Fprintf(w,"task updated")
		}

	}

}

func handleRequests() {
	//creating new router
	router := mux.NewRouter().StrictSlash(true)
	//middleware to change response type
	router.Use(changeResponseContentType)
	//handling functions
	router.HandleFunc("/", homePage)
	router.HandleFunc("/task", returnTaskAll).Methods("GET")
	router.HandleFunc("/task/{id}", returnTaskId).Methods("GET")
	router.HandleFunc("/task", updateTaskId).Methods("PUT")
	router.HandleFunc("/task", createTask).Methods("POST")
	router.HandleFunc("/task/{id}", deleteTaskId).Methods("DELETE")

	//http service
	log.Fatal(http.ListenAndServe(":10000", router))
}

var db *gorm.DB
var err error

func main() {
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Info,    // Log level
			IgnoreRecordNotFoundError: false,            // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,            // Disable color
		},
	)
	connstr := "user=suyog password=pass dbname=taskdb host=localhost sslmode=disable port=5432"
	db, err = gorm.Open(postgres.Open(connstr), &gorm.Config{Logger: dbLogger})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&task{})

	handleRequests()

	fmt.Println("closing program")
}
