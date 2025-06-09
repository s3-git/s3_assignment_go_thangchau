package utils

import (
	"assignment/internal/domain/entities"
	"regexp"
	"sort"
)

// ExtractEmailsFromText extracts email addresses mentioned in the text
func ExtractEmailsFromText(text string) []string {
	// Simple regex to find email patterns in text
	emailRegex := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	re := regexp.MustCompile(emailRegex)
	return re.FindAllString(text, -1)
}

// SortUsersByEmail sorts a slice of User entities by email in alphabetical order
func SortUsersByEmail(users []*entities.User) {
	sort.SliceStable(users, func(i, j int) bool {
		return users[i].Email < users[j].Email
	})
}