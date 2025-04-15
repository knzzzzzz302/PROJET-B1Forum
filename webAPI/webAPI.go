package webAPI

import (
	"FORUM-GO/databaseAPI"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	IsLoggedIn bool
	Username   string
}

type HomePage struct {
	User              User
	Categories        []string
	Icons             []string
	PostsByCategories [][]databaseAPI.Post
}

type PostsPage struct {
	User  User
	Title string
	Posts []databaseAPI.Post
	Icon  string
}

type PostPage struct {
	User User
	Post databaseAPI.Post
}

type EditPostPageData struct {
	User User
	Post databaseAPI.Post
}

var database *sql.DB

func SetDatabase(db *sql.DB) {
	database = db
}

// Index displays the Index page
func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if checkUserLoggedIn(r) {
		cookie, _ := r.Cookie("SESSION")
		payload := HomePage{
			User:              User{IsLoggedIn: true, Username: databaseAPI.GetUser(database, cookie.Value)},
			Categories:        databaseAPI.GetCategories(database),
			Icons:             databaseAPI.GetCategoriesIcons(database),
			PostsByCategories: databaseAPI.GetPostsByCategories(database),
		}
		t, _ := template.ParseGlob("public/HTML/*.html")
		t.ExecuteTemplate(w, "forum.html", payload)
		return
	}
	payload := HomePage{
		User:              User{IsLoggedIn: false},
		Categories:        databaseAPI.GetCategories(database),
		Icons:             databaseAPI.GetCategoriesIcons(database),
		PostsByCategories: databaseAPI.GetPostsByCategories(database),
	}
	t, _ := template.ParseGlob("public/HTML/*.html")
	t.ExecuteTemplate(w, "forum.html", payload)
	return
}

// DisplayPost displays a post on a template
func DisplayPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	post := databaseAPI.GetPost(database, id)
	comments := databaseAPI.GetComments(database, id)
	
	// Information sur l'utilisateur actuel
	var isUserLoggedIn bool
	var username string
	
	if checkUserLoggedIn(r) {
		cookie, _ := r.Cookie("SESSION")
		username = databaseAPI.GetUser(database, cookie.Value)
		isUserLoggedIn = true
		
		// Pour chaque commentaire, vérifier si l'utilisateur l'a liké ou disliké
		for i := range comments {
			// Vérifier si l'utilisateur a liké ce commentaire
			hasLiked := false
			rows, err := database.Query(`
				SELECT COUNT(*) FROM comment_likes 
				JOIN users ON comment_likes.user_id = users.id
				WHERE comment_likes.comment_id = ? AND users.username = ?
			`, comments[i].Id, username)
			
			if err == nil {
				if rows.Next() {
					var count int
					rows.Scan(&count)
					hasLiked = count > 0
				}
				rows.Close()
			}
			
			comments[i].UserLiked = hasLiked
			
			// Vérifier si l'utilisateur a disliké ce commentaire
			hasDisliked := false
			rows, err = database.Query(`
				SELECT COUNT(*) FROM comment_dislikes 
				JOIN users ON comment_dislikes.user_id = users.id
				WHERE comment_dislikes.comment_id = ? AND users.username = ?
			`, comments[i].Id, username)
			
			if err == nil {
				if rows.Next() {
					var count int
					rows.Scan(&count)
					hasDisliked = count > 0
				}
				rows.Close()
			}
			
			comments[i].UserDisliked = hasDisliked
			
			// Obtenir le nombre de likes
			rows, err = database.Query("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ?", comments[i].Id)
			if err == nil {
				if rows.Next() {
					var count int
					rows.Scan(&count)
					comments[i].Likes = count
				}
				rows.Close()
			}
			
			// Obtenir le nombre de dislikes
			rows, err = database.Query("SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = ?", comments[i].Id)
			if err == nil {
				if rows.Next() {
					var count int
					rows.Scan(&count)
					comments[i].Dislikes = count
				}
				rows.Close()
			}
		}
	} else {
		isUserLoggedIn = false
		
		// Pour chaque commentaire, obtenir juste le nombre de likes et dislikes
		for i := range comments {
			// Obtenir le nombre de likes
			rows, err := database.Query("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ?", comments[i].Id)
			if err == nil {
				if rows.Next() {
					var count int
					rows.Scan(&count)
					comments[i].Likes = count
				}
				rows.Close()
			}
			
			// Obtenir le nombre de dislikes
			rows, err = database.Query("SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = ?", comments[i].Id)
			if err == nil {
				if rows.Next() {
					var count int
					rows.Scan(&count)
					comments[i].Dislikes = count
				}
				rows.Close()
			}
		}
	}
	
	post.Comments = comments
	
	// Récupérer les images du post
	post.Images = databaseAPI.GetPostImages(database, post.Id)
	
	payload := PostPage{
		Post: post,
		User: User{IsLoggedIn: isUserLoggedIn, Username: username},
	}
	
	t, _ := template.ParseGlob("public/HTML/*.html")
	t.ExecuteTemplate(w, "detail.html", payload)
}

