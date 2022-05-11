package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/fuzzbuzz/go-fuzzing-tutorial/03-rest-api/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	dbPath := filepath.Join(os.TempDir(), "db.db")
	defer os.RemoveAll(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// In a production app, you'd have a more robust
	// method of migrating your db
	err = handlers.MigrateDB(db)
	if err != nil {
		panic(err)
	}

	handlers := handlers.NewHandlers(db)

	r := gin.Default()
	r.POST("/users", handlers.AddUser)
	r.GET("/users/:user_email", handlers.GetUser)
	r.PATCH("/users/:user_email", handlers.UpdateUser)
	r.DELETE("/users/:user_email", handlers.DeleteUser)
	r.Run() // listen and serve on 0.0.0.0:8080
}
