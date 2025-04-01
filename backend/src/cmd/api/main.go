package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Post 구조체 정의
type Post struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"` // SQLite의 DATETIME 형식을 문자열로 받음
}

var db *sql.DB

func main() {
	// 데이터베이스 연결
	var err error
	db, err = sql.Open("sqlite3", "./db/blog.db") // docker-compose.yml의 볼륨 마운트 경로와 일치해야 함
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 데이터베이스 파일이 없으면 생성 및 테이블 생성
	if _, err := os.Stat("./db/blog.db"); os.IsNotExist(err) {
		fmt.Println("Creating database file and table...")
		createDatabaseAndTable()
	}

	// Gin 라우터 생성
	router := gin.Default()

	// CORS 설정 (개발 환경에서 필요할 수 있음)
	router.Use(CORSMiddleware())

	// API 엔드포인트 등록
	router.GET("/api/posts", getPosts)
	router.GET("/api/posts/:id", getPost)
	router.POST("/api/posts", createPost)
	router.PUT("/api/posts/:id", updatePost)
	router.DELETE("/api/posts/:id", deletePost)

	// 서버 시작
	fmt.Println("Server listening on port 8080...")
	router.Run(":8080")
}

// 데이터베이스 생성 및 테이블 생성 함수
func createDatabaseAndTable() {
	file, err := os.Create("./db/blog.db")
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	createTableSQL := `
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

// CORS 미들웨어 (개발 환경에서 필요할 수 있음)
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// 글 목록 조회 API
func getPosts(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, content, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		posts = append(posts, post)
	}

	c.JSON(http.StatusOK, posts)
}

// 특정 글 조회 API
func getPost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	row := db.QueryRow("SELECT id, title, content, created_at FROM posts WHERE id = ?", id)
	var post Post
	err = row.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, post)
}

// 글 생성 API
func createPost(c *gin.Context) {
	var post Post
	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("INSERT INTO posts (title, content) VALUES (?, ?)", post.Title, post.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	post.ID = int(id)
	c.JSON(http.StatusCreated, post)
}

// 글 수정 API
func updatePost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post Post
	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = db.Exec("UPDATE posts SET title = ?, content = ? WHERE id = ?", post.Title, post.Content, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	post.ID = id
	c.JSON(http.StatusOK, post)
}

// 글 삭제 API
func deletePost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	_, err = db.Exec("DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
