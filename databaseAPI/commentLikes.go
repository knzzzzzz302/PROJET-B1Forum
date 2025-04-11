package databaseAPI

import (
    "database/sql"
    "time"
)

// LikeComment ajoute ou supprime un like sur un commentaire
func LikeComment(database *sql.DB, commentId int, username string) {
    // Vérifier si l'utilisateur a déjà liké ce commentaire
    if HasLikedComment(database, username, commentId) {
        // Supprimer le like existant
        statement, _ := database.Prepare("DELETE FROM comment_likes WHERE comment_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)")
        statement.Exec(commentId, username)
    } else {
        // Ajouter un nouveau like
        statement, _ := database.Prepare("INSERT INTO comment_likes (comment_id, user_id, created_at) VALUES (?, (SELECT id FROM users WHERE username = ?), ?)")
        statement.Exec(commentId, username, time.Now().Format("2006-01-02 15:04:05"))
    }
}

// HasLikedComment vérifie si un utilisateur a liké un commentaire
func HasLikedComment(database *sql.DB, username string, commentId int) bool {
    rows, _ := database.Query(`
        SELECT COUNT(*) FROM comment_likes 
        WHERE comment_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)`, 
        commentId, username)
    
    var count int
    for rows.Next() {
        rows.Scan(&count)
    }
    
    return count > 0
}

// GetCommentLikes récupère le nombre de likes pour un commentaire
func GetCommentLikes(database *sql.DB, commentId int) int {
    rows, _ := database.Query("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ?", commentId)
    
    var count int
    for rows.Next() {
        rows.Scan(&count)
    }
    
    return count
}

// GetCommentsByPostIDWithLikes récupère les commentaires d'un post avec les informations de likes
// GetCommentsByPostIDWithLikes récupère les commentaires d'un post avec les informations de likes et dislikes
func GetCommentsByPostIDWithLikes(database *sql.DB, postId string, username string) []Comment {
    query := `
        SELECT c.id, c.post_id, c.username, c.content, c.created_at,
               (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id) as likes,
               (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id AND user_id = (SELECT id FROM users WHERE username = ?)) as user_liked,
               (SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = c.id) as dislikes,
               (SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = c.id AND user_id = (SELECT id FROM users WHERE username = ?)) as user_disliked
        FROM comments c
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC`
    
    rows, _ := database.Query(query, username, username, postId)
    defer rows.Close()
    
    var comments []Comment
    for rows.Next() {
        var comment Comment
        var userLiked int
        var userDisliked int
        
        rows.Scan(&comment.Id, &comment.PostId, &comment.Username, &comment.Content, &comment.CreatedAt, &comment.Likes, &userLiked, &comment.Dislikes, &userDisliked)
        
        comment.UserLiked = userLiked > 0
        comment.UserDisliked = userDisliked > 0
        comments = append(comments, comment)
    }
    
    return comments
}


// DislikeComment ajoute ou supprime un dislike sur un commentaire
func DislikeComment(database *sql.DB, commentId int, username string) {
    // Vérifier si l'utilisateur a déjà disliké ce commentaire
    if HasDislikedComment(database, username, commentId) {
        // Supprimer le dislike existant
        statement, _ := database.Prepare("DELETE FROM comment_dislikes WHERE comment_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)")
        statement.Exec(commentId, username)
    } else {
        // Si l'utilisateur a liké le commentaire, supprimer d'abord le like
        if HasLikedComment(database, username, commentId) {
            statement, _ := database.Prepare("DELETE FROM comment_likes WHERE comment_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)")
            statement.Exec(commentId, username)
        }
        
        // Ajouter un nouveau dislike
        statement, _ := database.Prepare("INSERT INTO comment_dislikes (comment_id, user_id, created_at) VALUES (?, (SELECT id FROM users WHERE username = ?), ?)")
        statement.Exec(commentId, username, time.Now().Format("2006-01-02 15:04:05"))
    }
}

// HasDislikedComment vérifie si un utilisateur a disliké un commentaire
func HasDislikedComment(database *sql.DB, username string, commentId int) bool {
    rows, _ := database.Query(`
        SELECT COUNT(*) FROM comment_dislikes 
        WHERE comment_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)`, 
        commentId, username)
    
    var count int
    for rows.Next() {
        rows.Scan(&count)
    }
    
    return count > 0
}

// GetCommentDislikes récupère le nombre de dislikes pour un commentaire
func GetCommentDislikes(database *sql.DB, commentId int) int {
    rows, _ := database.Query("SELECT COUNT(*) FROM comment_dislikes WHERE comment_id = ?", commentId)
    
    var count int
    for rows.Next() {
        rows.Scan(&count)
    }
    
    return count
}