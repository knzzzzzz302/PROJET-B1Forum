package databaseAPI

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    IsLoggedIn bool
    Username   string
}

// GetUser get user by cookie
func GetUser(database *sql.DB, cookie string) string {
    rows, _ := database.Query("SELECT username FROM users WHERE cookie = ?", cookie)
    var username string
    for rows.Next() {
        rows.Scan(&username)
    }
    return username
}

// GetUserInfo returns the username, email and hashed password of a user
func GetUserInfo(database *sql.DB, submittedEmail string) (string, string, string) {
    var user string
    var email string
    var password string
    rows, _ := database.Query("SELECT username, email, password FROM users WHERE email = ?", submittedEmail)
    for rows.Next() {
        rows.Scan(&user, &email, &password)
    }
    return user, email, password
}

// EditUserProfile met à jour le profil d'un utilisateur
func EditUserProfile(database *sql.DB, username string, newUsername string, email string) bool {
    // Vérifier si le nouveau nom d'utilisateur est disponible (si changé)
    if username != newUsername && !UsernameNotTaken(database, newUsername) {
        return false
    }
    
    statement, err := database.Prepare("UPDATE users SET username = ?, email = ? WHERE username = ?")
    if err != nil {
        return false
    }
    _, err = statement.Exec(newUsername, email, username)
    if err != nil {
        return false
    }
    
    // Mettre à jour le nom d'utilisateur dans les posts
    statementPosts, err := database.Prepare("UPDATE posts SET username = ? WHERE username = ?")
    if err != nil {
        return false
    }
    _, err = statementPosts.Exec(newUsername, username)
    if err != nil {
        return false
    }
    
    // Mettre à jour le nom d'utilisateur dans les commentaires
    statementComments, err := database.Prepare("UPDATE comments SET username = ? WHERE username = ?")
    if err != nil {
        return false
    }
    _, err = statementComments.Exec(newUsername, username)
    if err != nil {
        return false
    }
    
    // Mettre à jour le nom d'utilisateur dans les votes
    statementVotes, err := database.Prepare("UPDATE votes SET username = ? WHERE username = ?")
    if err != nil {
        return false
    }
    _, err = statementVotes.Exec(newUsername, username)
    if err != nil {
        return false
    }
    
    return true
}

// ChangePassword change le mot de passe d'un utilisateur
func ChangePassword(database *sql.DB, username string, currentPassword string, newPassword string) bool {
    // Récupérer le mot de passe actuel
    var storedPassword string
    err := database.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedPassword)
    if err != nil {
        return false
    }
    
    // Vérifier si le mot de passe actuel est correct
    if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(currentPassword)); err != nil {
        return false
    }
    
    // Hasher le nouveau mot de passe
    hashedPassword, err := hashPassword(newPassword)
    if err != nil {
        return false
    }
    
    // Mettre à jour le mot de passe
    statement, err := database.Prepare("UPDATE users SET password = ? WHERE username = ?")
    if err != nil {
        return false
    }
    _, err = statement.Exec(hashedPassword, username)
    if err != nil {
        return false
    }
    
    return true
}

// GetUserByUsername récupère les informations d'un utilisateur par son nom d'utilisateur
func GetUserByUsername(database *sql.DB, username string) (string, string) {
    var email string
    rows, _ := database.Query("SELECT email FROM users WHERE username = ?", username)
    for rows.Next() {
        rows.Scan(&email)
    }
    return username, email
}

// UpdateProfileImage met à jour l'image de profil d'un utilisateur
func UpdateProfileImage(database *sql.DB, username string, imagePath string) bool {
    statement, err := database.Prepare("UPDATE users SET profile_image = ? WHERE username = ?")
    if err != nil {
        return false
    }
    _, err = statement.Exec(imagePath, username)
    return err == nil
}

// GetProfileImage récupère le chemin de l'image de profil d'un utilisateur
func GetProfileImage(database *sql.DB, username string) string {
    var imagePath string
    err := database.QueryRow("SELECT profile_image FROM users WHERE username = ?", username).Scan(&imagePath)
    if err != nil || imagePath == "" {
        return "default.png"
    }
    return imagePath
}