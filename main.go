package main

import (
    "database/sql"
    "errors"
    "fmt"
    "log"
	"net/http"
    "os"
    "strconv"
    "time"

	"github.com/gin-gonic/gin"
    // アンダーバーはブランクインポート
    // 実際の DB 処理には database/sql を使うため、パッケージの初期化処理だけ行う
    _ "github.com/go-sql-driver/mysql"
)

// int 型を基底型として Status 型を定義
type Status int

// Status の定数を定義
const (
    StatusPending   Status = 1
    StatusCompleted Status = 10
)

// DB 接続を保持するグローバル変数
var db *sql.DB

// todo の構造体を定義
type todo struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Status      Status    `json:"status"`
    DueDate     time.Time `json:"due_date"`
}

// gin.Context はその HTTP リクエストに関する情報を表す構造体
// HTTP リクエスト処理時に自動的に渡される
// Laravel の Request みたいなやつ
func getTodos(c *gin.Context) {
    todos, err := fetchAllTodos()
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.IndentedJSON(http.StatusOK, todos)
}

/*
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
            if !updateData.DueDate.IsZero() {
                todos[i].DueDate = updateData.DueDate
            }

            c.Status(http.StatusNoContent)
            return
        }
    }

    c.IndentedJSON(http.StatusNotFound, gin.H{"error": "TODO not found."})
}

func deleteTodo(c *gin.Context) {
    id, err := convertStrToInt(c.Param("id"))
    if err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }

    for i, b := range todos {
        if b.ID == id {
            todos = append(todos[:i], todos[i+1:]...)
            c.Status(http.StatusNoContent)
            return
        }
    }

    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "TODO not found."})
}
*/

func convertStrToInt(str string) (int, error) {
    num, err := strconv.Atoi(str)

    if err != nil {
        return 0, errors.New("Invalid syntax")
    }

    return num, nil
}

func fetchAllTodos() ([]todo, error) {
    query := "SELECT id, name, description, status, due_date FROM todos"
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var todos []todo
    for rows.Next() {
        var t todo
        if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Status, &t.DueDate); err != nil {
            return nil, err
        }
        todos = append(todos, t)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return todos, nil
}

func initDB() {
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    db_name := os.Getenv("DB_NAME")

    // DB に接続
    var err error
    // time.Time 型を適切に扱うために parseTime と loc を指定
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FTokyo", user, password, host, port, db_name)
    db, err = sql.Open("mysql", dsn)  // グローバル変数に入れたいため、:= は使わない
    if err != nil {
        log.Fatal("DB connection error: ", err)
    }

    // データベース接続の確認
    if err := db.Ping(); err != nil {
        log.Fatal("DB healthcheck error: ", err)
    }
}

func main() {
    initDB()
    defer db.Close()

    r := gin.Default()  // Engine インスタンスを生成し、そのポインターを返す

    v1 := r.Group("/v1")
    {
        v1.GET("/todos", getTodos)
        /*
        v1.GET("/todos/:id", getTodoById)
        v1.POST("/todos", createTodo)
        v1.PATCH("/todos/:id", updateTodo)
        v1.DELETE("/todos/:id", deleteTodo)
        */
    }

    r.Run()
}
