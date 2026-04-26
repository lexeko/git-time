package model

import "time"

type Commit struct {
	Hash        string    `json:"hash"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	Time        time.Time `json:"time"`
}

type Session struct {
	AuthorEmail string    `json:"author_email"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Commits     int       `json:"commits"`
	Minutes     int       `json:"minutes"`
}

type AuthorResult struct {
	Email        string    `json:"email"`
	TotalMinutes int       `json:"total_minutes"`
	TotalHours   int       `json:"total_hours"`
	Sessions     []Session `json:"sessions,omitempty"`
}

type Result struct {
	TotalMinutes int            `json:"total_minutes"`
	TotalHours   int            `json:"total_hours"`
	Authors      []AuthorResult `json:"authors,omitempty"`
}
