package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func enterTagSelectFromList(t *testing.T, m Model) Model {
	t.Helper()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	for i := 0; i < 3; i++ {
		tmp, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
		m2 = tmp.(Model)
	}
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	return updated2.(Model)
}

func TestDeleteTag_RemovesFromDB(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	m2 := enterTagSelectFromList(t, m)
	require.Len(t, m2.tagList, 1)

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	m3 := updated.(Model)

	// assert
	require.NotNil(t, m3.tagDeleteTarget, "削除確認状態になっていること")
	assert.Len(t, m3.tagList, 1, "確認前はまだ削除されていないこと")

	// act
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m4 := updated2.(Model)

	// assert
	assert.Nil(t, m4.lastErr)
	assert.Empty(t, m4.tagList, "tagList から削除されていること")
	all, err := m.tagRepo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, all)
	_ = tag
}

func TestDeleteTag_RemovesFromSelectedTags(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	m.selectedTags = append(m.selectedTags, *tag)
	m2 := enterTagSelectFromList(t, m)
	m2.selectedTags = m.selectedTags
	require.Len(t, m2.tagList, 1)

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	m3 := updated.(Model)
	require.NotNil(t, m3.tagDeleteTarget)
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m4 := updated2.(Model)

	// assert
	assert.Empty(t, m4.selectedTags, "selectedTags からも削除されていること")
}

func TestDeleteTag_CursorClamped(t *testing.T) {
	// arrange
	m := newTestModel(t)
	_, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	_, err = m.tagRepo.FindOrCreate("home", "project")
	require.NoError(t, err)
	m2 := enterTagSelectFromList(t, m)
	require.Len(t, m2.tagList, 2)

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, 1, m3.tagCursor)

	// act
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	m4 := updated2.(Model)
	require.NotNil(t, m4.tagDeleteTarget)
	updated3, _ := m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m5 := updated3.(Model)

	// assert
	assert.Len(t, m5.tagList, 1)
	assert.Equal(t, 0, m5.tagCursor, "カーソルがクランプされること")
}

func TestDeleteTag_NewEntryNotDeletable(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2 := enterTagSelectFromList(t, m)
	require.Empty(t, m2.tagList)

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	m3 := updated.(Model)

	// assert
	assert.Nil(t, m3.lastErr)
	assert.Nil(t, m3.tagDeleteTarget, "新規作成エントリは削除確認にならないこと")
}

func TestDeleteTag_CancelConfirmation(t *testing.T) {
	// arrange
	m := newTestModel(t)
	_, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	m2 := enterTagSelectFromList(t, m)
	require.Len(t, m2.tagList, 1)
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	m3 := updated.(Model)
	require.NotNil(t, m3.tagDeleteTarget, "削除確認状態になっていること")

	// act
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m4 := updated2.(Model)

	// assert
	assert.Nil(t, m4.tagDeleteTarget, "キャンセル後は確認状態が解除されること")
	assert.Len(t, m4.tagList, 1, "タグは削除されていないこと")
	all, err := m.tagRepo.FindAll()
	require.NoError(t, err)
	assert.Len(t, all, 1, "キャンセル後もDBにタグが残っていること")
}

func TestUpdateTagSelect_CursorMove(t *testing.T) {
	// arrange
	m := newTestModel(t)
	_, err := m.tagRepo.FindOrCreate("aaa", "project")
	require.NoError(t, err)
	_, err = m.tagRepo.FindOrCreate("bbb", "project")
	require.NoError(t, err)
	m2 := enterTagSelectFromList(t, m)
	require.Len(t, m2.tagList, 2)
	assert.Equal(t, 0, m2.tagCursor)

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, 1, m3.tagCursor)

	// act
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyUp})
	m4 := updated2.(Model)

	// assert
	assert.Equal(t, 0, m4.tagCursor)

	// act
	updated3, _ := m4.Update(tea.KeyMsg{Type: tea.KeyUp})
	m5 := updated3.(Model)

	// assert
	assert.Equal(t, 0, m5.tagCursor, "先頭でクランプ")
}

func TestUpdateTagSelect_CursorBottomClamp(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2 := enterTagSelectFromList(t, m)
	require.Empty(t, m2.tagList)

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, 0, m3.tagCursor, "末尾でクランプ")
}

func TestUpdateTagSelect_Backspace(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2 := enterTagSelectFromList(t, m)
	m2.tagInput = "abc"
	m2.tagCursor = 1

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, "ab", m3.tagInput)
	assert.Equal(t, 0, m3.tagCursor, "backspace 後にカーソルリセット")
}

