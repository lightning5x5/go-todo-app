package main

import (
    "database/sql"
    "errors"
    "fmt"
    "log"
	"net/http"
    "os"
    "strconv"
    "strings"
    "time"

	"github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    // アンダーバーはブランクインポート
    // 実際の DB 処理には database/sql を使うため、パッケージの初期化処理だけ行う
    _ "github.com/go-sql-driver/mysql"
    "golang.org/x/crypto/bcrypt"
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

// User の構造体を定義
type User struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Password string `json:"password"`
}

// gin.Context はその HTTP リクエストに関する情報を表す構造体
// HTTP リクエスト処理時に自動的に渡される
// Laravel の Request みたいなやつ
func getTodos(c *gin.Context) {
    todos, err := fetchAll()
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.IndentedJSON(http.StatusOK, todos)
}

func getTodoById(c *gin.Context) {
    id, err := convertStrToInt(c.Param("id"))
    if err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})  // gin.H() で JSON を生成する
        return
    }

    todo, err := fetchByID(id)
    if err != nil {
        if err.Error() == "404" {
            c.IndentedJSON(http.StatusNotFound, gin.H{"error": "TODO not found"})
            return
        }
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.IndentedJSON(http.StatusOK, todo)
}

func createTodo(c *gin.Context) {
    var newTodo todo  // todo 型の変数を定義

    // リクエストで送られてきたデータが問題なく構造体に代入できるか
    if err := c.BindJSON(&newTodo); err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid syntax."})
        return
    }

    if err := insert(newTodo); err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

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

    err = update(id, updateData)
    if err != nil {
        if err == sql.ErrNoRows {
            c.IndentedJSON(http.StatusNotFound, gin.H{"error": "TODO not found."})
        } else {
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    c.Status(http.StatusNoContent)
}

func deleteTodo(c *gin.Context) {
    id, err := convertStrToInt(c.Param("id"))
    if err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
        return
    }

    err = delete(id)
    if err != nil {
        if err.Error() == "404" {
            c.IndentedJSON(http.StatusNotFound, gin.H{"error": "TODO not found."})
        } else {
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    c.Status(http.StatusNoContent)
}

func convertStrToInt(str string) (int, error) {
    num, err := strconv.Atoi(str)

    if err != nil {
        return 0, errors.New("Invalid syntax")
    }

    return num, nil
}

func fetchAll() ([]todo, error) {
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

    return todos, nil  // スライス型自体が参照型なので、明示してポインターを返す必要なし
}

func fetchByID(id int) (*todo, error) {
    query := "SELECT id, name, description, status, due_date FROM todos WHERE id = ?"
    row := db.QueryRow(query, id)

    var t todo
    if err := row.Scan(&t.ID, &t.Name, &t.Description, &t.Status, &t.DueDate); err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("404")
        }
        return nil, err
    }

    return &t, nil
}

func insert(newTodo todo) (error) {
    query := "INSERT INTO todos (name, description, status, due_date) VALUES (?, ?, ?, ?)"
    _, err := db.Exec(query, newTodo.Name, newTodo.Description, newTodo.Status, newTodo.DueDate)
    return err
}

func update(id int, updateData todo) (error) {
    // 更新するフィールドと値を保持するマップ
    fields := map[string]interface{}{}

    if updateData.Name != "" {
        fields["name"] = updateData.Name
    }
    if updateData.Description != "" {
        fields["description"] = updateData.Description
    }
    if updateData.Status != 0 {
        fields["status"] = updateData.Status
    }
    if !updateData.DueDate.IsZero() {
        fields["due_date"] = updateData.DueDate
    }

    // フィールドが何も指定されていない場合は更新をスキップ
    if len(fields) == 0 {
        return nil
    }

    // SQL 文を動的に生成
    query := "UPDATE todos SET "
    args := []interface{}{}
    for field, value := range fields {
        query += fmt.Sprintf("%s = ?, ", field)
        args = append(args, value)
    }
    query = strings.TrimRight(query, ", ")  // 末尾のカンマを削除
    query += " WHERE id = ?"
    args = append(args, id)

    _, err := db.Exec(query, args...)
    return err
}

func delete(id int) (error) {
    query := "DELETE FROM todos WHERE id = ?"
    result, err := db.Exec(query, id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()  // 操作件数を取得
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("404")
    }

    return nil
}

func register(c *gin.Context) {
    // リクエストで送られてきたデータが問題なく構造体に代入できるか
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid syntax."})
		return
	}

    // パスワードをハッシュ化して戻す
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.Password = string(hashedPassword)

    // DB に登録
    query := "INSERT INTO users (username, password) VALUES (?, ?)"
    _, err = db.Exec(query, user.Username, user.Password)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}

func login(c *gin.Context) {
    // リクエストで送られてきたデータが問題なく構造体に代入できるか
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid syntax."})
		return
	}

    query := "SELECT id, username, password FROM users WHERE username = ?"
    row := db.QueryRow(query, user.Username)

    // 検索結果を foundUser に代入
	var foundUser User
    if err := row.Scan(&foundUser.ID, &foundUser.Username, &foundUser.Password); err != nil {
        // 検索結果が0件の場合
        if err == sql.ErrNoRows {
		    c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password."})
            return
        }

        // その他
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
    }

    // パスワード比較
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password)); err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password."})
		return
	}

    // 有効期限が1時間の JWT Claim を生成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "JWT Secret not found."})
		return
	}

    // 署名を追加
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"token": tokenString})
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
        v1.GET("/todos/:id", getTodoById)
        v1.POST("/todos", createTodo)
        v1.PATCH("/todos/:id", updateTodo)
        v1.DELETE("/todos/:id", deleteTodo)

        v1.POST("/login", login)
        v1.POST("/register", register)
    }

    r.Run()
}
