package databaseAPI

import (
	"database/sql"
	"fmt"      // On va utiliser fmt pour les logs
	"github.com/pquerna/otp/totp"
	"time"     // On va utiliser time pour des timestamps
)

// GenerateMFASecret génère un nouveau secret MFA pour un utilisateur
func GenerateMFASecret(database *sql.DB, username string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Forum Sekkay",
		AccountName: username,
	})
	if err != nil {
		return "", "", err
	}
	
	// Stocke le secret dans la BD
	statement, err := database.Prepare("UPDATE users SET mfa_secret = ? WHERE username = ?")
	if err != nil {
		return "", "", err
	}
	_, err = statement.Exec(key.Secret(), username)
	if err != nil {
		return "", "", err
	}
	
	fmt.Printf("MFA Secret généré pour l'utilisateur %s à %s\n", username, time.Now().Format("2006-01-02 15:04:05"))
	
	// Retourne le secret et l'URL pour le QR code
	return key.Secret(), key.URL(), nil
}

// VerifyMFACode vérifie si un code MFA est valide
func VerifyMFACode(database *sql.DB, username string, code string) (bool, error) {
	var secret string
	err := database.QueryRow("SELECT mfa_secret FROM users WHERE username = ?", username).Scan(&secret)
	if err != nil {
		return false, err
	}
	
	// Si pas de secret MFA ou secret vide, MFA n'est pas configuré
	if secret == "" {
		return false, nil
	}
	
	// Vérifier le code avec validation de temps
	currentTime := time.Now()
	valid := totp.Validate(code, secret)
	
	if valid {
		fmt.Printf("Code MFA valide pour %s à %s\n", username, currentTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("Tentative de code MFA invalide pour %s à %s\n", username, currentTime.Format("2006-01-02 15:04:05"))
	}
	
	// Vérifier le code
	return valid, nil
}

// IsMFAEnabled vérifie si MFA est activé pour un utilisateur
func IsMFAEnabled(database *sql.DB, username string) (bool, error) {
	var secret string
	err := database.QueryRow("SELECT mfa_secret FROM users WHERE username = ?", username).Scan(&secret)
	if err != nil {
		return false, err
	}
	
	// Si le secret existe et n'est pas vide, MFA est activé
	return secret != "", nil
}

// DisableMFA désactive le MFA pour un utilisateur
func DisableMFA(database *sql.DB, username string) error {
	statement, err := database.Prepare("UPDATE users SET mfa_secret = '' WHERE username = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(username)
	
	if err == nil {
		fmt.Printf("MFA désactivé pour l'utilisateur %s à %s\n", username, time.Now().Format("2006-01-02 15:04:05"))
	}
	
	return err
}