func TestUpdateTagSelect_Backspace_EmptyInput(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2 := enterTagSelectFromList(t, m)
	m2.tagInput = ""
	m2.tagCursor = 1

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, "", m3.tagInput)
	assert.Equal(t, 1, m3.tagCursor, "空入力時はカーソル変化なし")
}

func TestUpdateTagSelect_Esc(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2 := enterTagSelectFromList(t, m)
	m2.tagInput = "work"
	m2.tagCursor = 1

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeInput, m3.mode, "Esc で prevMode に戻る")
	assert.Equal(t, "", m3.tagInput, "tagInput がクリアされる")
	assert.Equal(t, 0, m3.tagCursor)
}

func TestUpdateTagSelect_TypedFilter(t *testing.T) {
	// arrange
	m := newTestModel(t)
	_, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	_, err = m.tagRepo.FindOrCreate("home", "project")
	require.NoError(t, err)
	m2 := enterTagSelectFromList(t, m)
	require.Len(t, m2.tagList, 2)

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, "w", m3.tagInput)
	assert.Equal(t, 0, m3.tagCursor, "フィルタ後にカーソルリセット")
	filtered := m3.filteredTags()
	assert.Len(t, filtered, 1)
	assert.Equal(t, "work", filtered[0].Name)
}

func TestUpdateTagSelect_JKTypedAsFilter(t *testing.T) {
	// arrange
	m := newTestModel(t)
	_, err := m.tagRepo.FindOrCreate("jira", "project")
	require.NoError(t, err)
	_, err = m.tagRepo.FindOrCreate("kanban", "project")
	require.NoError(t, err)
	m2 := enterTagSelectFromList(t, m)
	require.Len(t, m2.tagList, 2)

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, "j", m3.tagInput, "j は文字として tagInput に追加される")
	assert.Equal(t, 0, m3.tagCursor, "フィルタ後にカーソルリセット")
	filtered := m3.filteredTags()
	assert.Len(t, filtered, 1)
	assert.Equal(t, "jira", filtered[0].Name)

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m4 := updated2.(Model)

	// assert
	assert.Equal(t, "k", m4.tagInput, "k は文字として tagInput に追加される")
	filtered2 := m4.filteredTags()
	assert.Len(t, filtered2, 1)
	assert.Equal(t, "kanban", filtered2[0].Name)
}

func TestUpdateTagSelect_FilterNoMatch(t *testing.T) {
	// arrange
	m := newTestModel(t)
	_, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	m2 := enterTagSelectFromList(t, m)
	m2.tagInput = "xyz"

	// act + assert
	assert.Empty(t, m2.filteredTags(), "マッチなしは空スライス")
}

func TestUpdateTagSelect_NoDuplicateOnDoubleSelect(t *testing.T) {
	// arrange
	m := newTestModel(t)
	_, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	m2 := enterTagSelectFromList(t, m)
	require.Len(t, m2.tagList, 1)

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated.(Model)

	// assert
	assert.Len(t, m3.selectedTags, 1)

	// arrange
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	m4 := updated2.(Model)

	// act
	updated3, _ := m4.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m5 := updated3.(Model)

	// assert
	assert.Len(t, m5.selectedTags, 1, "同じタグを2回選択しても重複しない")
}

func TestUpdateTagSelect_NoDuplicateOnDoubleCreateNewTag(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2 := enterTagSelectFromList(t, m)
	m2.tagInput = "newtag"
	m2.tagCursor = 0

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated.(Model)

	// assert
	assert.Len(t, m3.selectedTags, 1)

	// arrange
	updated2, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	m4 := updated2.(Model)
	m4.tagInput = "newtag"
	m4.tagCursor = len(m4.filteredTags())

	// act
	updated3, _ := m4.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m5 := updated3.(Model)

	// assert
	assert.Len(t, m5.selectedTags, 1, "同名の新規タグを2回作成しても重複しない")
}

func TestUpdateTagSelect_SelectNewTag_EmptyInput(t *testing.T) {
	// arrange
	m := newTestModel(t)
	m2 := enterTagSelectFromList(t, m)
	m2.tagInput = ""
	m2.tagCursor = 0

	// act
	updated, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated.(Model)

	// assert
	assert.Equal(t, modeInput, m3.mode, "入力なしの Enter → prevMode に戻る")
	assert.Empty(t, m3.selectedTags, "タグは追加されない")
}
