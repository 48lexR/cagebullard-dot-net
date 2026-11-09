package main

import (
	"fmt"
	"encoding/json"
	"github.com/mattn/go-sqlite3"
	"database/sql"
	"net/http"
	"errors"
	"log"
	"os"
	"strings"
	"github.com/gorilla/mux"
)

type Post struct {
	Title string `json:"title"`
	Author string `json:"author"`
	Body string `json:"body"`
}

type PostRecord struct {
	PostId int `json:"postid"`
	Post
}

// type User struct {
// UserId int `json:"userid"`
// UserName string `json:"user"`
// }

var db *sql.DB

func main() {
	db, err = sql.Open("sqlite", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS posts(
postid SERIAL PRIMARY KEY,
title VARCHAR(255)
author VARCHAR(255)
body TEXT
)`)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter
	baseURL := "/api"

	/** posts ENDPOINT */
	router.HandleFunc(baseURL+"/posts", getPosts).Methods("GET")
	router.HandleFunc(baseURL+"/posts/{id}", getPost).Methods("GET")
	router.HandleFunc(baseURL+"/posts", createPost).Methods("POST")
	router.HandleFunc(baseURL+"/posts/{id}", updatePost).Methods("PUT")
	router.HandleFunc(baseURL+"/posts/{id}", deletePost).Methods("DELETE")
}

/** HELPER FUNCTIONS */
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("database error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func isJSON(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Content-Type"), "application/json")
}

func decodePost(w http.ResponseWriter, r *http.Request) (Property, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var p Post
	if err := dec.Decode(&p); err != nil {
		return p, fmt.Errorf("invalid json %w", err)
	}

	if strings.TrimSpace(p.Body) == "" {
		return p, errors.New("address required")
	}
	if strings.TrimSpace(p.Author) == "" {
		return p, errors.New("author required")
	}
	return p, nil
}

/** posts ENDPOINT */
func getPosts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	query_string := "SELECT * FROM posts"
	postid := query.Get("postid")
	author := query.Get("author")
	title := query.Get("title")
	contains := query.Get("contains")
	var conditions []string
	var slice []any
	if postid != "" {
		conditions = append(conditions, "name = ?")
		slice = append(slice, postid)
	}
	if author != "" {
		conditions = append(conditions, "author = ?")
		slice = append(slice, author)
	}
	if title != "" {
		conditions = append(conditions, "title = ?")
		slice = append(slice, title)
	}
	if contains != "" {
		conditions = append(conditions, "body LIKE ?")
		slice = append(slice, "%"+contains+"%")
	}

	if len(conditions) > 0 {
		query_string += " WHERE " + strings.Join(conditions, "AND")
	}
	query_string += "ORDER BY postid DESC"
	rows, err = db.Exec(query_string, params...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	posts := []PostRecord{}
	for rows.Next() {
		var p PostRecord
		if err := rows.Scan(&p.ID, &p.Address, &p.Rent, &p.Description); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posts = append(properties, p)
	}

	writeJSON(w, http.StatusOK, posts)
}

func getPost(w http.ResponseWriter, r *http.Request){
	postid := mux.Vars(r)["postid"]

	var p PostRecord
	err := db.QueryRow("SELECT postid, title, author, COALESCE(body, '') FROM posts WHERE postid = ?",
		postid).Scan(&p.PostId, &p.Title, &p.Author, &p.Body)

	if errors.Is(err, sql.NoRows) {
		http.Error(w, "post not found", http.StatusNotFound)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func createPost(w http.ResponseWriter, r *http.Request){
	if !isJSON(r) {
		http.Error(w, "Content-Type must be application/json",
			http.StatusUnsupportedMediaType)
		return
	}

	p, err := decodePost(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := db.Exec(
		"INSERT INTO posts (title, author, body) VALUES (?, ?, ?)",
		p.Title, p.Author, p.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	postid, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, PostRecord{PostId: int(postid), Post: p})
}

func updatePost(w http.ResponseWriter, r *http.Request){
	if !isJSON(r) {
		http.Error(w, "Content-Type must be application/json")
		return
	}

	id := mux.Vars(r)["postid"]

	p, err := decodePost(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := db.Exec(
		`UPDATE properties SET author=? title=? body=? WHERE postid=?`,
		p.Author, p.Title, p.Body, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	n, err := res.RowsAffected()
	if n == 0 {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deletePost(w http.ResponseWriter, r *http.Request){
	id := mux.Vars(r)["postid"]
	res, err := db.Exec("DELETE FROM posts WHERE postid = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	n, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n == 0 {
		http.Error(w, "post nto found", http.StatusNotFound)
	}
	w.WriteHEader(http.StatusNoContent)
}
