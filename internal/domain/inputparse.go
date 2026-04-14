package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// ParsedTask はタスク入力文字列のパース結果を保持する。
type ParsedTask struct {
	Title       string
	Priority    string    // "A"〜"Z"、なければ空文字列
	DueDate     *time.Time
	ProjectTags []string  // "+" なしのタグ名
	ContextTags []string  // "@" なしのタグ名
}

// ParseTaskInput はタスク入力文字列をパースして ParsedTask を返す。
// now は日付キーワードの解釈に使用する。
func ParseTaskInput(input string, now time.Time) (*ParsedTask, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("タイトルを入力してください")
	}

	tokens := tokenize(input)
	result := &ParsedTask{
		ProjectTags: []string{},
		ContextTags: []string{},
	}

	var titleTokens []string
	prioritySet := false

	for _, tok := range tokens {
		switch {
		case !prioritySet && isPriorityToken(tok):
			result.Priority = string([]rune(tok)[1]) // "(A)" → "A"
			prioritySet = true

		case strings.HasPrefix(tok, "+") && len(tok) > 1:
			result.ProjectTags = append(result.ProjectTags, tok[1:])

		case strings.HasPrefix(tok, "@") && len(tok) > 1:
			result.ContextTags = append(result.ContextTags, tok[1:])

		case strings.HasPrefix(tok, "due:") && len(tok) > 4:
			keyword := tok[4:]
			d, err := ParseDueDate(keyword, now)
			if err != nil {
				return nil, fmt.Errorf("締切日の解析エラー: %w", err)
			}
			result.DueDate = d

		default:
			titleTokens = append(titleTokens, tok)
		}
	}

	result.Title = strings.TrimSpace(strings.Join(titleTokens, " "))
	if result.Title == "" {
		return nil, fmt.Errorf("タイトルを入力してください")
	}

	return result, nil
}

// tokenize は入力文字列をスペース区切りでトークン分割する。
// 連続するスペースは1つにまとめる。
func tokenize(input string) []string {
	var tokens []string
	for _, tok := range strings.Fields(input) {
		if tok != "" {
			tokens = append(tokens, tok)
		}
	}
	return tokens
}

// isPriorityToken は "(A)"〜"(Z)" 形式かどうかを判定する。
func isPriorityToken(tok string) bool {
	runes := []rune(tok)
	return len(runes) == 3 &&
		runes[0] == '(' &&
		unicode.IsUpper(runes[1]) &&
		runes[2] == ')'
}
