CREATE DATABASE IF NOT EXISTS go_todo_app;

USE go_todo_app;

CREATE TABLE IF NOT EXISTS todos (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status INT NOT NULL,
    due_date DATETIME
);

INSERT INTO todos (id, name, description, status, due_date)
VALUES
(1, '買い物', '卵、牛乳', 10, DATE_ADD(NOW(), INTERVAL 24 HOUR)),
(2, '読書', 'Go 入門', 1, DATE_ADD(NOW(), INTERVAL 48 HOUR)),
(3, 'Gin のチュートリアル読む', 'https://go.dev/doc/tutorial/web-service-gin', 1, DATE_ADD(NOW(), INTERVAL 72 HOUR));
