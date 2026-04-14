package model

import (
	"fmt"
	"strings"
)

func (m Model) viewTagSelect() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%sを選択してください（Esc: 戻る）\n", tagTypeLabel(m.pendingTagType))
	fmt.Fprintf(&sb, "フィルタ: [%s]\n\n", m.tagInput)

	// 削除確認中の表示
	if m.tagDeleteTarget != nil {
		fmt.Fprintf(&sb, "%s「%s」を全タスクから削除しますか？ (y: 削除 / n,Esc: キャンセル)\n",
			tagTypeLabel(m.tagDeleteTarget.Type), m.tagDeleteTarget.Name)
		return sb.String()
	}

	prefix := "+"
	if m.pendingTagType == "context" {
		prefix = "@"
	}

	filtered := m.filteredTags()
	for i, tag := range filtered {
		marker := "○"
		if i == m.tagCursor {
			marker = "●"
			fmt.Fprintf(&sb, "▶ %s %s%s\n", marker, prefix, tag.Name)
		} else {
			fmt.Fprintf(&sb, "  %s %s%s\n", marker, prefix, tag.Name)
		}
	}

	// 新規作成エントリ
	newIdx := len(filtered)
	newLabel := fmt.Sprintf("新規%sを入力してください", tagTypeLabel(m.pendingTagType))
	if m.tagInput != "" {
		newLabel = fmt.Sprintf("新規作成: %s", m.tagInput)
	}
	if m.tagCursor == newIdx {
		fmt.Fprintf(&sb, "▶ ● %s %s\n", newLabel, prefix)
	} else {
		fmt.Fprintf(&sb, "  ○ %s %s\n", newLabel, prefix)
	}

	return sb.String()
}
