package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

func HandlePosts(w http.ResponseWriter, r *http.Request) {
	db := DB{}
	db.OpenConnection()
	defer db.CloseConnection()
	w.Header().Add("Content-Type", "application/json")
	var res []byte
	args := strings.Split(r.URL.String()[len(`/blog/`):], `/`)
	var err error
	if r.Method == http.MethodGet {
		switch true {
		case args[0] == "":
			res, err = json.Marshal(db.GetDistinctCol("Category"))
		case slices.Contains(db.GetDistinctCol("Category"), args[0]):
			res, err = json.Marshal(db.GetPostsByCol("Category", args[0]))
		case slices.Contains(db.GetDistinctCol("Author"), args[0]):
			res, err = json.Marshal(db.GetPostsByCol("Author", args[0]))
		case slices.Contains(db.GetTags(), args[0]):
			res, err = json.Marshal(db.GetPostsByCol("Tags", args[0]))
		case args[0] == "check_slug":
			u, s := db.CheckSlug(args[1])
			res, err = json.Marshal(
				struct {
					Unique bool
					Slug   string
				}{
					Unique: u,
					Slug:   s,
				},
			)
			w.WriteHeader(http.StatusOK)
		case args[0] == "check_category":
			nA, c := db.CheckCategoryIsNotAuthor(args[1])
			res, err = json.Marshal(
				struct {
					NotAuthor bool
					Category  string
				}{
					NotAuthor: nA,
					Category:  c,
				},
			)
			w.WriteHeader(http.StatusOK)
		default:
			err = errors.New("route not found")
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(res)
}

func (p *Post) Write(db *sql.DB) error {
	_, err := db.Exec(
		"INSERT INTO posts (Title, Content, Category, Author, Tags, Slug, PublishedAt) VALUES ('" +
			p.Title + "', '" +
			p.Content + "', '" +
			p.Category + "', '" +
			p.Author + "', '" +
			p.Tags + "', '" +
			p.Slug + "', '" +
			p.PublishedAt.Format(time.RFC3339) + "')",
	)
	return err
}

type Post struct {
	Title       string
	Content     string
	Category    string
	Author      string
	Tags        string
	Slug        string
	PublishedAt time.Time
}

func (db *DB) GetDistinctCol(col string) []string {
	var i string
	var r []string
	rows, _ := db.DB.Query("SELECT DISTINCT " + col + " FROM Posts")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&i)
		r = append(r, i)
	}
	return r
}

func (db *DB) GetPostsByCol(col, term string) []Post {
	var r []Post
	var q string
	if col == "Tags" {
		q = "SELECT * FROM Posts WHERE " + col + " LIKE '%" + term + "%'"
	} else {
		q = "SELECT * FROM Posts WHERE " + col + " = '" + term + "'"
	}
	rows, _ := db.DB.Query(q)
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
		var p Post
		j, _ := json.Marshal(row)
		json.Unmarshal(j, &p)
		r = append(r, p)
	}
	return r
}

func (db *DB) GetTags() []string {
	var r []string
	for _, i := range strings.Split(
		strings.Join(db.GetDistinctCol("Tags"), ","), ",") {
		i = strings.TrimSpace(i)
		if !slices.Contains(r, i) {
			r = append(r, i)
		}
	}
	return r
}

func (db *DB) CheckSlug(slug string) (bool, string) {
	unique := true
	cols := db.GetDistinctCol("Slug")
	if slices.Contains(cols, slug) {
		unique = false
		i := 1
		for slices.Contains(cols, slug) {
			slice := strings.Split(slug, "-")
			if _, err := strconv.Atoi(slice[len(slice)-1]); err == nil {
				slug = fmt.Sprintf("%s-%d", strings.Join(slice[:len(slice)-1], "-"), i)
			} else {
				slug = slug + "-1"
			}
			i++
		}
	}
	return unique, slug
}

func (db *DB) CheckCategoryIsNotAuthor(category string) (bool, string) {
	notAuthor := true
	cols := db.GetDistinctCol("Author")
	if slices.Contains(cols, category) {
		notAuthor = false
		category = category + "-topic"
	}
	return notAuthor, category
}
