package domain

import (
	"github.com/furuya-3150/fam-diary-log/pkg/validation"
)

func ValidateDiaryTitle(title string) error {
	return validation.NotEmptyAndMaxLength(title, MaxDiaryTitleLength, "title")
}

func ValidateDiaryContent(content string) error {
	return validation.NotEmptyAndMaxLength(content, MaxDiaryContentLength, "content")
}

func ValidateCreateDiaryRequest(req *Diary) error {
	if err := ValidateDiaryTitle(req.Title); err != nil {
		return err
	}

	if err := ValidateDiaryContent(req.Content); err != nil {
		return err
	}

	return nil
}
