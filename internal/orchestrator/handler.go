package orchestrator

import (
	"encoding/json"
	"kiskislaya/ConcurrencyCalc/internal/models"
	"net/http"
)

func RegisterHandlers() {
	http.HandleFunc("/api/v1/calculate", calculateHandler)
	http.HandleFunc("/api/v1/expressions", getExpressionsHandler)
	http.HandleFunc("/internal/task", getTaskHandler)
	http.HandleFunc("/internal/task", postTaskHandler)
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
	}
	expressions[expID] = exp

	go processExpression(expID, req.Expression)

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
	var result struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	exp, ok := expressions[result.ID]
	if !ok {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	exp.Status = "done"
	exp.Result = &result.Result
	w.WriteHeader(http.StatusNoContent)
}
