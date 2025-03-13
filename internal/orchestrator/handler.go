package orchestrator

import (
	"encoding/json"
	"kiskislaya/ConcurrencyCalc/internal/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	tasks       = make(chan models.Task, 100)
	results     = make(chan models.Result, 100)
	mu          sync.Mutex
	expressions = make(map[int]*models.Expression)
)

func RegisterHandlers() {
	go resultListener()
	http.HandleFunc("POST /api/v1/calculate", calculateHandler)
	http.HandleFunc("GET /api/v1/expressions", getExpressionsHandler)
	http.HandleFunc("GET /internal/task", getTaskHandler)
	http.HandleFunc("POST /internal/task", postTaskHandler)
	http.HandleFunc("GET /api/v1/expressions/{id}", getExpressionByID)
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	expID := len(expressions) + 1
	exp := &models.Expression{
		ID:     expID,
		Expr:   req.Expression,
		Status: "pending",
		Result: nil,
	}
	expressions[expID] = exp

	go parseAndQueueTasks(expID, req.Expression)

	w.Header().Set("Content-Type", "application/json")
	resp := struct {
		ID int `json:"id"`
	}{
		ID: expID,
	}
	json.NewEncoder(w).Encode(resp)
}

func getExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	exprList := make([]*models.Expression, 0, len(expressions))
	for _, exp := range expressions {
		exprList = append(exprList, exp)
	}
	resp := struct {
		Expressions []*models.Expression `json:"expressions"`
	}{
		Expressions: exprList,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case task := <-tasks:
		resp := struct {
			Task models.Task `json:"task"`
		}{
			Task: task,
		}
		json.NewEncoder(w).Encode(resp)
	default:
		http.Error(w, "No tasks available", http.StatusNotFound)
	}
}

func postTaskHandler(w http.ResponseWriter, r *http.Request) {
	var result models.Result

	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	log.Printf("info: received result for task %d: %f", result.ID, result.Value)

	mu.Lock()
	defer mu.Unlock()
	exp, ok := expressions[result.ID]
	if !ok {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	exp.Status = "done"
	exp.Result = &result.Value
	w.WriteHeader(http.StatusNoContent)
}

func getExpressionByID(w http.ResponseWriter, r *http.Request) {
	exprID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	mu.Lock()
	expr, exists := expressions[int(exprID)]
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(struct {
		Expression *models.Expression `json:"expression"`
	}{
		Expression: expr,
	})
}

func resultListener() {
	for result := range results {
		mu.Lock()
		exp, ok := expressions[result.ID]

		if !ok {
			continue
		}

		exp.Status = "done"
		exp.Result = &result.Value
		mu.Unlock()
	}
}

func parseAndQueueTasks(expID int, expression string) {
	tokens := strings.Fields(expression)
	var numStack []float64
	var opStack []string

	for _, token := range tokens {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			numStack = append(numStack, num)
		} else {
			opStack = append(opStack, token)
		}
	}

	for len(opStack) > 0 {
		if len(numStack) < 2 {
			log.Println("Expression error", expression)
			return
		}

		arg1, arg2 := numStack[0], numStack[1]
		op := opStack[0]
		opTime := getOperationTime(op)

		task := models.Task{
			ID:            expID,
			Arg1:          arg1,
			Arg2:          arg2,
			Operation:     op,
			OperationTime: opTime,
		}

		tasks <- task
		numStack = numStack[1:]
		opStack = opStack[1:]
	}
}

func getOperationTime(op string) int64 {
	env := ""
	switch op {
	case "+":
		env = "TIME_ADDITION_MS"
	case "-":
		env = "TIME_SUBTRACTION_MS"
	case "*":
		env = "TIME_MULTIPLICATIONS_MS"
	case "/":
		env = "TIME_DIVISIONS_MS"
	default:
		env = "TIME_OPERATION_MS"
	}
	opTime := os.Getenv(env)
	if opTime == "" {
		return -1
	}
	operationTime, _ := strconv.ParseInt(opTime, 10, 64)
	return operationTime
}
