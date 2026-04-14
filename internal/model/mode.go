package model

type mode int

const (
	modeList          mode = iota // タスク一覧（デフォルト）
	modeInput                     // タスク入力
	modeTagSelect                 // タグ選択（入力モード中のサブモード）
	modeEditDetail                // 詳細編集
	modeDeleteConfirm             // 削除確認モーダル
	modeTrackingAlert             // タイムトラッキング衝突アラート
	modeStats                     // 統計・異常セッションビュー
	modeTimeline                  // 1日のタイムライン
)
