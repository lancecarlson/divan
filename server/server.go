package server

import (
	"log"
	"database/sql"
	"github.com/gin-gonic/gin"
)


type Server struct {
	Db *sql.DB
	Tables map[string]Table
}

func (s *Server) FindTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		tbl := c.Param("tbl")
		t, ok := s.Tables[tbl];
		if !ok {
			c.JSON(404, gin.H{"error": "database not found"})
			return
		}
		c.Set("tbl", t)
	}
}

func (s *Server) Start(bootstrap bool) error {
	log.Println("Starting Divan...")

	if bootstrap {
		log.Println("Bootstrapping...")
		err := s.Bootstrap()
		if err != nil {
			return err
		}
	}

	log.Println("Loading Config...")
	if err := s.LoadConfig(); err != nil {
		return err
	}
	return nil
}

func (s *Server) LoadConfig() error {
	tables, err := TableList(s.Db)
	if err != nil {
		return err
	}
	s.Tables = tables

	tableNames := []string{}
	for name, _ := range tables {
		tableNames = append(tableNames, name)
	}
	log.Println("Tables:")
	log.Println(tableNames)
	
	return nil
}

func (s *Server) Bootstrap() error {
	t := NewTable("divan")
	t.Db = s.Db
	err := t.Create()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) TablePut() gin.HandlerFunc {
	return func(c *gin.Context) {
		tbl := c.Param("tbl")
		t := NewTable(tbl)
		if err := t.Create(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		// Reload config
		if err := s.LoadConfig(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		c.JSON(200, gin.H{"ok": true})
		return
	}
}

func (s *Server) DocPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		tbl, _ := c.Get("tbl")
		t := tbl.(Table)

		var j map[string]interface{}
		err := c.BindJSON(&j)
		if err != nil {
			c.JSON(400, gin.H{"error": "could not parse json"})
			return
		}
		
		doc := new(Doc)
		doc.Db = s.Db
		err = doc.Post(t, j)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ok": true, "id": doc.Id, "rev": doc.Rev})
	}
}

func (s *Server) DocGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		tbl, _ := c.Get("tbl")
		t := tbl.(Table)
		id := c.Param("id")

		doc := new(Doc)
		doc.Db = s.Db
		err := doc.Get(t, id)
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		data, err := doc.String()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.String(200, data)
	}
}

func (s *Server) DocPut() gin.HandlerFunc {
	return func(c *gin.Context) {
		tbl, _ := c.Get("tbl")
		t := tbl.(Table)
		id := c.Param("id")

		var j map[string]interface{}
		err := c.BindJSON(&j)
		if err != nil {
			c.JSON(400, gin.H{"error": "could not parse json"})
			return
		}

		doc := new(Doc)
		doc.Db = s.Db
		err = doc.Put(t, id, j)
		if err == ErrDocumentUpdateConflict {
			c.JSON(409, gin.H{"error": "conflict", "reason": err.Error()})
			return
		}
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ok": "true", "id": doc.Id, "rev": doc.Rev})
	}
}

func (s *Server) DocDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		tbl, _ := c.Get("tbl")
		t := tbl.(Table)
		id := c.Param("id")
		rev := c.Query("rev")
		doc := new(Doc)
		doc.Db = s.Db
		err := doc.Delete(t, id, rev)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ok": "true"})
	}
}

func (s *Server) DocHead() gin.HandlerFunc {
	return func(c *gin.Context) {
		tbl, _ := c.Get("tbl")
		t := tbl.(Table)
		id := c.Param("id")
		doc := new(Doc)
		doc.Db = s.Db
		err := doc.Head(t, id)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"error": "Not Implemented Yet"})
	}
}