// GetPostsByApi gets all post filtered by the given parameters
func GetPostsByApi(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("by")
	if method == "category" {
		category := r.URL.Query().Get("category")
		posts := databaseAPI.GetPostsByCategory(database, category)
		payload := PostsPage{
			Title: "Posts in category " + category,
			Posts: posts,
			Icon:  databaseAPI.GetCategoryIcon(database, category),
		}
		if checkUserLoggedIn(r) {
			payload.User = User{IsLoggedIn: true}
		}
		t, _ := template.ParseGlob("public/HTML/*.html")
		t.ExecuteTemplate(w, "posts.html", payload)
		return
	}
	if method == "myposts" {
		if checkUserLoggedIn(r) {
			cookie, _ := r.Cookie("SESSION")
			username := databaseAPI.GetUser(database, cookie.Value)
			posts := databaseAPI.GetPostsByUser(database, username)
			payload := PostsPage{
				User:  User{IsLoggedIn: true},
				Title: "My posts",
				Posts: posts,
				Icon:  "fa-user",
			}
			t, _ := template.ParseGlob("public/HTML/*.html")
			t.ExecuteTemplate(w, "posts.html", payload)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if method == "liked" {
		if checkUserLoggedIn(r) {
			cookie, _ := r.Cookie("SESSION")
			username := databaseAPI.GetUser(database, cookie.Value)
			posts := databaseAPI.GetLikedPosts(database, username)
			payload := PostsPage{
				User:  User{IsLoggedIn: true},
				Title: "Posts liked by me",
				Posts: posts,
				Icon:  "fa-heart",
			}
			t, _ := template.ParseGlob("public/HTML/*.html")
			t.ExecuteTemplate(w, "posts.html", payload)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// NewPost displays the NewPost page
func NewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !checkUserLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	t, _ := template.ParseGlob("public/HTML/*.html")
	t.ExecuteTemplate(w, "createThread.html", nil)
}

// EditPostPage affiche la page d'édition d'un post
func EditPostPage(w http.ResponseWriter, r *http.Request) {
    if !checkUserLoggedIn(r) {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }
    
    id := r.URL.Query().Get("postId")
    post := databaseAPI.GetPost(database, id)
    
    // Vérifier si l'utilisateur est le propriétaire du post
    cookie, _ := r.Cookie("SESSION")
    username := databaseAPI.GetUser(database, cookie.Value)
    
    postId, err := strconv.Atoi(id)
    if err != nil {
        http.Error(w, "ID de post invalide", http.StatusBadRequest)
        return
    }
    
    if !databaseAPI.IsPostOwner(database, username, postId) {
        http.Error(w, "Non autorisé - Vous n'êtes pas le propriétaire de ce post", http.StatusUnauthorized)
        return
    }
    
    payload := EditPostPageData{
        Post: post,
        User: User{IsLoggedIn: true, Username: username},
    }
    
    t, err := template.ParseFiles("public/HTML/editpost.html")
    if err != nil {
        fmt.Println("Erreur lors du chargement du template:", err)
        http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
        return
    }
    
    err = t.Execute(w, payload)
    if err != nil {
        fmt.Println("Erreur lors de l'exécution du template:", err)
        http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
    }
}

// CommentLikeApi gère les likes et dislikes des commentaires
func CommentLikeApi(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
        return
    }

    // Vérifier si l'utilisateur est connecté
    if !checkUserLoggedIn(r) {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    // Récupérer l'utilisateur courant
    cookie, err := r.Cookie("SESSION")
    if err != nil {
        http.Error(w, "Erreur de session", http.StatusUnauthorized)
        return
    }

    username := databaseAPI.GetUser(database, cookie.Value)
    
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Erreur lors du parsing du formulaire", http.StatusBadRequest)
        return
    }
    
    commentIdStr := r.FormValue("commentId")
    postIdStr := r.FormValue("postId")
    action := r.FormValue("action") // "like" ou "dislike"
    
    commentId, err := strconv.Atoi(commentIdStr)
    if err != nil {
        http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
        return
    }

    // Obtenir l'ID utilisateur
    var userId int
    err = database.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userId)
    if err != nil {
        http.Error(w, "Erreur de récupération d'utilisateur: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    if action == "like" {
        // Vérifier si l'utilisateur a déjà liké ce commentaire
        var likeExists bool
        err = database.QueryRow("SELECT COUNT(*) > 0 FROM comment_likes WHERE comment_id = ? AND user_id = ?", 
            commentId, userId).Scan(&likeExists)
        if err != nil {
            http.Error(w, "Erreur lors de la vérification du like: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Vérifier si l'utilisateur a déjà disliké ce commentaire
        var dislikeExists bool
        err = database.QueryRow("SELECT COUNT(*) > 0 FROM comment_dislikes WHERE comment_id = ? AND user_id = ?", 
            commentId, userId).Scan(&dislikeExists)
        if err != nil {
            http.Error(w, "Erreur lors de la vérification du dislike: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Si un dislike existe, le supprimer d'abord
        if dislikeExists {
            _, err = database.Exec("DELETE FROM comment_dislikes WHERE comment_id = ? AND user_id = ?", 
                commentId, userId)
            if err != nil {
                http.Error(w, "Erreur lors de la suppression du dislike: "+err.Error(), http.StatusInternalServerError)
                return
            }
        }
        
        if likeExists {
            // Si l'utilisateur a déjà liké, supprimer le like (toggle)
            _, err = database.Exec("DELETE FROM comment_likes WHERE comment_id = ? AND user_id = ?", 
                commentId, userId)
        } else {
            // Sinon, ajouter un nouveau like
            _, err = database.Exec("INSERT INTO comment_likes (comment_id, user_id, created_at) VALUES (?, ?, ?)", 
                commentId, userId, time.Now().Format("2006-01-02 15:04:05"))
        }
    } else if action == "dislike" {
        // Vérifier si l'utilisateur a déjà disliké ce commentaire
        var dislikeExists bool
        err = database.QueryRow("SELECT COUNT(*) > 0 FROM comment_dislikes WHERE comment_id = ? AND user_id = ?", 
            commentId, userId).Scan(&dislikeExists)
        if err != nil {
            http.Error(w, "Erreur lors de la vérification du dislike: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Vérifier si l'utilisateur a déjà liké ce commentaire
        var likeExists bool
        err = database.QueryRow("SELECT COUNT(*) > 0 FROM comment_likes WHERE comment_id = ? AND user_id = ?", 
            commentId, userId).Scan(&likeExists)
        if err != nil {
            http.Error(w, "Erreur lors de la vérification du like: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Si un like existe, le supprimer d'abord
        if likeExists {
            _, err = database.Exec("DELETE FROM comment_likes WHERE comment_id = ? AND user_id = ?", 
                commentId, userId)
            if err != nil {
                http.Error(w, "Erreur lors de la suppression du like: "+err.Error(), http.StatusInternalServerError)
                return
            }
        }
        
        if dislikeExists {
            // Si l'utilisateur a déjà disliké, supprimer le dislike (toggle)
            _, err = database.Exec("DELETE FROM comment_dislikes WHERE comment_id = ? AND user_id = ?", 
                commentId, userId)
        } else {
            // Sinon, ajouter un nouveau dislike
            _, err = database.Exec("INSERT INTO comment_dislikes (comment_id, user_id, created_at) VALUES (?, ?, ?)", 
                commentId, userId, time.Now().Format("2006-01-02 15:04:05"))
        }
    }
    
    if err != nil {
        http.Error(w, "Erreur lors du traitement de la réaction: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Rediriger vers la page du post
    http.Redirect(w, r, "/post?id="+postIdStr, http.StatusSeeOther)
}

// checkUserLoggedIn vérifie si l'utilisateur est connecté
func checkUserLoggedIn(r *http.Request) bool {
    cookie, err := r.Cookie("SESSION")
    if err != nil {
        return false
    }
    cookieExists := databaseAPI.CheckCookie(database, cookie.Value)
    if !cookieExists {
        return false
    }
    expires := databaseAPI.GetExpires(database, cookie.Value)
    
    expiresTime, err := time.Parse("2006-01-02 15:04:05", expires)
    if err != nil {
        return false
    }
    
    return !time.Now().After(expiresTime)
}

// inArray check if a string is in an array
func inArray(input string, array []string) bool {
	for _, v := range array {
		if v == input {
			return true
		}
	}
	return false
}