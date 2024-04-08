package main

import (
    "time"
	"net/http"

	"github.com/gin-gonic/gin"
)

// int 型を基底型として Status 型を定義
type Status int

// Status の定数を定義
const (
    StatusPending   Status = 0
    StatusCompleted Status = 1
)

// todo の構造体を定義
type todo struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Status      Status    `json:"status"`
    Duedate     time.Time `json:"duedate"`
}

// FIXME: 後で DB で管理するように直す
// TODO の実データをスライスで定義
var todos = []todo {
    {ID: "1", Name: "買い物", Description: "卵、牛乳", Status: StatusCompleted, Duedate: time.Now().Add(24 * time.Hour)},
    {ID: "2", Name: "読書", Description: "Go 入門", Status: StatusPending, Duedate: time.Now().Add(48 * time.Hour)},
    {ID: "3", Name: "Gin のチュートリアル読む", Description: "https://go.dev/doc/tutorial/web-service-gin", Status: StatusPending, Duedate: time.Now().Add(72 * time.Hour)},
}

func getTodos(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, todos)
}

func addTodo(c *gin.Context) {
    var newTodo todo  // todo 型の変数を定義

    // リクエストで送られてきたデータが問題なく構造体に代入できるか
    if err := c.BindJSON(&newTodo); err != nil {
        return
    }

    todos = append(todos, newTodo)

    c.IndentedJSON(http.StatusCreated, newTodo)
}

func main() {
    r := gin.Default()  // Engine インスタンスを生成し、そのポインターを返す

    // gin.Context はその HTTP リクエストに関する情報を表す構造体
    // HTTP リクエスト処理時に自動的に渡される
    // Laravel の Request みたいなやつ
    r.GET("/todos", getTodos)
    r.POST("/todos", addTodo)

    r.Run()
}
