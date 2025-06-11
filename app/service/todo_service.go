package service

import (
	"fmt"
	"myapp/db"
	"myapp/db/model"

	"gorm.io/gorm"
)

// TodoService Todoサービスのインターフェース
type TodoService interface {
	GetAllTodos() ([]*model.Todo, error)
	GetTodoByID(id uint) (*model.Todo, error)
	CreateTodo(req *model.TodoCreateRequest) (*model.Todo, error)
	UpdateTodo(id uint, req *model.TodoUpdateRequest) (*model.Todo, error)
	DeleteTodo(id uint) error
	GetTodosByPriority(priority model.Priority) ([]*model.Todo, error)
	GetCompletedTodos() ([]*model.Todo, error)
	GetPendingTodos() ([]*model.Todo, error)
}

// todoService Todoサービスの実装
type todoService struct {
	db *gorm.DB
}

// NewTodoService 新しいTodoサービスインスタンスを作成
func NewTodoService() TodoService {
	return &todoService{
		db: db.GetDB(),
	}
}

// GetAllTodos 全てのTodoを取得
func (s *todoService) GetAllTodos() ([]*model.Todo, error) {
	var todos []*model.Todo

	result := s.db.Order("created_at DESC").Find(&todos)
	if result.Error != nil {
		return nil, fmt.Errorf("Todoの取得に失敗しました: %w", result.Error)
	}

	return todos, nil
}

// GetTodoByID IDで特定のTodoを取得
func (s *todoService) GetTodoByID(id uint) (*model.Todo, error) {
	var todo model.Todo

	result := s.db.First(&todo, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ID %d のTodoが見つかりません", id)
		}
		return nil, fmt.Errorf("Todoの取得に失敗しました: %w", result.Error)
	}

	return &todo, nil
}

// CreateTodo 新しいTodoを作成
func (s *todoService) CreateTodo(req *model.TodoCreateRequest) (*model.Todo, error) {
	// 優先度の検証
	if req.Priority != "" && !req.Priority.IsValid() {
		return nil, fmt.Errorf("無効な優先度です: %s", req.Priority)
	}

	// デフォルト優先度の設定
	if req.Priority == "" {
		req.Priority = model.PriorityMedium
	}

	todo := &model.Todo{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		Completed:   false,
	}

	result := s.db.Create(todo)
	if result.Error != nil {
		return nil, fmt.Errorf("Todoの作成に失敗しました: %w", result.Error)
	}

	return todo, nil
}

// UpdateTodo 既存のTodoを更新
func (s *todoService) UpdateTodo(id uint, req *model.TodoUpdateRequest) (*model.Todo, error) {
	// 既存のTodoを取得
	todo, err := s.GetTodoByID(id)
	if err != nil {
		return nil, err
	}

	// 更新フィールドの適用
	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}
	if req.Priority != nil {
		if !req.Priority.IsValid() {
			return nil, fmt.Errorf("無効な優先度です: %s", *req.Priority)
		}
		todo.Priority = *req.Priority
	}
	if req.DueDate != nil {
		todo.DueDate = req.DueDate
	}

	result := s.db.Save(todo)
	if result.Error != nil {
		return nil, fmt.Errorf("Todoの更新に失敗しました: %w", result.Error)
	}

	return todo, nil
}

// DeleteTodo Todoを削除（ソフトデリート）
func (s *todoService) DeleteTodo(id uint) error {
	// 存在確認
	_, err := s.GetTodoByID(id)
	if err != nil {
		return err
	}

	result := s.db.Delete(&model.Todo{}, id)
	if result.Error != nil {
		return fmt.Errorf("Todoの削除に失敗しました: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("ID %d のTodoは既に削除されています", id)
	}

	return nil
}

// GetTodosByPriority 優先度でTodoをフィルタリング
func (s *todoService) GetTodosByPriority(priority model.Priority) ([]*model.Todo, error) {
	if !priority.IsValid() {
		return nil, fmt.Errorf("無効な優先度です: %s", priority)
	}

	var todos []*model.Todo

	result := s.db.Where("priority = ?", priority).Order("created_at DESC").Find(&todos)
	if result.Error != nil {
		return nil, fmt.Errorf("優先度 %s のTodo取得に失敗しました: %w", priority, result.Error)
	}

	return todos, nil
}

// GetCompletedTodos 完了済みTodoを取得
func (s *todoService) GetCompletedTodos() ([]*model.Todo, error) {
	var todos []*model.Todo

	result := s.db.Where("completed = ?", true).Order("updated_at DESC").Find(&todos)
	if result.Error != nil {
		return nil, fmt.Errorf("完了済みTodoの取得に失敗しました: %w", result.Error)
	}

	return todos, nil
}

// GetPendingTodos 未完了Todoを取得
func (s *todoService) GetPendingTodos() ([]*model.Todo, error) {
	var todos []*model.Todo

	result := s.db.Where("completed = ?", false).Order("priority DESC, created_at DESC").Find(&todos)
	if result.Error != nil {
		return nil, fmt.Errorf("未完了Todoの取得に失敗しました: %w", result.Error)
	}

	return todos, nil
}
