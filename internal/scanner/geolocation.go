package scanner

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type gitHubUser struct {
	Login    string `json:"login"`
	Location string `json:"location"`
}

func GetUserLocation(username string) (string, error) {
	if username == "" {
		return "Unknown", nil
	}
	apiURL := fmt.Sprintf("https://api.github.com/users/%s", username)
	req, _ := http.NewRequest("GET", apiURL, nil)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return "Unknown", nil
	}

	body, _ := io.ReadAll(res.Body)
	var user gitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return "Unknown", nil
	}
	if user.Location == "" {
		return "Unknown", nil
	}
	return user.Location, nil
}
