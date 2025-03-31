package db

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"
)

func HandleComments(w http.ResponseWriter, r *http.Request) {
	db := DB{}
	db.OpenConnection()
	defer db.CloseConnection()
	w.Header().Add("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		args := strings.Split(r.URL.String()[len(`/comments/`):], `/`)
		var res []byte
		var err error
		switch true {
		case slices.Contains(db.GetDistinctCol("Slug"), args[0]):
			res, err = json.Marshal(db.GetComments(args[0]))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
		if err != nil {
			panic(err)
		}
		w.Write(res)
	} else if r.Method == http.MethodPost {
		var c = Comment{}
		json.NewDecoder(r.Body).Decode(&c)
		c.PublishedAt = time.Now()
		if c.Author == "" {
			c.Author = "anon"
		}
		if c.Content == "" {
			return
		}
		err := c.Write(db.DB)
		if err != nil {
			panic(err)
		}
	}
}

func (c *Comment) Write(db *sql.DB) error {
	q := "INSERT INTO comments (Author, Slug, Content, PublishedAt) " +
		"VALUES ('" +
		c.Author + "', '" +
		c.Slug + "', '" +
		c.Content + "', '" +
		c.PublishedAt.Format("2006-01-02 15:04:05") + "')"
	db.Exec(q)
	return nil
}

type Comment struct {
	Author      string
	Slug        string
	PublishedAt time.Time
	Content     string
}

func (db *DB) GetComments(slug string) []Comment {
	var r []Comment
	rows, _ := db.DB.Query("SELECT * FROM comments WHERE Slug = '" + slug + "'")
	defer rows.Close()
	colNames, _ := rows.Columns()
	cols := make([]interface{}, len(colNames))
	for i, _ := range colNames {
		var ii interface{}
		cols[i] = &ii
	}
	for rows.Next() {
		row := make(map[string]interface{})
		rows.Scan(cols...)
		for i, k := range colNames {
			row[k] = &cols[i]
		}
		var p Comment
		j, _ := json.Marshal(row)
		json.Unmarshal(j, &p)
		r = append(r, p)
	}
	return r
}
