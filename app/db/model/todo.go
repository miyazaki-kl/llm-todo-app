package model

import (
	"time"

	"gorm.io/gorm"
)

// Todo Todoアイテムのモデル
type Todo struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"not null;size:255" validate:"required,max=255"`
	Description string         `json:"description" gorm:"type:text"`
	Completed   bool           `json:"completed" gorm:"default:false"`
	Priority    Priority       `json:"priority" gorm:"type:varchar(10);default:'medium'"`
	DueDate     *time.Time     `json:"due_date,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Priority 優先度の列挙型
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// IsValid 優先度が有効かチェック
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

// String 優先度を文字列で返す
func (p Priority) String() string {
	return string(p)
}

// TodoCreateRequest Todo作成リクエスト用の構造体
type TodoCreateRequest struct {
	Title       string     `json:"title" validate:"required,max=255"`
	Description string     `json:"description"`
	Priority    Priority   `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// TodoUpdateRequest Todo更新リクエスト用の構造体
type TodoUpdateRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,max=255"`
	Description *string    `json:"description,omitempty"`
	Completed   *bool      `json:"completed,omitempty"`
	Priority    *Priority  `json:"priority,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// TodoResponse APIレスポンス用のTodo構造体
type TodoResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	Priority    Priority   `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse TodoモデルをTodoResponseに変換
func (t *Todo) ToResponse() *TodoResponse {
	return &TodoResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		Priority:    t.Priority,
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// TableName テーブル名を指定
func (Todo) TableName() string {
	return "todos"
}
