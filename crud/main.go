package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"time"
	"strconv"
)

var db *pgxpool.Pool

type Article struct {
	Id       int       `json:"id"`
	Title    string    `json:"title"`
	Keywords string    `json:"keywords"`
	Content  string    `json:"content"`
	UserId   int       `json:"user_id"`
	Date     time.Time `json:"date"`
}

func main() {
	r := chi.NewRouter()

	/* DB Connection */
	var err error
	db, err = pgxpool.Connect(context.Background(), "postgres://postgres:12345@localhost:5432/crud")
	if err != nil {
		log.Fatal(err)
	}

	// /api/articles
	r.Route("/api", apiRouter)

	err = http.ListenAndServe(":8080", r)
	if err != http.ErrServerClosed {
		log.Fatal(err)
	} else {
		fmt.Printf("server closing\n")
	}
}

func apiRouter(r chi.Router) {
	r.Route("/articles", articlesRouter)
}

func articlesRouter(r chi.Router) {
	// /api/articles
	r.Get("/", listArticles)
	r.Post("/", createArticle)
	r.Put("/{id}", updateArticle)
	r.Delete("/{id}", deleteArticle)
}

func printError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func listArticles(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(context.Background(), "SELECT * FROM api.articles")
	if err != nil {
		printError(w, err)
		return
	}
	defer rows.Close()
	articles := []Article{}
	for rows.Next() {
		err = rows.Err()
		if err != nil {
			printError(w, err)
			return
		}

		values, err := rows.Values()
		if err != nil {
			printError(w, err)
			return
		}
		a := Article{
			Id:       int(values[0].(int32)),
			Title:    values[1].(string),
			Keywords: values[2].(string),
			Content:  values[3].(string),
			UserId:   int(values[4].(int32)),
			Date:     values[5].(time.Time),
		}
		articles = append(articles, a)
	}

	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(articles)
	articles = nil
	if err != nil {
		printError(w, err)
		return
	}
}

func createArticle(w http.ResponseWriter, r *http.Request) {
	var a Article
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		printError(w, err)
		return
	}

	// id, date
	// title, keywords, user_id
	row := db.QueryRow(context.Background(), "INSERT INTO api.articles(title, keywords, content, user_id) VALUES($1,$2,$3,$4) RETURNING id, date", a.Title, a.Keywords, a.Content, a.UserId)
	lastInsertID := 0
	var date time.Time
	err = row.Scan(&lastInsertID, &date)
	if err != nil {
		printError(w, err)
		return
	}

	a.Id = lastInsertID
	a.Date = date
	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(a)
	if err != nil {
		printError(w, err)
		return
	}
}

func updateArticle(w http.ResponseWriter, r *http.Request) {
	id_string := chi.URLParam(r, "id")
	id, err := strconv.Atoi(id_string)
	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id must be greater than zero"))
		return
	}
	var a map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		printError(w, err)
		return
	}
	

	// Check fields
	var title *string
	var keywords *string
	var content *string
	var user_id *int
	
	title_val, ok := a["title"]
	if !ok {
		title = nil
	} else {
		tsr := title_val.(string)
		title = &tsr
	}

	keywords_val, ok := a["keywords"]
	if !ok {
		keywords = nil
	} else {
		tsr := keywords_val.(string)
		keywords = &tsr
	}

	content_val, ok := a["content"]
	if !ok {
		content = nil
	} else {
		tsr := content_val.(string)
		content = &tsr
	}

	user_id_val, ok := a["user_id"]
	if !ok {
		user_id = nil
	} else {
		tsr := int(user_id_val.(float64))
		user_id = &tsr
	}

	/*
	   UPDATE api.articles SET
	   title = COALESCE($1, title),
	   keywords = COALESCE($2, keywords),
	   content = COALESCE($3, content),
	   user_id = COALESCE($4, content)
	   WHERE id = $5
	   RETURNING title, keywords, content, user_id, date
	*/
	row := db.QueryRow(context.Background(), "UPDATE api.articles SET title = COALESCE($1, title), keywords = COALESCE($2, keywords), content = COALESCE($3, content), user_id = COALESCE($4, user_id) WHERE id = $5 RETURNING title, keywords, content, user_id, date", title, keywords, content, user_id, id)
	var ua Article
	ua.Id = id
	err = row.Scan(&ua.Title, &ua.Keywords, &ua.Content, &ua.UserId, &ua.Date)
	if err != nil {
		printError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(ua)
	if err != nil {
		printError(w, err)
		return
	}
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {
	id_string := chi.URLParam(r, "id")
	id, err := strconv.Atoi(id_string)
	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("id must be greater than zero"))
		return
	}

	row := db.QueryRow(context.Background(), "DELETE FROM api.articles WHERE id = $1 RETURNING title, keywords, content, user_id, date", id)
	var da Article
	da.Id = id
	err = row.Scan(&da.Title, &da.Keywords, &da.Content, &da.UserId, &da.Date)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(da)
	if err != nil {
		printError(w, err)
		return
	}
}
