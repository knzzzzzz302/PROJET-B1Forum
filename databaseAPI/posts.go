package databaseAPI

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"strings"
	"time"
)

// GetPost récupère un post par son ID
func GetPost(database *sql.DB, id string) Post {
	rows, _ := database.Query("SELECT username, title, categories, content, created_at, upvotes, downvotes FROM posts WHERE id = ?", id)
	var post Post
	post.Id, _ = strconv.Atoi(id)
	for rows.Next() {
		catString := ""
		rows.Scan(&post.Username, &post.Title, &catString, &post.Content, &post.CreatedAt, &post.UpVotes, &post.DownVotes)
		categoriesArray := strings.Split(catString, ",")
		post.Categories = categoriesArray
	}
	
	// Récupérer les images du post
	post.Images = GetPostImages(database, post.Id)
	
	return post
}

// GetPostImages récupère les images d'un post
func GetPostImages(database *sql.DB, postId int) []string {
	rows, err := database.Query("SELECT image_path FROM post_images WHERE post_id = ?", postId)
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	
	var images []string
	for rows.Next() {
		var imagePath string
		err := rows.Scan(&imagePath)
		if err != nil {
			continue
		}
		images = append(images, imagePath)
	}
	
	return images
}

// AddPostImage ajoute une image à un post
func AddPostImage(database *sql.DB, postId int, imagePath string) error {
	statement, err := database.Prepare("INSERT INTO post_images (post_id, image_path) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()
	
	_, err = statement.Exec(postId, imagePath)
	return err
}

// GetComments récupère les commentaires d'un post
func GetComments(database *sql.DB, id string) []Comment {
	rows, _ := database.Query("SELECT id, username, content, created_at FROM comments WHERE post_id = ?", id)
	var comments []Comment
	for rows.Next() {
		var comment Comment
		rows.Scan(&comment.Id, &comment.Username, &comment.Content, &comment.CreatedAt)
		comments = append(comments, comment)
	}
	return comments
}

// GetPostsByCategory récupère tous les posts d'une catégorie
func GetPostsByCategory(database *sql.DB, category string) []Post {
	rows, _ := database.Query("SELECT id, username, title, categories, content, created_at, upvotes, downvotes FROM posts WHERE categories LIKE ?", "%"+category+"%")
	var posts []Post
	for rows.Next() {
		var post Post
		var catString string
		rows.Scan(&post.Id, &post.Username, &post.Title, &catString, &post.Content, &post.CreatedAt, &post.UpVotes, &post.DownVotes)
		post.Categories = strings.Split(catString, ",")
		post.Images = GetPostImages(database, post.Id)
		posts = append(posts, post)
	}
	return posts
}

// GetPostsByCategories récupère tous les posts pour toutes les catégories
func GetPostsByCategories(database *sql.DB) [][]Post {
	categories := GetCategories(database)
	var posts [][]Post
	for _, category := range categories {
		posts = append(posts, GetPostsByCategory(database, category))
	}
	return posts
}

// GetPostsByUser récupère tous les posts d'un utilisateur
func GetPostsByUser(database *sql.DB, username string) []Post {
	rows, _ := database.Query("SELECT id, username, title, categories, content, created_at, upvotes, downvotes FROM posts WHERE username = ?", username)
	var posts []Post
	for rows.Next() {
		var post Post
		var catString string
		rows.Scan(&post.Id, &post.Username, &post.Title, &catString, &post.Content, &post.CreatedAt, &post.UpVotes, &post.DownVotes)
		post.Categories = strings.Split(catString, ",")
		post.Images = GetPostImages(database, post.Id)
		posts = append(posts, post)
	}
	return posts
}

// GetLikedPosts récupère les posts qu'un utilisateur a likés
func GetLikedPosts(database *sql.DB, username string) []Post {
	rows, _ := database.Query("SELECT id, username, title, categories, content, created_at, upvotes, downvotes FROM posts WHERE id IN (SELECT post_id FROM votes WHERE username = ? AND vote = 1)", username)
	var posts []Post
	for rows.Next() {
		var post Post
		var catString string
		rows.Scan(&post.Id, &post.Username, &post.Title, &catString, &post.Content, &post.CreatedAt, &post.UpVotes, &post.DownVotes)
		post.Categories = strings.Split(catString, ",")
		post.Images = GetPostImages(database, post.Id)
		posts = append(posts, post)
	}
	return posts
}

// GetCategories récupère toutes les catégories
func GetCategories(database *sql.DB) []string {
	rows, _ := database.Query("SELECT name FROM categories")
	var categories []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		categories = append(categories, name)
	}
	return categories
}

