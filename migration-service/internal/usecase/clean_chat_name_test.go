package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanChatName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Название с одинарными кавычками и пробелами",
			input:    "'                                      ЮРм-С251ГПз, ЮРм-С251ГПзб'",
			expected: "ЮРм-С251ГПз, ЮРм-С251ГПзб",
		},
		{
			name:     "Название с двойными кавычками",
			input:    "\"Группа ИВТ-31\"",
			expected: "Группа ИВТ-31",
		},
		{
			name:     "Название без кавычек",
			input:    "Колледж ИП-22 (2024 ОФО МГТУ",
			expected: "Колледж ИП-22 (2024 ОФО МГТУ",
		},
		{
			name:     "Название с пробелами по краям",
			input:    "   Группа с пробелами   ",
			expected: "Группа с пробелами",
		},
		{
			name:     "Название с кавычками и пробелами",
			input:    "  '  Название  '  ",
			expected: "Название",
		},
		{
			name:     "Пустая строка",
			input:    "",
			expected: "",
		},
		{
			name:     "Только пробелы",
			input:    "     ",
			expected: "",
		},
		{
			name:     "Только кавычки",
			input:    "''",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanChatName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
