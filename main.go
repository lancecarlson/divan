package main

import (
	"os"
	"log"
	"flag"
	"database/sql"
	"net/http"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/lancecarlson/divan/server"
)

func ContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
	}
}

func main() {
	bootstrap := flag.Bool("b", false, "bootstrap environment.")
	flag.Parse()

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL required")
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	s := server.Server{Db: db}
	err = s.Start(*bootstrap)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.Use(ContentTypeMiddleware())
	router.Handle("PUT", "/:tbl", s.TablePut())
	router.Handle("POST", "/:tbl", s.FindTable(), s.DocPost())
	router.Handle("GET", "/:tbl/:id", s.FindTable(), s.DocGet())
	router.Handle("PUT", "/:tbl/:id", s.FindTable(), s.DocPut())
	router.Handle("DELETE", "/:tbl/:id", s.FindTable(), s.DocDelete())
	router.Handle("HEAD", "/:tbl/:id", s.FindTable(), s.DocHead())
	log.Fatal(http.ListenAndServe(":8080", router))
}