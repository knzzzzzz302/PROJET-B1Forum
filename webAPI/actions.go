package webAPI

import (
	"FORUM-GO/databaseAPI"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Vote struct {
	PostId int
	Vote   int
}

// CreatePostApi crée un post
func CreatePostApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse les formulaires
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Erreur de ParseForm(): %v", err), http.StatusBadRequest)
		return
	}

	// Vérifie si l'utilisateur est connecté
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Récupérer le cookie de session
	cookie, err := r.Cookie("SESSION")
	if err != nil {
		http.Error(w, "Erreur de cookie SESSION", http.StatusUnauthorized)
		return
	}

	username := databaseAPI.GetUser(database, cookie.Value)
	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.Form["categories[]"]

	// Vérifie si les catégories sont valides
	validCategories := databaseAPI.GetCategories(database)
	for _, category := range categories {
		if !inArray(category, validCategories) {
			http.Error(w, fmt.Sprintf("Catégorie invalide : %s", category), http.StatusBadRequest)
			return
		}
	}

	// Joindre les catégories en une seule chaîne
	stringCategories := strings.Join(categories, ",")

	// Créer le post
	now := time.Now()
	databaseAPI.CreatePost(database, username, title, stringCategories, content, now)
	fmt.Printf("Post créé par %s avec le titre %s à %s\n", username, title, now.Format("2006-01-02 15:04:05"))

	// Redirige vers la page des posts de l'utilisateur
	http.Redirect(w, r, "/filter?by=myposts", http.StatusFound)
	return
}

// CommentsApi crée un commentaire
func CommentsApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse les formulaires
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Erreur de ParseForm(): %v", err), http.StatusBadRequest)
		return
	}

	// Vérifie si l'utilisateur est connecté
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Récupérer le cookie de session
	cookie, err := r.Cookie("SESSION")
	if err != nil {
		http.Error(w, "Erreur de cookie SESSION", http.StatusUnauthorized)
		return
	}

	username := databaseAPI.GetUser(database, cookie.Value)
	postId := r.FormValue("postId")
	content := r.FormValue("content")
	now := time.Now()

	// Convertir postId en entier
	postIdInt, err := strconv.Atoi(postId)
	if err != nil {
		http.Error(w, "Post ID invalide", http.StatusBadRequest)
		return
	}

	// Ajouter le commentaire
	databaseAPI.AddComment(database, username, postIdInt, content, now)
	fmt.Printf("Commentaire créé par %s sur le post %s à %s\n", username, postId, now.Format("2006-01-02 15:04:05"))

	// Rediriger vers le post
	http.Redirect(w, r, "/post?id="+postId, http.StatusFound)
	return
}

// VoteApi permet de voter sur un post
func VoteApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	if !isLoggedIn(r) {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erreur de traitement du formulaire", http.StatusBadRequest)
		return
	}

	cookie, _ := r.Cookie("SESSION")
	username := databaseAPI.GetUser(database, cookie.Value)
	postId := r.FormValue("postId")
	postIdInt, err := strconv.Atoi(postId)
	if err != nil {
		http.Error(w, "Identifiant de post invalide", http.StatusBadRequest)
		return
	}

	vote := r.FormValue("vote")
	voteInt, err := strconv.Atoi(vote)
	if err != nil || (voteInt != 1 && voteInt != -1) {
		http.Error(w, "Valeur de vote invalide", http.StatusBadRequest)
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	logPrefix := fmt.Sprintf("Vote - Utilisateur %s, Post %s, %s: ", username, postId, now)

	if voteInt == 1 {
		if databaseAPI.HasUpvoted(database, username, postIdInt) {
			// Annulation d'un vote positif
			databaseAPI.RemoveVote(database, postIdInt, username)
			databaseAPI.DecreaseUpvotes(database, postIdInt)
			fmt.Println(logPrefix + "Suppression d'un vote positif")
		} else if databaseAPI.HasDownvoted(database, username, postIdInt) {
			// Changement de vote négatif à positif
			databaseAPI.RemoveVote(database, postIdInt, username)
			databaseAPI.DecreaseDownvotes(database, postIdInt)
			databaseAPI.AddVote(database, postIdInt, username, 1)
			databaseAPI.IncreaseUpvotes(database, postIdInt)
			fmt.Println(logPrefix + "Changement de vote négatif à positif")
		} else {
			// Nouveau vote positif
			databaseAPI.AddVote(database, postIdInt, username, 1)
			databaseAPI.IncreaseUpvotes(database, postIdInt)
			fmt.Println(logPrefix + "Nouveau vote positif")
		}
	} else { // voteInt == -1
		if databaseAPI.HasDownvoted(database, username, postIdInt) {
			// Annulation d'un vote négatif
			databaseAPI.RemoveVote(database, postIdInt, username)
			databaseAPI.DecreaseDownvotes(database, postIdInt)
			fmt.Println(logPrefix + "Suppression d'un vote négatif")
		} else if databaseAPI.HasUpvoted(database, username, postIdInt) {
			// Changement de vote positif à négatif
			databaseAPI.RemoveVote(database, postIdInt, username)
			databaseAPI.DecreaseUpvotes(database, postIdInt)
			databaseAPI.AddVote(database, postIdInt, username, -1)
			databaseAPI.IncreaseDownvotes(database, postIdInt)
			fmt.Println(logPrefix + "Changement de vote positif à négatif")
		} else {
			// Nouveau vote négatif
			databaseAPI.AddVote(database, postIdInt, username, -1)
			databaseAPI.IncreaseDownvotes(database, postIdInt)
			fmt.Println(logPrefix + "Nouveau vote négatif")
		}
	}

	http.Redirect(w, r, "/post?id="+strconv.Itoa(postIdInt), http.StatusFound)
}

