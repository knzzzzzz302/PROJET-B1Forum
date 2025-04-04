package webAPI

import (
	"FORUM-GO/databaseAPI"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Dossiers de stockage
const (
	ProfilesDir = "public/uploads/profiles/"
	PostsDir    = "public/uploads/posts/"
)

// UploadProfileImageHandler gère l'upload d'images de profil
func UploadProfileImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	// Parsing du formulaire multipart
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Erreur lors du parsing du formulaire", http.StatusBadRequest)
		return
	}
	
	// Récupération du fichier
	file, handler, err := r.FormFile("profile_image")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du fichier", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Création d'un nom de fichier unique
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
	
	// Création du fichier de destination
	os.MkdirAll(ProfilesDir, 0755) // Création du dossier s'il n'existe pas
	dst, err := os.Create(ProfilesDir + filename)
	if err != nil {
		http.Error(w, "Erreur serveur lors de la création du fichier", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	
	// Copie du contenu
	if _, err = io.Copy(dst, file); err != nil {
		http.Error(w, "Erreur lors de l'enregistrement du fichier", http.StatusInternalServerError)
		return
	}
	
	// Mise à jour de la base de données
	cookie, _ := r.Cookie("SESSION")
	username := databaseAPI.GetUser(database, cookie.Value)
	
	// Ajout de cette fonction dans databaseAPI/user.go
	statement, _ := database.Prepare("UPDATE users SET profile_image = ? WHERE username = ?")
	statement.Exec(filename, username)
	
	// Redirection
	http.Redirect(w, r, "/profile", http.StatusFound)
}

// UploadPostImageHandler gère l'upload d'images pour les posts
func UploadPostImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	// Parsing du formulaire multipart
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Erreur lors du parsing du formulaire", http.StatusBadRequest)
		return
	}
	
	// Récupération du postId
	postIdStr := r.FormValue("postId")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}
	
	// Récupération du fichier
	file, handler, err := r.FormFile("post_image")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du fichier", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Création d'un nom de fichier unique
	ext := filepath.Ext(handler.Filename)
	filename := fmt.Sprintf("%d_%d%s", postId, time.Now().Unix(), ext)
	
	// Création du fichier de destination
	os.MkdirAll(PostsDir, 0755) // Création du dossier s'il n'existe pas
	dst, err := os.Create(PostsDir + filename)
	if err != nil {
		http.Error(w, "Erreur serveur lors de la création du fichier", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	
	// Copie du contenu
	if _, err = io.Copy(dst, file); err != nil {
		http.Error(w, "Erreur lors de l'enregistrement du fichier", http.StatusInternalServerError)
		return
	}
	
	// Ajout de l'image à la base de données - simplification
	statement, _ := database.Prepare("INSERT INTO post_images (post_id, image_path) VALUES (?, ?)")
	statement.Exec(postId, filename)
	
	// Redirection
	http.Redirect(w, r, "/post?id="+postIdStr, http.StatusFound)
}