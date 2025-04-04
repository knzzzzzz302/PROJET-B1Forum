package webAPI

import (
    "FORUM-GO/databaseAPI"
    "fmt"
    "html/template"
    "net/http"
)

// Structure pour la page de profil
type ProfilePage struct {
    User         User
    Username     string
    Email        string
    Message      string
    ProfileImage string
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
    
    message := r.URL.Query().Get("msg")
    
    payload := ProfilePage{
        User:         User{IsLoggedIn: true, Username: username},
        Username:     username,
        Email:        email,
        Message:      message,
        ProfileImage: profileImage,
    }
    
    // Créer des fonctions pour les templates
    funcMap := template.FuncMap{
        "GetProfileImage": func(username string) string {
            return databaseAPI.GetProfileImage(database, username)
        },
    }
    
    t, err := template.New("").Funcs(funcMap).ParseFiles("public/HTML/profile.html")
    if err != nil {
        http.Error(w, "Erreur lors du chargement de la page: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    err = t.ExecuteTemplate(w, "profile.html", payload)
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