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
func GetCommentsByPostIDWithLikes(database *sql.DB, postId string, username string) []Comment {
    query := `
        SELECT c.id, c.post_id, c.username, c.content, c.created_at,
               (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id) as likes,
               (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id AND user_id = (SELECT id FROM users WHERE username = ?)) as user_liked
        FROM comments c
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC`
    
    rows, _ := database.Query(query, username, postId)
    defer rows.Close()
    
    var comments []Comment
    for rows.Next() {
        var comment Comment
        var userLiked int
        
        rows.Scan(&comment.Id, &comment.PostId, &comment.Username, &comment.Content, &comment.CreatedAt, &comment.Likes, &userLiked)
        
        comment.UserLiked = userLiked > 0
        comments = append(comments, comment)
    }
    
    return comments
}