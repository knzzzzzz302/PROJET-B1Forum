package main

import (
	"FORUM-GO/databaseAPI"
	"FORUM-GO/webAPI"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
)

func main() {
	// check if DB exists
	var _, err = os.Stat("database.db")

	// create DB if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create("database.db")
		if err != nil {
			return
		}
		defer file.Close()
	}

	database, _ := sql.Open("sqlite3", "./database.db")

	// Initialisation de la base de données
	databaseAPI.CreateUsersTable(database)
	databaseAPI.AddProfileImageColumnIfNotExists(database)
	databaseAPI.CreatePostTable(database)
	databaseAPI.CreateCommentTable(database)
	databaseAPI.CreateVoteTable(database)
	databaseAPI.CreateCategoriesTable(database)
	databaseAPI.CreateCategories(database)
	databaseAPI.CreateCategoriesIcons(database)
	databaseAPI.CreateCommentLikesTable(database)
	databaseAPI.CreateCommentDislikesTable(database)
	
	// Créer le dossier pour stocker les images des profils
	os.MkdirAll("public/uploads/profiles", os.ModePerm)

	webAPI.SetDatabase(database)

	fs := http.FileServer(http.Dir("public"))
	router := http.NewServeMux()

	// Configuration des routes
	router.HandleFunc("/", webAPI.Index)
	router.HandleFunc("/register", webAPI.Register)
	router.HandleFunc("/login", webAPI.Login)
	router.HandleFunc("/post", webAPI.DisplayPost)
	router.HandleFunc("/filter", webAPI.GetPostsByApi)
	router.HandleFunc("/newpost", webAPI.NewPost)
	router.HandleFunc("/api/register", webAPI.RegisterApi)
	router.HandleFunc("/api/login", webAPI.LoginApi)
	router.HandleFunc("/api/logout", webAPI.LogoutAPI)
	router.HandleFunc("/api/createpost", webAPI.CreatePostApi)
	router.HandleFunc("/api/comments", webAPI.CommentsApi)
	router.HandleFunc("/api/vote", webAPI.VoteApi)
	router.HandleFunc("/api/deletepost", webAPI.DeletePostHandler)
    router.HandleFunc("/profile", webAPI.DisplayProfile)
    router.HandleFunc("/api/editprofile", webAPI.EditProfileHandler)
    router.HandleFunc("/api/changepassword", webAPI.ChangePasswordHandler)
    router.HandleFunc("/api/uploadprofileimage", webAPI.UploadProfileImageHandler)
	router.HandleFunc("/api/editcomment", webAPI.EditCommentHandler)
	router.Handle("/public/", http.StripPrefix("/public/", fs))
	router.HandleFunc("/editpost", webAPI.EditPostPage)
	router.HandleFunc("/api/editpost", webAPI.EditPostHandler)
	router.HandleFunc("/api/deletecomment", webAPI.DeleteCommentHandler)
	router.HandleFunc("/api/commentlike", webAPI.CommentLikeApi)
	router.HandleFunc("/auth/google/login", webAPI.GoogleLogin)
	router.HandleFunc("/auth/google/callback", webAPI.GoogleCallback)
	router.HandleFunc("/auth/github/login", webAPI.GitHubLogin)
	router.HandleFunc("/auth/github/callback", webAPI.GitHubCallback)


	// Flags pour configurer le mode HTTP/HTTPS
	var useHTTPS = flag.Bool("https", false, "Démarrer le serveur en mode HTTPS")
	var port = flag.String("port", "3030", "Port d'écoute du serveur")
	var certFile = flag.String("cert", "certs/cert.pem", "Chemin vers le fichier de certificat SSL")
	var keyFile = flag.String("key", "certs/key.pem", "Chemin vers le fichier de clé privée SSL")
	flag.Parse()

	addr := ":" + *port

	if *useHTTPS {
		fmt.Printf("Démarrage du serveur HTTPS sur https://localhost%s/\n", addr)
		log.Fatal(http.ListenAndServeTLS(addr, *certFile, *keyFile, router))
	} else {
		fmt.Printf("Démarrage du serveur HTTP sur http://localhost%s/\n", addr)
		log.Fatal(http.ListenAndServe(addr, router))
	}
}