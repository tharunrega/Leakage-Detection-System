package scanner

import (
	"encoding/json"
	"os"

	"stackguard-detector/internal/models"
)

func LoadInventory(path string) ([]models.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tokens []models.Token
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}
