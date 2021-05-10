package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
)

type Article struct {
	Id                    uint16
	Title, Anons, Article string
}

var posts []Article
var viewer Article

func rootPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		_, err := fmt.Fprintf(w, err.Error())
		if err != nil {
			return
		}
	}

	db, err := sql.Open("mysql", "gotfry:gotfry@tcp(192.168.42.10:3306)/gotfry")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM `articles`")
	if err != nil {
		panic(err)
	}

	posts = []Article{}
	for res.Next() {
		var post Article

		err = res.Scan(&post.Id, &post.Title, &post.Anons, &post.Article)
		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
	}

	defer res.Close()

	err = t.ExecuteTemplate(w, "index", posts)
	if err != nil {
		return
	}

}

func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	err = t.ExecuteTemplate(w, "create", nil)
	if err != nil {
		return
	}
}

func saveArticle(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	anons := r.FormValue("anons")
	article := r.FormValue("article")

	if title == "" || anons == "" || article == "" {
		_, err := fmt.Fprintf(w, "Please fill all fields")
		if err != nil {
			return
		}
	} else {
		db, err := sql.Open("mysql", "gotfry:gotfry@tcp(192.168.42.10:3306)/gotfry")
		if err != nil {
			panic(err)
		}

		defer db.Close()

		insert, err := db.Query(fmt.Sprintf("INSERT INTO `articles` (`title`, `anons`, `article`) VALUES('%s', '%s', '%s')", title, anons, article))
		if err != nil {
			panic(err)
		}
		defer insert.Close()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func articlePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	t, err := template.ParseFiles("templates/viewer.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		_, err := fmt.Fprintf(w, err.Error())
		if err != nil {
			return
		}
	}

	//Connect to DB
	db, err := sql.Open("mysql", "gotfry:gotfry@tcp(192.168.42.10:3306)/gotfry")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	//Get data from DB
	res, err := db.Query(fmt.Sprintf("SELECT * FROM `articles` WHERE id = %s", vars["id"]))
	if err != nil {
		panic(err)
	}

	viewer = Article{}
	for res.Next() {
		var post Article

		err = res.Scan(&post.Id, &post.Title, &post.Anons, &post.Article)
		if err != nil {
			panic(err)
		}

		viewer = post
	}

	defer res.Close()

	err = t.ExecuteTemplate(w, "viewer", viewer)
	if err != nil {
		return
	}
}

func Handler() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	rtr := mux.NewRouter()
	rtr.HandleFunc("/", rootPage).Methods("GET")
	rtr.HandleFunc("/create/", create).Methods("GET")
	rtr.HandleFunc("/save_article", saveArticle).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}", articlePage).Methods("GET")

	http.Handle("/", rtr)

	err := http.ListenAndServe(":8070", nil)
	if err != nil {
		return
	}
}

func main() {
	Handler()
}
