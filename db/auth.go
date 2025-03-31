package db

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HandleUser(w http.ResponseWriter, r *http.Request) {
	var res []byte
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusAccepted)
		w.Header().Add("Access-Control-Allow-Methods", http.MethodPatch)
	}
	username, password, ok := r.BasicAuth()
	if ok {
		db := DB{}
		db.OpenConnection()
		defer db.CloseConnection()
		args := strings.Split(r.URL.String()[len(`/user/`):], `/`)
		var passHash string

		rows, _ := db.DB.Query(
			"SELECT password FROM users WHERE username=$1",
			username,
		)
		defer rows.Close()
		for rows.Next() {
			rows.Scan(&passHash)
		}

		if VerifyPassword(password, passHash) {
			switch r.Method {
			case http.MethodGet:
				switch args[0] {
				case "posts":
					res, _ = json.Marshal(db.GetPostsByCol("Author", username))
					w.WriteHeader(http.StatusOK)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			case http.MethodPost:
				switch args[0] {
				case "":
					// create new user
					var u = struct{ Username, Password string }{}
					json.NewDecoder(r.Body).Decode(&u)
					passHash, _ := HashPassword(u.Password)
					_, err := db.DB.Exec(
						"INSERT INTO users (username, password) VALUES ($1,$2)",
						u.Username,
						passHash,
					)
					if err != nil {
						w.WriteHeader(http.StatusConflict)
					}
					w.WriteHeader(http.StatusCreated)
				case "post":
					// insert post
					var p = Post{}
					json.NewDecoder(r.Body).Decode(&p)
					p.PublishedAt = time.Now()
					p.Author = username
					err := p.Write(db.DB)
					if err != nil {
						log.Println(err.Error())
						w.WriteHeader(http.StatusUnprocessableEntity)
					}
					w.WriteHeader(http.StatusCreated)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			case http.MethodDelete:
				switch args[0] {
				case "post":
					if len(args) == 1 || args[1] == "" {
						w.WriteHeader(http.StatusNotFound)
					} else {
						res, err := db.DB.Exec("DELETE FROM posts WHERE Slug = $1", args[1])
						n, _ := res.RowsAffected()
						if err != nil || n == 0 {
							w.WriteHeader(http.StatusNotFound)
						} else {
							w.WriteHeader(http.StatusOK)
						}
					}
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			case http.MethodPatch:
				switch args[0] {
				case "post":
					if len(args) == 1 || args[1] == "" {
						w.WriteHeader(http.StatusNotFound)
					} else {
						var p = Post{}
						json.NewDecoder(r.Body).Decode(&p)
						res, err := db.DB.Exec(
							"UPDATE posts SET Title = $1, Category = $2, Tags = $3, Content = $4 WHERE Slug = $5",
							p.Title,
							p.Category,
							p.Tags,
							p.Content,
							args[1],
						)
						n, _ := res.RowsAffected()
						if err != nil || n == 0 {
							w.WriteHeader(http.StatusNotFound)
						} else {
							w.WriteHeader(http.StatusOK)
						}
					}
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			default:
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
	w.Write(res)
}

// HashPassword generates a bcrypt hash for the given password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// VerifyPassword verifies if the given password matches the stored hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
