package webAPI

import (
    "FORUM-GO/databaseAPI"
    "fmt"
    "html/template"
    "io"
    "net/http"
    "os"
    "time"
)

// Structure pour la page de profil
type ProfilePage struct {
    User          User
    Username      string
    Email         string
    ProfileImage  string
    PostCount     int
    CommentCount  int
    LikesReceived int
    RecentPosts   []databaseAPI.Post
    Message       string
}

// DisplayProfile affiche la page de profil de l'utilisateur
func DisplayProfile(w http.ResponseWriter, r *http.Request) {
    if !isLoggedIn(r) {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    cookie, _ := r.Cookie("SESSION")
    username := databaseAPI.GetUser(database, cookie.Value)
    
    username, email := databaseAPI.GetUserByUsername(database, username)
    profileImage := databaseAPI.GetProfileImage(database, username)
    
    postCount, commentCount, likesReceived := databaseAPI.GetUserStats(database, username)
    recentPosts := databaseAPI.GetRecentPosts(database, username, 5)
    
    message := r.URL.Query().Get("msg")
    
    payload := ProfilePage{
        User:          User{IsLoggedIn: true, Username: username},
        Username:      username,
        Email:         email,
        ProfileImage:  profileImage,
        PostCount:     postCount,
        CommentCount:  commentCount,
        LikesReceived: likesReceived,
        RecentPosts:   recentPosts,
        Message:       message,
    }
    
    t, err := template.ParseFiles("public/HTML/profile.html")
    if err != nil {
        http.Error(w, "Erreur lors du chargement de la page: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    err = t.Execute(w, payload)
    if err != nil {
        http.Error(w, "Erreur lors de l'affichage de la page: "+err.Error(), http.StatusInternalServerError)
    }
}

// EditProfileHandler traite les requêtes d'édition de profil
func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
        return
    }

    if !isLoggedIn(r) {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, fmt.Sprintf("Erreur de ParseForm(): %v", err), http.StatusBadRequest)
        return
    }

    cookie, _ := r.Cookie("SESSION")
    username := databaseAPI.GetUser(database, cookie.Value)
    
    newUsername := r.FormValue("username")
    email := r.FormValue("email")
    
    if newUsername == "" || email == "" {
        http.Redirect(w, r, "/profile?msg=empty_fields", http.StatusFound)
        return
    }
    
    success := databaseAPI.EditUserProfile(database, username, newUsername, email)
    if !success {
        http.Redirect(w, r, "/profile?msg=update_failed", http.StatusFound)
        return
    }
    
    http.Redirect(w, r, "/profile?msg=profile_updated", http.StatusFound)
}

// ChangePasswordHandler traite les requêtes de changement de mot de passe
func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
        return
    }

    if !isLoggedIn(r) {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, fmt.Sprintf("Erreur de ParseForm(): %v", err), http.StatusBadRequest)
        return
    }

    cookie, _ := r.Cookie("SESSION")
    username := databaseAPI.GetUser(database, cookie.Value)
    
    currentPassword := r.FormValue("current_password")
    newPassword := r.FormValue("new_password")
    confirmPassword := r.FormValue("confirm_password")
    
    if currentPassword == "" || newPassword == "" || confirmPassword == "" {
        http.Redirect(w, r, "/profile?msg=empty_password_fields", http.StatusFound)
        return
    }
    
    if newPassword != confirmPassword {
        http.Redirect(w, r, "/profile?msg=passwords_dont_match", http.StatusFound)
        return
    }
    
    success := databaseAPI.ChangePassword(database, username, currentPassword, newPassword)
    if !success {
        http.Redirect(w, r, "/profile?msg=password_change_failed", http.StatusFound)
        return
    }
    
    http.Redirect(w, r, "/profile?msg=password_changed", http.StatusFound)
}

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
    
    // Création du dossier d'uploads si nécessaire
    uploadDir := "public/uploads/profiles"
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        http.Error(w, fmt.Sprintf("Erreur lors de la création du dossier: %v", err), http.StatusInternalServerError)
        return
    }
    
    // Création du fichier de destination
    dst, err := os.Create(fmt.Sprintf("%s/%s", uploadDir, filename))
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
    
    // Mise à jour de l'image de profil
    success := databaseAPI.UpdateProfileImage(database, username, filename)
    if !success {
        http.Error(w, "Erreur lors de la mise à jour de l'image de profil", http.StatusInternalServerError)
        return
    }
    
    // Redirection
    http.Redirect(w, r, "/profile", http.StatusFound)
}