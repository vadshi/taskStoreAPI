package taskstore

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Task struct {
	Id   int       `json:"id"`
	Text string    `json:"text"`
	Tags []string  `json:"tags"`
	Due  time.Time `json:"due"`
}

type TaskStore struct {
	sync.Mutex
	db *sql.DB
}

func New() *TaskStore {
	ts := &TaskStore{}

	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()
	log.Println("OPENED SUCCESS")

	sqlStmt := `
		CREATE TABLE IF NOT EXISTS task (id INTEGER NOT NULL PRIMARY KEY , task TEXT, tags TEXT,  time TEXT);
		`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	ts.db = db

	return ts
}

func (ts *TaskStore) CreateTask(text string, tags []string, due time.Time) int {
	ts.Lock()
	defer ts.Unlock()

	tx, err := ts.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("INSERT INTO task(id, task, tags, time) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(nil, text, strings.Join(tags, " "), due.String())
	if err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	stmt, err = ts.db.Prepare("SELECT last_insert_rowid()")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var id int
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id)
	}

	return id
}

func (ts *TaskStore) GetTask(id int) (Task, error) {
	ts.Lock()
	defer ts.Unlock()

	stmt, err := ts.db.Prepare("SELECT id, task, tags, time from task where id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	rowsCount := 0

	var tsk Task
	for rows.Next() {
		var id int
		var task string
		var tags string
		var tm string
		// используем указатели для доступа к значениям
		err = rows.Scan(&id, &task, &tags, &tm)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, task, tags, tm)

		ptm, error := time.Parse("2006-01-02 15:04:05 Z0700 MST", tm)
		if error != nil {
			fmt.Println(error)
			log.Fatal(err)
		}

		tsk = Task{
			Id:   id,
			Text: task,
			Tags: []string{tags},
			Due:  ptm}

		rowsCount++
	}

	if rowsCount == 0 {
		return Task{}, fmt.Errorf("task with id=%d not found", id)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return tsk, nil
}

func (ts *TaskStore) GetTag(name string) ([]Task, error) {
	ts.Lock()
	defer ts.Unlock()

	stmt, err := ts.db.Prepare("SELECT id, task, tags, time FROM task WHERE tags LIKE ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	name = "%" + name + "%"
	rows, err := stmt.Query(name)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	rowsCount := 0
	var tsk Task
	var allTasks []Task

	for rows.Next() {
		var id int
		var task string
		var tags string
		var tm string
		err = rows.Scan(&id, &task, &tags, &tm)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, task, tags, tm)

		ptm, error := time.Parse("2006-01-02 15:04:05 Z0700 MST", tm)
		if error != nil {
			fmt.Println(error)
			log.Fatal(err)
		}

		tsk = Task{
			Id:   id,
			Text: task,
			Tags: []string{tags},
			Due:  ptm}

		allTasks = append(allTasks, tsk)
		rowsCount++
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	if rowsCount == 0 {
		return []Task{}, fmt.Errorf("tag with name=%v not found", name)
	}

	return allTasks, nil
}

func (ts *TaskStore) GetDue(date string) ([]Task, error) {
	ts.Lock()
	defer ts.Unlock()

	stmt, err := ts.db.Prepare("SELECT id, task, tags, time from task where time like ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	date = "%" + date + "%"
	rows, err := stmt.Query(date)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tsk Task
	var allTasks []Task

	rowsCount := 0
	for rows.Next() {
		var id int
		var task string
		var tags string
		var tm string
		err = rows.Scan(&id, &task, &tags, &tm)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, task, tags, tm)

		ptm, error := time.Parse("2006-01-02 15:04:05 Z0700 MST", tm)
		if error != nil {
			fmt.Println(error)
			log.Fatal(err)
		}

		tsk = Task{
			Id:   id,
			Text: task,
			Tags: []string{tags},
			Due:  ptm}

		allTasks = append(allTasks, tsk)
		rowsCount++
	}

	if rowsCount == 0 {
		return []Task{}, fmt.Errorf("task with due=%v not found", date)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return allTasks, nil
}

func (ts *TaskStore) GetAllTasks() []Task {
	ts.Lock()
	defer ts.Unlock()

	stmt, err := ts.db.Prepare("SELECT id, task, tags, time from task")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tsk Task
	var allTasks []Task

	rowsCount := 0
	for rows.Next() {
		var id int
		var task string
		var tags string
		var tm string
		err = rows.Scan(&id, &task, &tags, &tm)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, task, tags, tm)

		ptm, error := time.Parse("2006-01-02 15:04:05 Z0700 MST", tm)
		if error != nil {
			fmt.Println(error)
			log.Fatal(err)
		}

		tsk = Task{
			Id:   id,
			Text: task,
			Tags: []string{tags},
			Due:  ptm}

		allTasks = append(allTasks, tsk)
		rowsCount++
	}
	if rowsCount == 0 {
		return []Task{}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return allTasks
}

func (ts *TaskStore) DeleteAllTasks() error {
	ts.Lock()
	defer ts.Unlock()

	_, err := ts.db.Exec("DELETE from task")
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (ts *TaskStore) DeleteTask(id int) error {
	ts.Lock()
	defer ts.Unlock()

	stmt, err := ts.db.Prepare("DELETE FROM task WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	return nil

}
