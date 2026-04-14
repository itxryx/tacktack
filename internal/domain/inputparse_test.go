package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTaskInput(t *testing.T) {
	fullNow := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		input           string
		now             time.Time
		wantErr         bool
		wantTitle       string
		wantPriority    string
		wantProjectTags []string
		wantContextTags []string
		wantDueNil      bool
	}{
		{
			name:            "フル入力",
			input:           "(A) メールを確認 +work @office due:tomorrow",
			now:             fullNow,
			wantTitle:       "メールを確認",
			wantPriority:    "A",
			wantProjectTags: []string{"work"},
			wantContextTags: []string{"office"},
			wantDueNil:      false,
		},
		{
			name:            "優先度なし",
			input:           "タスク名 +work",
			now:             time.Now(),
			wantTitle:       "タスク名",
			wantPriority:    "",
			wantProjectTags: []string{"work"},
			wantContextTags: []string{},
			wantDueNil:      true,
		},
		{
			name:            "タグなし",
			input:           "(B) タスク名 due:today",
			now:             time.Now(),
			wantTitle:       "タスク名",
			wantPriority:    "B",
			wantProjectTags: []string{},
			wantContextTags: []string{},
			wantDueNil:      false,
		},
		{
			name:            "日本語タイトル",
			input:           "資料を作成する +work",
			now:             time.Now(),
			wantTitle:       "資料を作成する",
			wantPriority:    "",
			wantProjectTags: []string{"work"},
			wantContextTags: []string{},
			wantDueNil:      true,
		},
		{
			name:    "空入力",
			input:   "",
			now:     time.Now(),
			wantErr: true,
		},
		{
			name:    "タグのみ",
			input:   "+work @office",
			now:     time.Now(),
			wantErr: true,
		},
		{
			name:            "複数タグ",
			input:           "タスク +work +personal @office @home",
			now:             time.Now(),
			wantTitle:       "タスク",
			wantProjectTags: []string{"work", "personal"},
			wantContextTags: []string{"office", "home"},
			wantDueNil:      true,
		},
		{
			name:            "締切なし",
			input:           "タスク",
			now:             time.Now(),
			wantTitle:       "タスク",
			wantProjectTags: []string{},
			wantContextTags: []string{},
			wantDueNil:      true,
		},
		{
			name:    "不正締切",
			input:   "タスク due:invalid",
			now:     time.Now(),
			wantErr: true,
		},
		{
			name:            "小文字優先度",
			input:           "(a) タスク名",
			now:             time.Now(),
			wantTitle:       "(a) タスク名",
			wantPriority:    "",
			wantProjectTags: []string{},
			wantContextTags: []string{},
			wantDueNil:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// act
			result, err := ParseTaskInput(tc.input, tc.now)
			if tc.wantErr {
				// assert
				assert.Error(t, err)
				return
			}

			// assert
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.wantTitle, result.Title)
			assert.Equal(t, tc.wantPriority, result.Priority)
			assert.Equal(t, tc.wantProjectTags, result.ProjectTags)
			assert.Equal(t, tc.wantContextTags, result.ContextTags)
			if tc.wantDueNil {
				assert.Nil(t, result.DueDate)
			} else {
				assert.NotNil(t, result.DueDate)
			}
		})
	}
}
