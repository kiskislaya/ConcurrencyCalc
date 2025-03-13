# **Concurrency Calculator (Распределённый вычислитель арифметических выражений)**  

## **Описание**  
Этот проект реализует распределённую систему вычислений арифметических выражений, где:  
- **Оркестратор** принимает выражения от пользователей, разбивает их на отдельные операции и управляет их выполнением.  
- **Агенты** запрашивают у оркестратора задачи, выполняют операции и отправляют результаты обратно.  

Система поддерживает параллельное выполнение частей выражения и масштабируется добавлением новых агентов.  

---

## **Функционал**  

### Оркестратор:
- Принимает арифметические выражения через API.  
- Разбивает выражение на последовательные вычислительные задачи.  
- Назначает задачи агентам и собирает результаты.  
- Обеспечивает порядок выполнения операций.  

### Агент:
- Запрашивает задачи у оркестратора.  
- Выполняет арифметические операции (сложение, вычитание, умножение, деление).  
- Отправляет результаты обратно оркестратору.  
- Работает в многопоточной среде (количество горутин регулируется).  

---

## **Файловая структура**  

```
/ConcurrencyCalc
│── /cmd
│   ├── /orchestrator
│   │   ├── main.go
│   ├── /agent
│   │   ├── main.go
│── /internal
│   ├── /models
│   │   ├── expression.go
│   │   ├── task.go 
│   │   ├── result.go
│   ├── /orchestrator
│   │   ├── handler.go
│   ├── /agent
│   │   ├── worker.go
│── .gitignore
│── go.mod
│── README.md
```

---

## **Установка и запуск**  

### **1. Клонирование репозитория**  
```sh
git clone https://github.com/kiskislaya/ConcurrencyCalc.git
cd ConcurrencyCalc
```

### **2. Запуск оркестратора**  
```sh
go run cmd/orchestrator/main.go
```
Сервер запустится на `http://localhost:8080/`

### **3. Запуск агента (с указанием мощности вычислений)**  
```sh
COMPUTING_POWER=3 go run cmd/agent/main.go
```
`COMPUTING_POWER` определяет количество параллельных горутин для вычислений.

---

## **API Оркестратора**  

### **1. Добавление выражения на вычисление**
```sh
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2 + 2 * 2"
}'
```
**Ответ:**
```json
{
    "id": "unique-expression-id"
}
```

### **2. Получение списка выражений**
```sh
curl --location 'localhost:8080/api/v1/expressions'
```
**Ответ:**
```json
{
    "expressions": [
        {
            "id": "unique-expression-id",
            "status": "pending",
            "result": null
        }
    ]
}
```

### **3. Получение конкретного выражения**
```sh
curl --location 'localhost:8080/api/v1/expressions/unique-expression-id'
```
**Ответ:**
```json
{
    "expression": {
        "id": "unique-expression-id",
        "status": "done",
        "result": 6
    }
}
```

---

## **Как работает агент?**  
Агент взаимодействует с оркестратором следующим образом:  

1. Запрашивает задачу на вычисление:  
```sh
curl --location 'localhost:8080/internal/task'
```
2. Выполняет операцию (например, `2 * 2 = 4`).
3. Отправляет результат обратно в оркестратор:
```sh
curl --location 'localhost:8080/internal/task' \
--header 'Content-Type: application/json' \
--data '{
  "id": "task-id",
  "result": 4
}'
```

---

## **Настройки окружения**
Эти переменные можно задать перед запуском:
```sh
export TIME_ADDITION_MS=500
export TIME_SUBTRACTION_MS=400
export TIME_MULTIPLICATIONS_MS=700
export TIME_DIVISIONS_MS=800
export COMPUTING_POWER=3
```

---

## **Масштабирование**
Чтобы увеличить вычислительную мощность, можно запустить несколько агентов на разных машинах:
```sh
COMPUTING_POWER=5 go run cmd/agent/main.go
```