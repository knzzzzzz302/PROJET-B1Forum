package main

import (
	"FORUM-GO/databaseAPI"
	"FORUM-GO/webAPI"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
)

type Post struct {
	Id         int
	Username   string
	Title      string
	Categories []string
	Content    string
	CreatedAt  string
	UpVotes    int
	DownVotes  int
	Comments   []Comment
	Images     []string // Nouveau champ pour les images
}

type Comment struct {
	Id        int
	PostId    int
	Username  string
	Content   string
	CreatedAt string
}

// Database
var database *sql.DB

func main() {
	// Création des dossiers pour les uploads d'images
	if err := os.MkdirAll("public/uploads/profiles", 0755); err != nil {
		fmt.Println("Erreur lors de la création du dossier profiles:", err)
	}
	if err := os.MkdirAll("public/uploads/posts", 0755); err != nil {
		fmt.Println("Erreur lors de la création du dossier posts:", err)
	}

	// Vérification de l'existence de la base de données
	var _, err = os.Stat("database.db")

	// Création de la base de données si elle n'existe pas
	if os.IsNotExist(err) {
		var file, err = os.Create("database.db")
		if err != nil {
			fmt.Println("Erreur lors de la création de la base de données:", err)
			return
		}
		defer file.Close()
	}

	// Ouverture de la base de données
	database, _ = sql.Open("sqlite3", "./database.db")
	defer database.Close()

	// Création des tables
	databaseAPI.CreateUsersTable(database)
	databaseAPI.CreatePostTable(database)
	databaseAPI.CreateCommentTable(database)
	databaseAPI.CreateVoteTable(database)
	databaseAPI.CreateCategoriesTable(database)
	databaseAPI.CreatePostImagesTable(database) // Nouvelle table pour les images des posts
	databaseAPI.CreateCategories(database)
	databaseAPI.CreateCategoriesIcons(database)

	// Configuration du serveur web
	webAPI.SetDatabase(database)

	// Configuration des routes statiques
	fs := http.FileServer(http.Dir("public"))
	
	// Création du routeur personnalisé
	router := webAPI.NewCustomRouter()
	fmt.Println("Démarrage du serveur sur le port http://localhost:8000/")

	// Routes de page
	router.HandleFunc("/", webAPI.Index)
	router.HandleFunc("/register", webAPI.Register)
	router.HandleFunc("/login", webAPI.Login)
	router.HandleFunc("/post", webAPI.DisplayPost)
	router.HandleFunc("/filter", webAPI.GetPostsByApi)
	router.HandleFunc("/newpost", webAPI.NewPost)
	router.HandleFunc("/profile", webAPI.DisplayProfile)
	
	// Routes API
	router.HandleFunc("/api/register", webAPI.RegisterApi)
	router.HandleFunc("/api/login", webAPI.LoginApi)
	router.HandleFunc("/api/logout", webAPI.LogoutAPI)
	router.HandleFunc("/api/createpost", webAPI.CreatePostApi)
	router.HandleFunc("/api/comments", webAPI.CommentsApi)
	router.HandleFunc("/api/vote", webAPI.VoteApi)
	
	// Routes pour l'édition et la suppression
	router.HandleFunc("/api/editpost", webAPI.EditPostHandler)
	router.HandleFunc("/api/deletepost", webAPI.DeletePostHandler)
	router.HandleFunc("/api/editcomment", webAPI.EditCommentHandler)
	router.HandleFunc("/api/deletecomment", webAPI.DeleteCommentHandler)
	
	// Routes pour la gestion du profil
	router.HandleFunc("/api/editprofile", webAPI.EditProfileHandler)
	router.HandleFunc("/api/changepassword", webAPI.ChangePasswordHandler)
	
	// Nouvelles routes pour l'upload d'images
	router.HandleFunc("/api/uploadprofileimage", webAPI.UploadProfileImageHandler)
	router.HandleFunc("/api/uploadpostimage", webAPI.UploadPostImageHandler)

	// Route pour les fichiers statiques
	router.Handle("/public/", http.StripPrefix("/public/", fs))

	// Application du middleware de rate limiting
	limitedRouter := webAPI.ApplyMiddleware(router, webAPI.RateLimiterMiddleware)

	// Démarrage du serveur
	http.ListenAndServe(":8000", limitedRouter)
}