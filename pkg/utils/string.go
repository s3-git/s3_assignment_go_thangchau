package utils

import "regexp"

// ExtractEmailsFromText extracts email addresses mentioned in the text
func ExtractEmailsFromText(text string) []string {
	// Simple regex to find email patterns in text
	emailRegex := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	re := regexp.MustCompile(emailRegex)
	return re.FindAllString(text, -1)
}