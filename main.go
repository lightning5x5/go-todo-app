package main

import (
	"net/http"
    "strconv"
    "time"

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
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    // id が整数ではない場合、エラーレスポンスを返す
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
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
    r.GET("/todos/:id", getTodoById)
    r.POST("/todos", addTodo)

    r.Run()
}
