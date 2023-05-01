package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Book struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
}


func main() {
	db, err := sql.Open("postgres", "host=localhost port=6543 user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS books (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			author TEXT NOT NULL,
			description TEXT
		)
	`)

	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.GET("/books", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, title, author, description FROM books")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var books []Book

		for rows.Next() {
			var b Book
			err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Description)
			if err != nil {
				log.Fatal(err)
			}
			books = append(books, b)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, books)
	})


	router.GET("/books/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Fatal(err)
		}

		var b Book
		err = db.QueryRow("SELECT id, title, author, description FROM books WHERE id = $1", id).Scan(&b.ID, &b.Title, &b.Author, &b.Description)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, b)
	})

	router.POST("/books", func(c *gin.Context) {
		var b Book
		if err := c.BindJSON(&b); err != nil {
			log.Fatal(err)
		}

		err := db.QueryRow("INSERT INTO books (title, author, description) VALUES ($1, $2, $3) RETURNING id", b.Title, b.Author, b.Description).Scan(&b.ID)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, b)
	})

	router.PUT("/books/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Fatal(err)
		}

		var b Book
		if err := c.BindJSON(&b); err != nil {
			log.Fatal(err)
		}
	
		_, err = db.Exec("UPDATE books SET title = $1, author = $2, description = $3 WHERE id = $4", b.Title, b.Author, b.Description, id)
		if err != nil {
			log.Fatal(err)
		}
	
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Book %d has been updated", id)})
	})
	
	router.DELETE("/books/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Fatal(err)
		}
	
		result, err := db.Exec("DELETE FROM books WHERE id = $1", id)
		if err != nil {
			log.Fatal(err)
		}
	
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}
	
		if rowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Book %d not found", id)})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Book %d has been deleted", id)})
		}
	})	
	
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}