// GetCategoriesIcons récupère toutes les icônes des catégories
func GetCategoriesIcons(database *sql.DB) []string {
	rows, _ := database.Query("SELECT icon FROM categories")
	var icons []string
	for rows.Next() {
		var icon string
		rows.Scan(&icon)
		icons = append(icons, icon)
	}
	return icons
}

// GetCategoryIcon récupère l'icône d'une catégorie
func GetCategoryIcon(database *sql.DB, category string) string {
	rows, _ := database.Query("SELECT icon FROM categories WHERE name = ?", category)
	var icon string
	for rows.Next() {
		rows.Scan(&icon)
	}
	return icon
}

// CreatePost crée un nouveau post
func CreatePost(database *sql.DB, username string, title string, categories string, content string, createdAt time.Time) int64 {
	createdAtString := createdAt.Format("2006-01-02 15:04:05")
	statement, _ := database.Prepare("INSERT INTO posts (username, title, categories, content, created_at, upvotes, downvotes) VALUES (?, ?, ?, ?, ?, ?, ?)")
	result, _ := statement.Exec(username, title, categories, content, createdAtString, 0, 0)
	postId, _ := result.LastInsertId()
	return postId
}

// AddComment ajoute un commentaire à un post
func AddComment(database *sql.DB, username string, postId int, content string, createdAt time.Time) {
	createdAtString := createdAt.Format("2006-01-02 15:04:05")
	statement, _ := database.Prepare("INSERT INTO comments (username, post_id, content, created_at) VALUES (?, ?, ?, ?)")
	statement.Exec(username, postId, content, createdAtString)
}

// EditPost modifie un post existant
func EditPost(database *sql.DB, postId int, title string, categories string, content string) bool {
	statement, err := database.Prepare("UPDATE posts SET title = ?, categories = ?, content = ? WHERE id = ?")
	if err != nil {
		return false
	}
	_, err = statement.Exec(title, categories, content, postId)
	if err != nil {
		return false
	}
	return true
}

// DeletePost supprime un post et ses commentaires associés
func DeletePost(database *sql.DB, postId int) bool {
	// Supprimer d'abord les images associées
	statementImages, err := database.Prepare("DELETE FROM post_images WHERE post_id = ?")
	if err != nil {
		return false
	}
	_, err = statementImages.Exec(postId)
	if err != nil {
		return false
	}
	
	// Supprimer les votes associés
	statementVotes, err := database.Prepare("DELETE FROM votes WHERE post_id = ?")
	if err != nil {
		return false
	}
	_, err = statementVotes.Exec(postId)
	if err != nil {
		return false
	}
	
	// Supprimer ensuite les commentaires associés
	statementComments, err := database.Prepare("DELETE FROM comments WHERE post_id = ?")
	if err != nil {
		return false
	}
	_, err = statementComments.Exec(postId)
	if err != nil {
		return false
	}
	
	// Enfin, supprimer le post lui-même
	statementPost, err := database.Prepare("DELETE FROM posts WHERE id = ?")
	if err != nil {
		return false
	}
	_, err = statementPost.Exec(postId)
	if err != nil {
		return false
	}
	
	return true
}

// EditComment modifie un commentaire
func EditComment(database *sql.DB, commentId int, content string) bool {
	statement, err := database.Prepare("UPDATE comments SET content = ? WHERE id = ?")
	if err != nil {
		return false
	}
	_, err = statement.Exec(content, commentId)
	if err != nil {
		return false
	}
	return true
}

// DeleteComment supprime un commentaire
func DeleteComment(database *sql.DB, commentId int) bool {
	statement, err := database.Prepare("DELETE FROM comments WHERE id = ?")
	if err != nil {
		return false
	}
	_, err = statement.Exec(commentId)
	if err != nil {
		return false
	}
	return true
}

// IsPostOwner vérifie si l'utilisateur est le propriétaire du post
func IsPostOwner(database *sql.DB, username string, postId int) bool {
	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ? AND username = ?", postId, username).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// IsCommentOwner vérifie si l'utilisateur est le propriétaire du commentaire
func IsCommentOwner(database *sql.DB, username string, commentId int) bool {
	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM comments WHERE id = ? AND username = ?", commentId, username).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}