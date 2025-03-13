package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"kiskislaya/ConcurrencyCalc/internal/models"
)

func Worker() {
	for {
		task := fetchTask()
		if task == nil {
			// log.Printf("info: no task available, sleeping for 500ms")
			time.Sleep(500 * time.Millisecond)
			continue
		}

		result := executeTask(task)
		sendResult(task.ID, result)
	}
}

func fetchTask() *models.Task {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		log.Printf("error: failed to fetch task: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// log.Printf("warn: got bad status code trying to fetch task: %d", resp.StatusCode)
		return nil
	}

	var data struct {
		Task models.Task `json:"task"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	log.Printf("info: got task with id %d", data.Task.ID)

	return &data.Task
}

func executeTask(task *models.Task) float64 {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	var result float64

	switch task.Operation {
	case "+":
		result = task.Arg1 + task.Arg2
	case "-":
		result = task.Arg1 - task.Arg2
	case "*":
		result = task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 == 0 {
			log.Println("Error: division by zero")
			result = 0
		}
		result = task.Arg1 / task.Arg2
	}

	log.Printf("info: executing task with id %d: %f %s %f = %f", task.ID, task.Arg1, task.Operation, task.Arg2, result)

	return result
}

func sendResult(id int, value float64) {
	data := models.Result{ID: id, Value: value}
	jsonData, _ := json.Marshal(data)

	_, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("error: failed to send result: %v", err)
		return
	}
}
