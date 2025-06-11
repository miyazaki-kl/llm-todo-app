package handler

import (
	"context"
	"fmt"
	"myapp/db/model"
	"myapp/service"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// Huma用のレスポンス構造体

// TodoListResponse Todoリスト取得のレスポンス
type TodoListResponse struct {
	Body struct {
		Data    []*model.TodoResponse `json:"data" doc:"Todoアイテムのリスト"`
		Message string                `json:"message" doc:"レスポンスメッセージ"`
		Count   int                   `json:"count" doc:"Todoアイテムの総数"`
	}
}

// TodoResponse 単一Todo取得のレスポンス
type TodoResponse struct {
	Body struct {
		Data    *model.TodoResponse `json:"data" doc:"Todoアイテム"`
		Message string              `json:"message" doc:"レスポンスメッセージ"`
	}
}

// TodoCreateRequest Todo作成リクエスト
type TodoCreateRequest struct {
	Body model.TodoCreateRequest `doc:"作成するTodoの情報"`
}

// TodoUpdateRequest Todo更新リクエスト
type TodoUpdateRequest struct {
	ID   int                     `path:"id" doc:"更新するTodoのID" minimum:"1"`
	Body model.TodoUpdateRequest `doc:"更新するTodoの情報"`
}

// TodoIDRequest ID指定リクエスト
type TodoIDRequest struct {
	ID int `path:"id" doc:"TodoのID" minimum:"1"`
}

// TodoQueryRequest クエリパラメータ付きリクエスト
type TodoQueryRequest struct {
	Priority  string `query:"priority" enum:"low,medium,high,urgent" doc:"優先度でフィルタリング"`
	Completed string `query:"completed" doc:"完了状態でフィルタリング"`
}

// DeleteResponse 削除レスポンス
type DeleteResponse struct {
	Body struct {
		Message string `json:"message" doc:"削除結果のメッセージ"`
	}
}

// HumaErrorResponse エラーレスポンス
type HumaErrorResponse struct {
	Body struct {
		Error   string `json:"error" doc:"エラータイプ"`
		Message string `json:"message" doc:"エラーメッセージ"`
		Code    int    `json:"code" doc:"HTTPステータスコード"`
	}
}

// HealthResponse ヘルスチェックレスポンス
type HealthResponse struct {
	Body struct {
		Message   string    `json:"message" doc:"ヘルスチェック結果"`
		Timestamp time.Time `json:"timestamp" doc:"チェック実行時刻"`
		Status    string    `json:"status" doc:"ステータス"`
	}
}

// HumaTodoHandler Huma用のTodoハンドラー
type HumaTodoHandler struct {
	todoService service.TodoService
}

// NewHumaTodoHandler 新しいHumaTodoハンドラーインスタンスを作成
func NewHumaTodoHandler(todoService service.TodoService) *HumaTodoHandler {
	return &HumaTodoHandler{
		todoService: todoService,
	}
}

// GetAllTodos 全てのTodoを取得
func (h *HumaTodoHandler) GetAllTodos(ctx context.Context, input *TodoQueryRequest) (*TodoListResponse, error) {
	var todos []*model.Todo
	var err error

	// フィルタリング処理
	if input.Priority != "" {
		priority := model.Priority(input.Priority)
		todos, err = h.todoService.GetTodosByPriority(priority)
	} else if input.Completed != "" {
		if input.Completed == "true" {
			todos, err = h.todoService.GetCompletedTodos()
		} else if input.Completed == "false" {
			todos, err = h.todoService.GetPendingTodos()
		} else {
			todos, err = h.todoService.GetAllTodos()
		}
	} else {
		todos, err = h.todoService.GetAllTodos()
	}

	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// TodoResponseに変換
	responses := make([]*model.TodoResponse, len(todos))
	for i, todo := range todos {
		responses[i] = todo.ToResponse()
	}

	return &TodoListResponse{
		Body: struct {
			Data    []*model.TodoResponse `json:"data" doc:"Todoアイテムのリスト"`
			Message string                `json:"message" doc:"レスポンスメッセージ"`
			Count   int                   `json:"count" doc:"Todoアイテムの総数"`
		}{
			Data:    responses,
			Message: "Todoリストを取得しました",
			Count:   len(responses),
		},
	}, nil
}

// GetTodoByID 特定のTodoを取得
func (h *HumaTodoHandler) GetTodoByID(ctx context.Context, input *TodoIDRequest) (*TodoResponse, error) {
	todo, err := h.todoService.GetTodoByID(uint(input.ID))
	if err != nil {
		if err.Error() == fmt.Sprintf("ID %d のTodoが見つかりません", input.ID) {
			return nil, huma.Error404NotFound(err.Error())
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}

	return &TodoResponse{
		Body: struct {
			Data    *model.TodoResponse `json:"data" doc:"Todoアイテム"`
			Message string              `json:"message" doc:"レスポンスメッセージ"`
		}{
			Data:    todo.ToResponse(),
			Message: "Todoを取得しました",
		},
	}, nil
}

// CreateTodo 新しいTodoを作成
func (h *HumaTodoHandler) CreateTodo(ctx context.Context, input *TodoCreateRequest) (*TodoResponse, error) {
	todo, err := h.todoService.CreateTodo(&input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}

	resp := &TodoResponse{
		Body: struct {
			Data    *model.TodoResponse `json:"data" doc:"Todoアイテム"`
			Message string              `json:"message" doc:"レスポンスメッセージ"`
		}{
			Data:    todo.ToResponse(),
			Message: "Todoを作成しました",
		},
	}

	return resp, nil
}

// UpdateTodo 既存のTodoを更新
func (h *HumaTodoHandler) UpdateTodo(ctx context.Context, input *TodoUpdateRequest) (*TodoResponse, error) {
	todo, err := h.todoService.UpdateTodo(uint(input.ID), &input.Body)
	if err != nil {
		if err.Error() == fmt.Sprintf("ID %d のTodoが見つかりません", input.ID) {
			return nil, huma.Error404NotFound(err.Error())
		}
		return nil, huma.Error400BadRequest(err.Error())
	}

	return &TodoResponse{
		Body: struct {
			Data    *model.TodoResponse `json:"data" doc:"Todoアイテム"`
			Message string              `json:"message" doc:"レスポンスメッセージ"`
		}{
			Data:    todo.ToResponse(),
			Message: "Todoを更新しました",
		},
	}, nil
}

// DeleteTodo Todoを削除
func (h *HumaTodoHandler) DeleteTodo(ctx context.Context, input *TodoIDRequest) (*DeleteResponse, error) {
	err := h.todoService.DeleteTodo(uint(input.ID))
	if err != nil {
		if err.Error() == fmt.Sprintf("ID %d のTodoが見つかりません", input.ID) {
			return nil, huma.Error404NotFound(err.Error())
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}

	return &DeleteResponse{
		Body: struct {
			Message string `json:"message" doc:"削除結果のメッセージ"`
		}{
			Message: fmt.Sprintf("ID %d のTodoを削除しました", input.ID),
		},
	}, nil
}
