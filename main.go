package main

import (
	"fmt"
	"net/http"
	"trash_archive_blog/db"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/blog/", db.HandlePosts)
	mux.HandleFunc("/comments/", db.HandleComments)
	mux.HandleFunc("/user/", db.HandleUser)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodPost, http.MethodGet, http.MethodOptions, http.MethodDelete, http.MethodPatch},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})
	handler := c.Handler(mux)
	fmt.Println("Server listening on http://localhost:8080/")
	http.ListenAndServe(":8080", handler)
}
