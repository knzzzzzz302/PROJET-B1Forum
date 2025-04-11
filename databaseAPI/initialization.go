package databaseAPI

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// CreateUsersTable creates the users table
func CreateUsersTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, email TEXT, password TEXT, cookie TEXT, expires TEXT)")
	statement.Exec()
}

// CreatePostTable create post table
func CreatePostTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, title TEXT, categories TEXT, content TEXT, created_at TEXT, upvotes INTEGER, downvotes INTEGER)")
	statement.Exec()
}

// CreateCommentTable creates a comment table
func CreateCommentTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, post_id INTEGER, content TEXT, created_at TEXT)")
	statement.Exec()
}

// CreateVoteTable create the vote table into given database
func CreateVoteTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS votes (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, post_id INTEGER, vote INTEGER)")
	statement.Exec()
}

// CreateCategoriesTable create the categories' table into given database
func CreateCategoriesTable(database *sql.DB) {
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, name TEXT, icon TEXT)")
	statement.Exec()
}


// CreateCommentLikesTable crée la table des likes de commentaires
func CreateCommentLikesTable(database *sql.DB) {
    statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS comment_likes (id INTEGER PRIMARY KEY AUTOINCREMENT, comment_id INTEGER NOT NULL, user_id INTEGER NOT NULL, created_at TEXT, FOREIGN KEY (comment_id) REFERENCES comments(id), UNIQUE(comment_id, user_id))")
    statement.Exec()
}


// CreateCommentDislikesTable crée la table des dislikes de commentaires
func CreateCommentDislikesTable(database *sql.DB) {
    statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS comment_dislikes (id INTEGER PRIMARY KEY AUTOINCREMENT, comment_id INTEGER NOT NULL, user_id INTEGER NOT NULL, created_at TEXT, FOREIGN KEY (comment_id) REFERENCES comments(id), UNIQUE(comment_id, user_id))")
    statement.Exec()
}
// CreateCategories creates categories in the database
// CreateCategories creates categories in the database
func CreateCategories(database *sql.DB) {
    statement, _ := database.Prepare("INSERT INTO categories (name) SELECT ? WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = ?)")
    statement.Exec("Général", "Général")
    statement.Exec("Technologie", "Technologie")
    statement.Exec("Science", "Science")
    statement.Exec("Sports", "Sports") 
    statement.Exec("Jeux Vidéo", "Jeux Vidéo")
    statement.Exec("Musique", "Musique")
    statement.Exec("Livres", "Livres")
    statement.Exec("Films", "Films")
    statement.Exec("Télévision", "Télévision")
    statement.Exec("Cuisine", "Cuisine")
    statement.Exec("Voyage", "Voyage")
    statement.Exec("Photographie", "Photographie")
    statement.Exec("Art", "Art")
    statement.Exec("Écriture", "Écriture")
    statement.Exec("Programmation", "Programmation")
    statement.Exec("Autre", "Autre")
}

// createCategoriesIcons creates categories' icons in the database
func CreateCategoriesIcons(database *sql.DB) {
    statement, _ := database.Prepare("UPDATE categories SET icon = ? WHERE name = ?")
    statement.Exec("fa-globe", "Général")
    statement.Exec("fa-laptop", "Technologie")
    statement.Exec("fa-flask", "Science")
    statement.Exec("fa-futbol-o", "Sports")
    statement.Exec("fa-gamepad", "Jeux Vidéo")
    statement.Exec("fa-music", "Musique")
    statement.Exec("fa-book", "Livres")
    statement.Exec("fa-film", "Films")
    statement.Exec("fa-tv", "Télévision")
    statement.Exec("fa-cutlery", "Cuisine")
    statement.Exec("fa-plane", "Voyage")
    statement.Exec("fa-camera", "Photographie")
    statement.Exec("fa-paint-brush", "Art")
    statement.Exec("fa-pencil", "Écriture")
    statement.Exec("fa-code", "Programmation")
    statement.Exec("fa-question", "Autre")
}