package calculation

import (
	"kiskislaya/ConcurrencyCalc/internal/models"
	"os"
	"strconv"
	"strings"
)

func isDigit(char rune) bool {
	if char >= '0' && char <= '9' {
		return true
	}
	return false
}

func isOp(char rune) bool {
	if char == '+' || char == '-' || char == '*' || char == '/' {
		return true
	}
	return false
}

func parseNumber(s string, i int) (float64, int) {
	num := ""
	for i < len(s) {
		if (s[i] >= '0' && s[i] <= '9') || s[i] == '.' {
			num += string(s[i])
		} else {
			break
		}
		i++

	}
	res, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, -1
	}
	return res, i

}

func priorityOp(op rune) int {
	switch op {
	case '+':
		return 1
	case '-':
		return 1
	case '*':
		return 2
	case '/':
		return 2
	default:
		return -1
	}
}

func calculate(tasks chan models.Task, numStack *[]float64, opStack *[]rune, expID int) error {
	if len(*numStack) < 2 || len(*opStack) == 0 {
		return ErrIncorrectOperator
	}

	var res float64

	arg1 := (*numStack)[len(*numStack)-2]
	arg2 := (*numStack)[len(*numStack)-1]
	op := (*opStack)[len(*opStack)-1]
	opTime := getOperationTime(string(op))

	task := models.Task{
		ID:            expID,
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     string(op),
		OperationTime: opTime,
	}
	tasks <- task

	*numStack = (*numStack)[:len(*numStack)-2]
	*opStack = (*opStack)[:len(*opStack)-1]

	*numStack = append(*numStack, res)
	return nil
}

func Calc(tasks chan models.Task, expression string, expID int) (float64, error) {
	if len(expression) == 0 {
		return 0, ErrEmptyExpression
	}
	expression = strings.ReplaceAll(expression, " ", "")

	numStack := make([]float64, 0)
	opStack := make([]rune, 0)

	for i := 0; i < len(expression); i++ {
		char := rune(expression[i])
		if isDigit(char) {
			num, j := parseNumber(expression, i)
			numStack = append(numStack, num)
			i = j - 1
		} else if isOp(char) {
			for len(opStack) > 0 && priorityOp(opStack[len(opStack)-1]) >= priorityOp(char) {
				if err := calculate(tasks, &numStack, &opStack, expID); err != nil {
					return 0, err
				}
			}
			opStack = append(opStack, char)
		} else if char == '(' {
			opStack = append(opStack, char)
		} else if char == ')' {
			for len(opStack) > 0 && opStack[len(opStack)-1] != '(' {
				if err := calculate(tasks, &numStack, &opStack, expID); err != nil {
					return 0, err
				}
			}
			if len(opStack) == 0 {
				return 0, ErrInvalidExpression
			}
			opStack = opStack[:len(opStack)-1]
		} else {
			return 0, ErrInvalidExpression
		}
	}

	for len(opStack) > 0 {
		if err := calculate(tasks, &numStack, &opStack, expID); err != nil {
			return 0, err
		}
	}

	if len(numStack) > 1 {
		return 0, ErrInvalidExpression
	}

	return numStack[0], nil
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
