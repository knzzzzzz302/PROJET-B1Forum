// Créez un fichier webAPI/mfa.go avec ce contenu

package webAPI

import (
	"FORUM-GO/databaseAPI"
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"net/http"
	"time"
    uuid "github.com/satori/go.uuid"
)

// Structure pour passer les données à la page MFA
type MFASetupData struct {
	User       User
	QRCodeURL  string
	Secret     string
	Error      string
	Success    string
	MFAEnabled bool
}

// Structure pour passer les données à la page de vérification MFA
type MFAVerifyData struct {
	Username  string
	TempToken string
	Error     string
}

// Map pour stocker les jetons temporaires de connexion MFA
var mfaTempTokens = make(map[string]MFATempToken)

// Structure pour stocker les informations de token temporaire
type MFATempToken struct {
	Username  string
	Email     string
	ExpiresAt time.Time
}

// MFASetup affiche la page de configuration MFA
func MFASetup(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	cookie, _ := r.Cookie("SESSION")
	username := databaseAPI.GetUser(database, cookie.Value)
	
	// Vérifier si MFA est déjà activé
	mfaEnabled, err := databaseAPI.IsMFAEnabled(database, username)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification MFA", http.StatusInternalServerError)
		return
	}
	
	data := MFASetupData{
		User:       User{IsLoggedIn: true, Username: username},
		MFAEnabled: mfaEnabled,
		Error:      r.URL.Query().Get("error"),
		Success:    r.URL.Query().Get("success"),
	}
	
	// Si MFA n'est pas activé et qu'on demande de générer un QR code
	if !mfaEnabled && r.URL.Path == "/mfa/setup" && r.Method == "GET" {
		secret, qrURL, err := databaseAPI.GenerateMFASecret(database, username)
		if err != nil {
			http.Error(w, "Erreur lors de la génération du secret MFA", http.StatusInternalServerError)
			return
		}
		
		data.QRCodeURL = qrURL
		data.Secret = secret
	}
	
	t, err := template.ParseFiles("public/HTML/mfa_setup.html")
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	t.Execute(w, data)
}

// MFAVerifySetup vérifie le code MFA lors de la configuration
func MFAVerifySetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erreur de formulaire", http.StatusBadRequest)
		return
	}
	
	cookie, _ := r.Cookie("SESSION")
	username := databaseAPI.GetUser(database, cookie.Value)
	
	code := r.FormValue("code")
	
	// Vérifier le code
	valid, err := databaseAPI.VerifyMFACode(database, username, code)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	if !valid {
		http.Redirect(w, r, "/mfa/setup?error=Code+invalide.+Veuillez+réessayer.", http.StatusFound)
		return
	}
	
	// Code valide, MFA est maintenant configuré
	http.Redirect(w, r, "/mfa/setup?success=L'authentification+à+deux+facteurs+est+maintenant+activée!", http.StatusFound)
}

// MFADisable désactive le MFA
func MFADisable(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	
	cookie, _ := r.Cookie("SESSION")
	username := databaseAPI.GetUser(database, cookie.Value)
	
	// Désactiver MFA
	err := databaseAPI.DisableMFA(database, username)
	if err != nil {
		http.Error(w, "Erreur lors de la désactivation MFA: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	http.Redirect(w, r, "/mfa/setup?success=L'authentification+à+deux+facteurs+a+été+désactivée.", http.StatusFound)
}

// MFALoginCheck vérifie si MFA est requis lors de la connexion
func MFALoginCheck(w http.ResponseWriter, r *http.Request, username string, email string) bool {
	// Vérifier si MFA est activé
	mfaEnabled, err := databaseAPI.IsMFAEnabled(database, username)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification MFA", http.StatusInternalServerError)
		return false
	}
	
	if !mfaEnabled {
		// MFA non activé, continuer avec la connexion normale
		return false
	}
	
	// MFA activé, créer un token temporaire
	token := generateTempToken()
	
	// Stocker le token avec les informations utilisateur
	mfaTempTokens[token] = MFATempToken{
		Username:  username,
		Email:     email,
		ExpiresAt: time.Now().Add(10 * time.Minute), // Expire après 10 minutes
	}
	
	// Rediriger vers la page de vérification MFA
	http.Redirect(w, r, "/mfa/verify?token="+token, http.StatusFound)
	return true
}

// MFAVerify affiche la page de vérification MFA
func MFAVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	error := r.URL.Query().Get("error")
	
	// Vérifier que le token existe
	tempToken, exists := mfaTempTokens[token]
	if !exists || time.Now().After(tempToken.ExpiresAt) {
		http.Redirect(w, r, "/login?err=session_expired", http.StatusFound)
		return
	}
	
	data := MFAVerifyData{
		Username:  tempToken.Username,
		TempToken: token,
		Error:     error,
	}
	
	t, err := template.ParseFiles("public/HTML/mfa_verify.html")
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	t.Execute(w, data)
}

// MFAValidate valide le code MFA et termine la connexion
func MFAValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erreur de formulaire", http.StatusBadRequest)
		return
	}
	
	token := r.FormValue("tempToken")
	code := r.FormValue("code")
	
	// Vérifier que le token existe
	tempToken, exists := mfaTempTokens[token]
	if !exists || time.Now().After(tempToken.ExpiresAt) {
		http.Redirect(w, r, "/login?err=session_expired", http.StatusFound)
		return
	}
	
	// Vérifier le code MFA
	valid, err := databaseAPI.VerifyMFACode(database, tempToken.Username, code)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	if !valid {
		http.Redirect(w, r, "/mfa/verify?token="+token+"&error=Code+invalide.+Veuillez+réessayer.", http.StatusFound)
		return
	}
	
	// Code valide, créer une session
	expiration := time.Now().Add(31 * 24 * time.Hour)
	sessionID := uuid.NewV4().String()
	
	cookie := http.Cookie{
		Name:     "SESSION",
		Value:    sessionID,
		Expires:  expiration,
		Path:     "/",
		HttpOnly: true,
	}
	
	http.SetCookie(w, &cookie)
	
	// Mettre à jour le cookie dans la BD
	databaseAPI.UpdateCookie(database, sessionID, expiration, tempToken.Email)
	
	// Supprimer le token temporaire
	delete(mfaTempTokens, token)
	
	// Rediriger vers la page d'accueil
	http.Redirect(w, r, "/", http.StatusFound)
}

// generateTempToken génère un token aléatoire pour la session temporaire MFA
func generateTempToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}