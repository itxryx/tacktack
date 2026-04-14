package model

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/itxryx/tacktack/internal/repository"
	"gorm.io/gorm"
)

// Model は Bubble Tea アプリケーションの全状態を保持する。
type Model struct {
	// 画面モード
	mode mode

	// データ
	tasks     []db.Task
	cursor    int          // 一覧カーソル位置
	activeLog *db.TimeLog  // 現在計測中のセッション（nil = 計測なし）
	nullLogs  []db.TimeLog // 起動時に検出した end_at=NULL のセッション（異常終了）

	// リポジトリ
	taskRepo    repository.TaskRepository
	tagRepo     repository.TagRepository
	timeLogRepo repository.TimeLogRepository

	// タスク削除確認用
	deleteTargetID uint

	// タスク入力モード用
	selectedTags []db.Tag

	// タグ選択モード用
	pendingTagType    string   // "project" or "context"
	tagList           []db.Tag // 選択肢として表示するタグ一覧
	tagCursor         int
	tagInput          string
	tagSelectPrevMode mode    // タグ選択完了後に戻るモード
	tagDeleteTarget   *db.Tag // 削除確認中のタグ（nil = 確認中ではない）

	// 詳細編集モード用
	editTask         *db.Task
	editField        int
	editInputs       []textinput.Model
	editTagCursor    int // editFieldTags でのタグリストカーソル
	editLogCursor    int
	editLogFocus     bool
	editLogInputs    []textinput.Model
	editLogEditing   int
	editPrevMode     mode      // M2: 編集モード遷移元のモード（Esc で戻る先）
	editStoppedLog   *db.TimeLog // 完了トグル時に遅延 Stop するログ（保存時に Stop を実行）

	// 衝突アラート用
	conflictTaskID uint

	// 統計ビュー用
	statsCursor        int       // 異常セッション一覧のカーソル
	statsTagPeriodIdx  int       // タグ別統計の期間インデックス（0=Day〜4=Year）
	statsTasks         []db.Task // 統計用全タスクデータ（期間制限なし、初回遷移時に取得）

	// タイムラインビュー用
	timelineDate   time.Time // 表示対象の日付（ゼロ値なら未初期化）
	timelineScroll int       // 縦スクロール位置（スロット単位）

	// エラー表示
	lastErr error

	// ターミナルサイズ
	width  int
	height int
}

// New は GORM DB からリポジトリを初期化して Model を作成する。
func New(database *gorm.DB) Model {
	return Model{
		mode:        modeList,
		taskRepo:    repository.NewTaskRepository(database),
		tagRepo:     repository.NewTagRepository(database),
		timeLogRepo: repository.NewTimeLogRepository(database),
	}
}

// Init は起動時の初期化処理を行う。
// タスク一覧取得・アクティブセッション確認・異常セッション検出を非同期で実行する。
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		from := time.Now().AddDate(0, -1, 0)
		tasks, err := m.taskRepo.FindAll(repository.WithRecentCompleted(from))
		if err != nil {
			return initDoneMsg{err: err}
		}
		active, err := m.timeLogRepo.FindActive()
		if err != nil {
			return initDoneMsg{err: err}
		}
		nulls, err := m.timeLogRepo.FindNullEndAt()
		if err != nil {
			return initDoneMsg{err: err}
		}

		// アクティブセッションがある場合は nullLogs から除外
		var filteredNulls []db.TimeLog
		for _, log := range nulls {
			if active == nil || log.ID != active.ID {
				filteredNulls = append(filteredNulls, log)
			}
		}

		return initDoneMsg{tasks: tasks, activeLog: active, nullLogs: filteredNulls}
	}
}
