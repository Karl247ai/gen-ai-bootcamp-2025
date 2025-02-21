package validator

import (
    "github.com/karl247ai/lang-portal/internal/models"
    "errors"
    "strings"
)

func ValidateWord(word *models.Word) error {
    if strings.TrimSpace(word.Japanese) == "" {
        return errors.New("japanese is required")
    }
    if strings.TrimSpace(word.Romaji) == "" {
        return errors.New("romaji is required")
    }
    if strings.TrimSpace(word.English) == "" {
        return errors.New("english is required")
    }
    return nil
}