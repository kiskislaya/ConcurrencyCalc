package agent

import (
	"bytes"
	"encoding/json"
	"kiskislaya/ConcurrencyCalc/internal/models"
	"log"
	"net/http"
	"time"
)

func Worker() {
	for {
		task := fetchTask()
		if task == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		result := executeTask(task)
		saveResult(task.ID, result)

	}
}

func fetchTask() *models.Task {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Unexpected status code", resp.Status)
		return nil
	}
	defer resp.Body.Close()

	var data map[string]models.Task
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println(err)
		return nil
	}
	task, ok := data["task"]
	if !ok {
		log.Println("Task not found in response")
		return nil
	}

	return &task
}

func executeTask(task *models.Task) float64 {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 == 0 {
			log.Println("Division by zero")
			return 0
		}
		return task.Arg1 / task.Arg2
	default:
		log.Println("Unknown operation", task.Operation)
		return 0
	}
}

func saveResult(taskID int, result float64) {
	var data = struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}{
		ID:     taskID,
		Result: result,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error submitting result:", err)
		return
	}

	_, err = http.Post("http://localhost:8080/task", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error submitting result:", err)
	}
}
