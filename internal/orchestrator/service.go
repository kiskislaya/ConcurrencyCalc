package orchestrator

import (
	"kiskislaya/ConcurrencyCalc/internal/models"
	"os"
	"strconv"
	"strings"
	"sync"
)

var expressions = make(map[int]*models.Expression)
var tasks = make(chan models.Task, 100)
var mu sync.Mutex

func processExpression(expID int, expression string) {
	tokens := strings.Split(expression, " ")
	if len(tokens) != 3 {
		mu.Lock()
		expressions[expID].Status = "error"
		mu.Unlock()
		return
	}

	arg1, _ := strconv.ParseFloat(tokens[0], 64)
	arg2, _ := strconv.ParseFloat(tokens[2], 64)
	op := tokens[1]
	opTime := getOperationTime(op)
	task := models.Task{
		ID:            expID,
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     op,
		OperationTime: opTime,
	}

	mu.Lock()
	tasks <- task
	mu.Unlock()
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
