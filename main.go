package main

import (
    "errors"
	"net/http"
    "strconv"
    "time"

	"github.com/gin-gonic/gin"
)

// int 型を基底型として Status 型を定義
type Status int

// Status の定数を定義
const (
    StatusPending   Status = 1
    StatusCompleted Status = 10
)

// todo の構造体を定義
type todo struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Status      Status    `json:"status"`
    Duedate     time.Time `json:"duedate"`
}

// FIXME: 後で DB で管理するように直す
// TODO の実データをスライスで定義
var todos = []todo {
    {ID: 1, Name: "買い物", Description: "卵、牛乳", Status: StatusCompleted, Duedate: time.Now().Add(24 * time.Hour)},
    {ID: 2, Name: "読書", Description: "Go 入門", Status: StatusPending, Duedate: time.Now().Add(48 * time.Hour)},
    {ID: 3, Name: "Gin のチュートリアル読む", Description: "https://go.dev/doc/tutorial/web-service-gin", Status: StatusPending, Duedate: time.Now().Add(72 * time.Hour)},
}

func getTodos(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, todos)
}

func getTodoById(c *gin.Context) {
    id, err := convertStrToInt(c.Param("id"))
    if err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }

    // PHP の foreach みたいなやつ
    // インデックスは不要なので _ で破棄
    for _, b := range todos {
        if b.ID == id {
            c.IndentedJSON(http.StatusOK, b)
        return
        }
    }

    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "TODO not found."})  // gin.H() で JSON を生成する
}

func createTodo(c *gin.Context) {
    var newTodo todo  // todo 型の変数を定義

    // リクエストで送られてきたデータが問題なく構造体に代入できるか
    if err := c.BindJSON(&newTodo); err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid syntax."})
        return
    }

    todos = append(todos, newTodo)

    c.IndentedJSON(http.StatusCreated, newTodo)
}

func updateTodo(c *gin.Context) {
    id, err := convertStrToInt(c.Param("id"))
    if err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }

    // リクエストで送られてきたデータが問題なく構造体に代入できるか
    var updateData todo
    if err := c.BindJSON(&updateData); err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid syntax."})
        return
    }

    for i, b := range todos {
        if b.ID == id {
            if updateData.Name != "" {
                todos[i].Name = updateData.Name
            }
            if updateData.Description != "" {
                todos[i].Description = updateData.Description
            }
            if updateData.Status > 0 {
                todos[i].Status = updateData.Status
            }
            if !updateData.Duedate.IsZero() {
                todos[i].Duedate = updateData.Duedate
            }

            c.Status(http.StatusNoContent)
            return
        }
    }

    c.IndentedJSON(http.StatusNotFound, gin.H{"error": "TODO not found."})
}

func convertStrToInt(str string) (int, error) {
    num, err := strconv.Atoi(str)

    if err != nil {
        return 0, errors.New("Invalid syntax")
    }

    return num, nil
}

func main() {
    r := gin.Default()  // Engine インスタンスを生成し、そのポインターを返す

    // gin.Context はその HTTP リクエストに関する情報を表す構造体
    // HTTP リクエスト処理時に自動的に渡される
    // Laravel の Request みたいなやつ
    r.GET("/todos", getTodos)
    r.GET("/todos/:id", getTodoById)
    r.POST("/todos", createTodo)
    r.PATCH("/todos/:id", updateTodo)

    r.Run()
}
