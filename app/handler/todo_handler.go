package handler

import (
	"encoding/json"
	"fmt"
	"myapp/db/model"
	"myapp/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// TodoHandler TodoのHTTPハンドラー
type TodoHandler struct {
	todoService service.TodoService
}

// NewTodoHandler 新しいTodoハンドラーインスタンスを作成
func NewTodoHandler(todoService service.TodoService) *TodoHandler {
	return &TodoHandler{
		todoService: todoService,
	}
}

// APIエラーレスポンス用の構造体
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// APIサクセスレスポンス用の構造体
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Count   *int        `json:"count,omitempty"`
}

// エラーレスポンスを送信
func (h *TodoHandler) sendErrorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ErrorResponse{
		Error:   "API Error",
		Message: message,
		Code:    code,
	}

	json.NewEncoder(w).Encode(response)
}

// サクセスレスポンスを送信
func (h *TodoHandler) sendSuccessResponse(w http.ResponseWriter, data interface{}, message string, count *int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := SuccessResponse{
		Data:    data,
		Message: message,
		Count:   count,
	}

	json.NewEncoder(w).Encode(response)
}

// GetAllTodos GET /todos - 全てのTodoを取得
func (h *TodoHandler) GetAllTodos(w http.ResponseWriter, r *http.Request) {
	// クエリパラメータの解析
	query := r.URL.Query()

	var todos []*model.Todo
	var err error

	// フィルタリング処理
	if priority := query.Get("priority"); priority != "" {
		priorityEnum := model.Priority(priority)
		todos, err = h.todoService.GetTodosByPriority(priorityEnum)
	} else if completed := query.Get("completed"); completed != "" {
		if completed == "true" {
			todos, err = h.todoService.GetCompletedTodos()
		} else if completed == "false" {
			todos, err = h.todoService.GetPendingTodos()
		} else {
			h.sendErrorResponse(w, "completedパラメータはtrueまたはfalseである必要があります", http.StatusBadRequest)
			return
		}
	} else {
		todos, err = h.todoService.GetAllTodos()
	}

	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TodoResponseに変換
	responses := make([]*model.TodoResponse, len(todos))
	for i, todo := range todos {
		responses[i] = todo.ToResponse()
	}

	count := len(responses)
	h.sendSuccessResponse(w, responses, "Todoリストを取得しました", &count)
}

// GetTodoByID GET /todos/{id} - 特定のTodoを取得
func (h *TodoHandler) GetTodoByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		h.sendErrorResponse(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendErrorResponse(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	todo, err := h.todoService.GetTodoByID(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		} else {
			h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	h.sendSuccessResponse(w, todo.ToResponse(), "Todoを取得しました", nil)
}

// CreateTodo POST /todos - 新しいTodoを作成
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req model.TodoCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "無効なJSONフォーマットです", http.StatusBadRequest)
		return
	}

	// バリデーション
	if strings.TrimSpace(req.Title) == "" {
		h.sendErrorResponse(w, "タイトルは必須です", http.StatusBadRequest)
		return
	}

	todo, err := h.todoService.CreateTodo(&req)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	h.sendSuccessResponse(w, todo.ToResponse(), "Todoを作成しました", nil)
}

// UpdateTodo PUT /todos/{id} - 既存のTodoを更新
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		h.sendErrorResponse(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendErrorResponse(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	var req model.TodoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "無効なJSONフォーマットです", http.StatusBadRequest)
		return
	}

	todo, err := h.todoService.UpdateTodo(uint(id), &req)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		} else {
			h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	h.sendSuccessResponse(w, todo.ToResponse(), "Todoを更新しました", nil)
}

// DeleteTodo DELETE /todos/{id} - Todoを削除
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists {
		h.sendErrorResponse(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendErrorResponse(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	err = h.todoService.DeleteTodo(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			h.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		} else {
			h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	h.sendSuccessResponse(w, nil, fmt.Sprintf("ID %d のTodoを削除しました", id), nil)
}