// EditPostHandler gère l'édition d'un post
func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier si l'utilisateur est connecté
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Analyser le formulaire
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Erreur de ParseForm(): %v", err), http.StatusBadRequest)
		return
	}

	// Récupérer le cookie de session
	cookie, err := r.Cookie("SESSION")
	if err != nil {
		http.Error(w, "Erreur de cookie SESSION", http.StatusUnauthorized)
		return
	}

	// Récupérer les données du formulaire
	username := databaseAPI.GetUser(database, cookie.Value)
	postIdStr := r.FormValue("postId")
	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.Form["categories[]"]

	// Convertir postId en entier
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}

	// Vérifier si l'utilisateur est le propriétaire du post
	if !databaseAPI.IsPostOwner(database, username, postId) {
		http.Error(w, "Non autorisé - Vous n'êtes pas le propriétaire de ce post", http.StatusUnauthorized)
		return
	}

	// Si les catégories sont vides, récupérer les catégories existantes
	var stringCategories string
	if len(categories) == 0 {
		post := databaseAPI.GetPost(database, postIdStr)
		stringCategories = strings.Join(post.Categories, ",")
	} else {
		stringCategories = strings.Join(categories, ",")
	}

	// Mettre à jour le post
	success := databaseAPI.EditPost(database, postId, title, stringCategories, content)
	if !success {
		http.Error(w, "Erreur lors de la mise à jour du post", http.StatusInternalServerError)
		return
	}

	// Rediriger vers le post mis à jour
	http.Redirect(w, r, "/post?id="+postIdStr, http.StatusFound)
}

// DeletePostHandler gère la suppression d'un post
func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier si l'utilisateur est connecté
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Analyser le formulaire
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Erreur de ParseForm(): %v", err), http.StatusBadRequest)
		return
	}

	// Récupérer le cookie de session
	cookie, err := r.Cookie("SESSION")
	if err != nil {
		http.Error(w, "Erreur de cookie SESSION", http.StatusUnauthorized)
		return
	}

	// Récupérer les données du formulaire
	username := databaseAPI.GetUser(database, cookie.Value)
	postIdStr := r.FormValue("postId")

	// Convertir postId en entier
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}

	// Vérifier si l'utilisateur est le propriétaire du post
	if !databaseAPI.IsPostOwner(database, username, postId) {
		http.Error(w, "Non autorisé - Vous n'êtes pas le propriétaire de ce post", http.StatusUnauthorized)
		return
	}

	// Supprimer le post
	success := databaseAPI.DeletePost(database, postId)
	if !success {
		http.Error(w, "Erreur lors de la suppression du post", http.StatusInternalServerError)
		return
	}

	// Rediriger vers la page d'accueil
	http.Redirect(w, r, "/", http.StatusFound)
}

// EditCommentHandler gère l'édition d'un commentaire
func EditCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier si l'utilisateur est connecté
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Analyser le formulaire
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Erreur de ParseForm(): %v", err), http.StatusBadRequest)
		return
	}

	// Récupérer le cookie de session
	cookie, err := r.Cookie("SESSION")
	if err != nil {
		http.Error(w, "Erreur de cookie SESSION", http.StatusUnauthorized)
		return
	}

	// Récupérer les données du formulaire
	username := databaseAPI.GetUser(database, cookie.Value)
	commentIdStr := r.FormValue("commentId")
	postIdStr := r.FormValue("postId")
	content := r.FormValue("content")

	// Convertir commentId en entier
	commentId, err := strconv.Atoi(commentIdStr)
	if err != nil {
		http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
		return
	}

	// Vérifier si l'utilisateur est le propriétaire du commentaire
	if !databaseAPI.IsCommentOwner(database, username, commentId) {
		http.Error(w, "Non autorisé - Vous n'êtes pas le propriétaire de ce commentaire", http.StatusUnauthorized)
		return
	}

	// Mettre à jour le commentaire
	success := databaseAPI.EditComment(database, commentId, content)
	if !success {
		http.Error(w, "Erreur lors de la mise à jour du commentaire", http.StatusInternalServerError)
		return
	}

	// Rediriger vers le post
	http.Redirect(w, r, "/post?id="+postIdStr, http.StatusFound)
}

// DeleteCommentHandler gère la suppression d'un commentaire
func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier si l'utilisateur est connecté
	if !isLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Analyser le formulaire
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Erreur de ParseForm(): %v", err), http.StatusBadRequest)
		return
	}

	// Récupérer le cookie de session
	cookie, err := r.Cookie("SESSION")
	if err != nil {
		http.Error(w, "Erreur de cookie SESSION", http.StatusUnauthorized)
		return
	}

	// Récupérer les données du formulaire
	username := databaseAPI.GetUser(database, cookie.Value)
	commentIdStr := r.FormValue("commentId")
	postIdStr := r.FormValue("postId")

	// Convertir commentId en entier
	commentId, err := strconv.Atoi(commentIdStr)
	if err != nil {
		http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
		return
	}

	// Vérifier si l'utilisateur est le propriétaire du commentaire
	if !databaseAPI.IsCommentOwner(database, username, commentId) {
		http.Error(w, "Non autorisé - Vous n'êtes pas le propriétaire de ce commentaire", http.StatusUnauthorized)
		return
	}

	// Supprimer le commentaire
	success := databaseAPI.DeleteComment(database, commentId)
	if !success {
		http.Error(w, "Erreur lors de la suppression du commentaire", http.StatusInternalServerError)
		return
	}

	// Rediriger vers le post
	http.Redirect(w, r, "/post?id="+postIdStr, http.StatusFound)
}