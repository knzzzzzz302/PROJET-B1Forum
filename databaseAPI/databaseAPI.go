package databaseAPI

import (
	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	Id         int
	Username   string
	Title      string
	Categories []string
	Content    string
	CreatedAt  string
	UpVotes    int
	DownVotes  int
	Comments   []Comment
	Images     []string 
}

type Comment struct {
	Id           int
	PostId       int
	Username     string
	Content      string
	CreatedAt    string
	Likes        int
	UserLiked    bool
	Dislikes     int
	UserDisliked bool
}