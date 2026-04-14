package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateInput_EnterMode(t *testing.T) {
	// arrange
	m := newTestModel(t)

	// act
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	// assert
	assert.Equal(t, modeInput, m2.mode)
	assert.NotNil(t, m2.editInputs, "editInputs が初期化されていること")
}

func TestUpdateInput_CreateTask(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	m2.editInputs[editFieldTitle].SetValue("テストタスク")

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
	tasks, err := m.taskRepo.FindAll()
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "テストタスク", tasks[0].Title)
}

func TestUpdateInput_ParsePriorityAndDue(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	m2.editInputs[editFieldTitle].SetValue("タスク名")
	m2.editInputs[editFieldPriority].SetValue("A")
	m2.editInputs[editFieldDueDate].SetValue("today")

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
	tasks, err := m.taskRepo.FindAll()
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "A", tasks[0].Priority)
	assert.NotNil(t, tasks[0].DueDate)
}

func TestUpdateInput_Cancel(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	m2.editInputs[editFieldTitle].SetValue("キャンセルするタスク")

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeList, m3.mode)
	tasks, err := m.taskRepo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, tasks, "タスクが作成されていないこと")
}

func TestUpdateInput_EmptyInput_Error(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeInput, m3.mode, "空入力でもmodeInputのまま")
	assert.NotNil(t, m3.lastErr)
}

func TestUpdateInput_TabFieldNavigation(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	assert.Equal(t, editFieldTitle, m2.editField, "初期フィールドはタイトル")

	// act + assert
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
	m3 := updated2.(Model)
	assert.Equal(t, editFieldPriority, m3.editField)

	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	m4 := updated3.(Model)
	assert.Equal(t, editFieldDueDate, m4.editField)

	updated4, _ := m4.Update(tea.KeyMsg{Type: tea.KeyTab})
	m5 := updated4.(Model)
	assert.Equal(t, editFieldTags, m5.editField)

	updated5, _ := m5.Update(tea.KeyMsg{Type: tea.KeyTab})
	m6 := updated5.(Model)
	assert.Equal(t, editFieldTitle, m6.editField)
}

func TestUpdateInput_ShiftTabNavigation(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	// act + assert
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m3 := updated2.(Model)
	assert.Equal(t, editFieldTitle, m3.editField, "先頭でクランプ")

	m2.editField = editFieldTags
	updated3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m4 := updated3.(Model)
	assert.Equal(t, editFieldDueDate, m4.editField)
}

func TestUpdateInput_TagSelectTransition(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	for i := 0; i < 3; i++ {
		tmp, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
		m2 = tmp.(Model)
	}
	assert.Equal(t, editFieldTags, m2.editField)

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeTagSelect, m3.mode)
	assert.Equal(t, "project", m3.pendingTagType)
}

func TestUpdateInput_ContextTagTransition(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	for i := 0; i < 3; i++ {
		tmp, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
		m2 = tmp.(Model)
	}

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("@")})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeTagSelect, m3.mode)
	assert.Equal(t, "context", m3.pendingTagType)
}

func TestUpdateInput_PlusOutsideTagsField_NoTransition(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	assert.Equal(t, editFieldTitle, m2.editField, "初期はタイトルフィールド")

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeInput, m3.mode)
}

func TestUpdateInput_DefaultKey(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeInput, m3.mode)
	assert.Equal(t, "a", m3.editInputs[editFieldTitle].Value())
}

func TestUpdateTagSelect_ExistingTag(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	for i := 0; i < 3; i++ {
		tmp, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
		m2 = tmp.(Model)
	}

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	m3 := updated2.(Model)
	assert.Contains(t, m3.tagList, *tag)

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, modeInput, m4.mode)
	assert.Len(t, m4.selectedTags, 1)
	assert.Equal(t, "work", m4.selectedTags[0].Name)
}

