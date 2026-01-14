package domain

import (
	"testing"
)

// valid title tests
func TestValidateDiaryTitle_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "single character",
			title: "A",
		},
		{
			name:  "normal title",
			title: "My Diary Entry",
		},
		{
			name:  "max length title",
			title: generateString(MaxDiaryTitleLength),
		},
		{
			name:  "unicode characters",
			title: "日記エントリー 📝",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiaryTitle(tt.title)
			if err != nil {
				t.Fatalf("ValidateDiaryTitle failed: %v", err)
			}
		})
	}
}

// invalid title tests
func TestValidateDiaryTitle_EmptyTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "empty string",
			title: "",
		},
		{
			name:  "only spaces",
			title: "   ",
		},
		{
			name:  "only tabs",
			title: "\t\t\t",
		},
		{
			name:  "only newlines",
			title: "\n\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiaryTitle(tt.title)
			if err == nil {
				t.Fatal("expected validation error for empty title")
			}
		})
	}
}

// invalid title that is too long tests
func TestValidateDiaryTitle_TooLong(t *testing.T) {
	t.Parallel()

	title := generateString(MaxDiaryTitleLength + 1)

	err := ValidateDiaryTitle(title)
	if err == nil {
		t.Fatal("expected validation error for too long title")
	}
}

// valid content tests
func TestValidateDiaryContent_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "single character",
			content: "A",
		},
		{
			name:    "normal content",
			content: "This is my diary entry for today.",
		},
		{
			name:    "multiline content",
			content: "Line 1\nLine 2\nLine 3\n",
		},
		{
			name:    "max length content",
			content: generateString(MaxDiaryContentLength),
		},
		{
			name:    "unicode content",
			content: "今日は良い日でした。😊\n明日も頑張ります。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiaryContent(tt.content)
			if err != nil {
				t.Fatalf("ValidateDiaryContent failed: %v", err)
			}
		})
	}
}

// invalid content tests
func TestValidateDiaryContent_EmptyContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "empty string",
			content: "",
		},
		{
			name:    "only spaces",
			content: "   ",
		},
		{
			name:    "only tabs",
			content: "\t\t\t",
		},
		{
			name:    "only newlines",
			content: "\n\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiaryContent(tt.content)
			if err == nil {
				t.Fatal("expected validation error for empty content")
			}
		})
	}
}

// invalid content that is too long tests
func TestValidateDiaryContent_TooLong(t *testing.T) {
	t.Parallel()

	content := generateString(MaxDiaryContentLength + 1)

	err := ValidateDiaryContent(content)
	if err == nil {
		t.Fatal("expected validation error for too long content")
	}
}

// valid request tests
func TestValidateCreateDiaryRequest_Success(t *testing.T) {
	t.Parallel()

	diary := &Diary{
		Title:   "Valid Title",
		Content: "Valid content with some text",
	}

	err := ValidateCreateDiaryRequest(diary)
	if err != nil {
		t.Fatalf("ValidateCreateDiaryRequest failed: %v", err)
	}
}

// invalid title tests
func TestValidateCreateDiaryRequest_InvalidTitle(t *testing.T) {
	t.Parallel()

	diary := &Diary{
		Title:   "",
		Content: "Valid content",
	}

	err := ValidateCreateDiaryRequest(diary)
	if err == nil {
		t.Fatal("expected validation error for empty title")
	}
}

// invalid content tests
func TestValidateCreateDiaryRequest_InvalidContent(t *testing.T) {
	t.Parallel()

	diary := &Diary{
		Title:   "Valid Title",
		Content: "   ",
	}

	err := ValidateCreateDiaryRequest(diary)
	if err == nil {
		t.Fatal("expected validation error for empty content")
	}
}

// invalid request with both title and content invalid tests
func TestValidateCreateDiaryRequest_InvalidBoth(t *testing.T) {
	t.Parallel()

	diary := &Diary{
		Title:   "",
		Content: "",
	}

	err := ValidateCreateDiaryRequest(diary)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// helper function to generate string of specific length
func generateString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = 'a'
	}
	return string(result)
}
