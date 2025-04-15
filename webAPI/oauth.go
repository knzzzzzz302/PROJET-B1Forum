package webAPI

import (
	"FORUM-GO/databaseAPI"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Configuration pour Google OAuth
var googleOauthConfig = &oauth2.Config{
	ClientID:     "1019617553153",
	ClientSecret: "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",
	RedirectURL:  "http://localhost:3030/auth/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// Générer un état aléatoire pour la sécurité OAuth
func generateStateOauthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return state
}

// Gestion de la connexion via Google
func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateStateOauthCookie(w)
	url := googleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Structure pour stocker les infos utilisateur Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// Callback pour Google
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Récupérer le cookie d'état
	oauthState, err := r.Cookie("oauthstate")
	if err != nil {
		http.Error(w, "État de cookie invalide", http.StatusBadRequest)
		return
	}

	// Vérifier l'état pour éviter les attaques CSRF
	if r.FormValue("state") != oauthState.Value {
		http.Error(w, "État invalide", http.StatusBadRequest)
		return
	}

	// Échanger le code contre un token
	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Échec d'échange de code", http.StatusInternalServerError)
		return
	}

	// Récupérer les infos utilisateur
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Error(w, "Échec de récupération des infos utilisateur", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Échec de lecture des données utilisateur", http.StatusInternalServerError)
		return
	}

	// Désérialiser les infos utilisateur
	var userInfo GoogleUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		http.Error(w, "Échec de désérialisation JSON", http.StatusInternalServerError)
		return
	}

	// Vérifier si l'utilisateur existe déjà
	userExists := !databaseAPI.EmailNotTaken(database, userInfo.Email)
	
	if !userExists {
		// Créer un nouvel utilisateur avec des infos Google
		// Générer un mot de passe aléatoire pour ces utilisateurs
		password := generateRandomPassword(16)
		expiration := time.Now().Add(31 * 24 * time.Hour)
		value := uuid.NewV4().String()
		
		// Ajouter l'utilisateur
		databaseAPI.AddUser(database, userInfo.Name, userInfo.Email, password, value, expiration.Format("2006-01-02 15:04:05"))
		fmt.Printf("Nouvel utilisateur créé via Google: %s (%s)\n", userInfo.Name, userInfo.Email)
	} else {
		// Connecter un utilisateur existant
		username, _, _ := databaseAPI.GetUserInfo(database, userInfo.Email)
		expiration := time.Now().Add(31 * 24 * time.Hour)
		value := uuid.NewV4().String()
		
		// Mettre à jour le cookie
		databaseAPI.UpdateCookie(database, value, expiration, userInfo.Email)
		fmt.Printf("Utilisateur connecté via Google: %s (%s)\n", username, userInfo.Email)
	}

	// Définir un cookie de session
	expiration := time.Now().Add(31 * 24 * time.Hour)
	value := uuid.NewV4().String()
	cookie := http.Cookie{Name: "SESSION", Value: value, Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)
	
	// Mettre à jour le cookie dans la BD
	databaseAPI.UpdateCookie(database, value, expiration, userInfo.Email)
	
	// Rediriger vers la page d'accueil
	http.Redirect(w, r, "/", http.StatusFound)
}

// Génère un mot de passe aléatoire pour les utilisateurs OAuth
func generateRandomPassword(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}