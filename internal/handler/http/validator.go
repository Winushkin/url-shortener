package http


import (
	"fmt"
	"net/url"
)

func IsValidURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func main() {
	urls := []string{
		"https://example.com",
		"ftp://192.168.1.1",
		"invalid-url",
	}

	for _, u := range urls {
		fmt.Printf("%-40s | %v\n", u, IsValidURL(u))
	}
}