func TestUpdateTagSelect_NewTag(t *testing.T) {
	// arrange
	m := newTestModel(t)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	for i := 0; i < 3; i++ {
		tmp, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
		m2 = tmp.(Model)
	}

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	m3 := updated2.(Model)

	m3.tagInput = "newproject"
	m3.tagCursor = len(m3.filteredTags())

	// act
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := updated3.(Model)

	// assert
	assert.Equal(t, modeInput, m4.mode)
	assert.Len(t, m4.selectedTags, 1)
	assert.Equal(t, "newproject", m4.selectedTags[0].Name)
	tags, err := m.tagRepo.FindByType("project")
	require.NoError(t, err)
	assert.Len(t, tags, 1)
}

func TestUpdateTagSelect_FilterByType(t *testing.T) {
	// arrange
	m := newTestModel(t)
	_, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	_, err = m.tagRepo.FindOrCreate("office", "context")
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)

	for i := 0; i < 3; i++ {
		tmp, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
		m2 = tmp.(Model)
	}

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	m3 := updated2.(Model)

	// assert
	assert.Len(t, m3.tagList, 1)
	assert.Equal(t, "work", m3.tagList[0].Name)
}

func TestUpdateInput_NoDuplicateSelectedTags(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	m2.editInputs[editFieldTitle].SetValue("タスク名")
	m2.selectedTags = []db.Tag{*tag, *tag}

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated2.(Model)
	require.Equal(t, modeList, m3.mode)

	// assert
	tasks, err := m.taskRepo.FindAll()
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	task, err := m.taskRepo.FindByID(tasks[0].ID)
	require.NoError(t, err)
	assert.Len(t, task.Tags, 1, "selectedTags の重複タグは1件に絞られる")
	assert.Equal(t, "work", task.Tags[0].Name)
}

func TestUpdateInput_DiffNameDiffTypeBothSaved(t *testing.T) {
	// arrange
	m := newTestModel(t)
	projTag, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	ctxTag, err := m.tagRepo.FindOrCreate("office", "context")
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	m2.editInputs[editFieldTitle].SetValue("タスク名")
	m2.selectedTags = []db.Tag{*projTag, *ctxTag}

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated2.(Model)
	require.Equal(t, modeList, m3.mode)

	// assert
	tasks, err := m.taskRepo.FindAll()
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	task, err := m.taskRepo.FindByID(tasks[0].ID)
	require.NoError(t, err)
	assert.Len(t, task.Tags, 2, "プロジェクトタグとコンテキストタグは両方保存される")
}

func TestUpdateInput_TagCursorAndDelete(t *testing.T) {
	// arrange
	m := newTestModel(t)
	tag1, err := m.tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	tag2, err := m.tagRepo.FindOrCreate("home", "project")
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	m2.selectedTags = []db.Tag{*tag1, *tag2}

	for i := 0; i < 3; i++ {
		tmp, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
		m2 = tmp.(Model)
	}
	assert.Equal(t, editFieldTags, m2.editField)
	assert.Equal(t, 0, m2.editTagCursor)

	// act + assert
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m3 := updated2.(Model)
	assert.Equal(t, 1, m3.editTagCursor)

	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m4 := updated3.(Model)
	assert.Equal(t, 0, m4.editTagCursor)

	updated4, _ := m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m5 := updated4.(Model)
	assert.Len(t, m5.selectedTags, 1, "タグが1件削除される")
	assert.Equal(t, "home", m5.selectedTags[0].Name)
}

func TestUpdateInput_InvalidPriority(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	m2.editInputs[editFieldTitle].SetValue("タスク名")
	m2.editInputs[editFieldPriority].SetValue("1")

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeInput, m3.mode, "エラー時はmodeInputのまま")
	require.NotNil(t, m3.lastErr)
	assert.Contains(t, m3.lastErr.Error(), "優先度は A〜Z", "エラーメッセージに優先度の説明が含まれること")
}

func TestUpdateInput_InvalidDueDate(t *testing.T) {
	// arrange
	m := newTestModel(t)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m2 := updated.(Model)
	m2.editInputs[editFieldTitle].SetValue("タスク名")
	m2.editInputs[editFieldDueDate].SetValue("notadate")

	// act
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := updated2.(Model)

	// assert
	assert.Equal(t, modeInput, m3.mode, "エラー時はmodeInputのまま")
	require.NotNil(t, m3.lastErr)
	assert.Contains(t, m3.lastErr.Error(), "締切日の形式", "エラーメッセージに締切日の説明が含まれること")
}
