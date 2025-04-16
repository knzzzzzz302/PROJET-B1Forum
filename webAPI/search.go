package webAPI

import (
	"FORUM-GO/databaseAPI"
	"html/template"
	"net/http"
)

// AdvancedSearch affiche la page de recherche avancée
func AdvancedSearch(w http.ResponseWriter, r *http.Request) {
	var username string
	var isLoggedIn bool
	
	if checkUserLoggedIn(r) {
		cookie, _ := r.Cookie("SESSION")
		username = databaseAPI.GetUser(database, cookie.Value)
		isLoggedIn = true
	}
	
	payload := struct {
		User       User
		Categories []string
	}{
		User:       User{IsLoggedIn: isLoggedIn, Username: username},
		Categories: databaseAPI.GetCategories(database),
	}
	
	t, _ := template.ParseFiles("public/HTML/advanced-search.html")
	t.Execute(w, payload)
}