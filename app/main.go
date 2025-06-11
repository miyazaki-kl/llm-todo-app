package main

import (
	"context"
	"fmt"
	"log"
	"myapp/db"
	"myapp/handler"
	"myapp/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ヘルスチェック用のレスポンス構造体
type HealthCheckResponse struct {
	Body struct {
		Message   string    `json:"message" doc:"ヘルスチェック結果"`
		Timestamp time.Time `json:"timestamp" doc:"チェック実行時刻"`
		Status    string    `json:"status" doc:"ステータス"`
	}
}

// ヘルスチェック用のハンドラー
func healthHandler(ctx context.Context, input *struct{}) (*HealthCheckResponse, error) {
	return &HealthCheckResponse{
		Body: struct {
			Message   string    `json:"message" doc:"ヘルスチェック結果"`
			Timestamp time.Time `json:"timestamp" doc:"チェック実行時刻"`
			Status    string    `json:"status" doc:"ステータス"`
		}{
			Message:   "Goサーバーが正常に動作しています",
			Timestamp: time.Now(),
			Status:    "healthy",
		},
	}, nil
}

// ホームページ用のハンドラー
func homeHandler(ctx context.Context, input *struct{}) (*HealthCheckResponse, error) {
	return &HealthCheckResponse{
		Body: struct {
			Message   string    `json:"message" doc:"ヘルスチェック結果"`
			Timestamp time.Time `json:"timestamp" doc:"チェック実行時刻"`
			Status    string    `json:"status" doc:"ステータス"`
		}{
			Message:   "Todo API サーバーへようこそ！",
			Timestamp: time.Now(),
			Status:    "success",
		},
	}, nil
}

// データベース接続状態チェック用のハンドラー
func dbHealthHandler(ctx context.Context, input *struct{}) (*HealthCheckResponse, error) {
	database := db.GetDB()

	if database == nil {
		return nil, huma.Error503ServiceUnavailable("データベース接続が初期化されていません")
	}

	sqlDB, err := database.DB()
	if err != nil || sqlDB.Ping() != nil {
		return nil, huma.Error503ServiceUnavailable("データベース接続に問題があります")
	}

	return &HealthCheckResponse{
		Body: struct {
			Message   string    `json:"message" doc:"ヘルスチェック結果"`
			Timestamp time.Time `json:"timestamp" doc:"チェック実行時刻"`
			Status    string    `json:"status" doc:"ステータス"`
		}{
			Message:   "データベース接続は正常です",
			Timestamp: time.Now(),
			Status:    "healthy",
		},
	}, nil
}

func main() {
	// データベース接続
	log.Println("データベースに接続中...")
	if err := db.Connect(); err != nil {
		log.Fatalf("データベース接続エラー: %v", err)
	}

	// マイグレーション実行
	log.Println("データベースマイグレーション実行中...")
	if err := db.Migrate(); err != nil {
		log.Fatalf("マイグレーションエラー: %v", err)
	}

	// サービスとハンドラーの初期化
	todoService := service.NewTodoService()
	todoHandler := handler.NewHumaTodoHandler(todoService)

	// Chi routerの設定
	router := chi.NewRouter()

	// ミドルウェアの追加
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// CORSの設定
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// HumaのAPIインスタンスを作成
	config := huma.DefaultConfig("Todo API", "1.0.0")
	config.Info.Description = "Go製のTodo管理API"
	config.Info.Contact = &huma.Contact{Name: "API Support"}

	api := humachi.New(router, config)

	// ヘルスチェックエンドポイント
	huma.Register(api, huma.Operation{
		OperationID: "get-health",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "アプリケーションヘルスチェック",
		Tags:        []string{"health"},
	}, healthHandler)

	huma.Register(api, huma.Operation{
		OperationID: "get-home",
		Method:      http.MethodGet,
		Path:        "/",
		Summary:     "ホームページ",
		Tags:        []string{"health"},
	}, homeHandler)

	huma.Register(api, huma.Operation{
		OperationID: "get-db-health",
		Method:      http.MethodGet,
		Path:        "/health/db",
		Summary:     "データベースヘルスチェック",
		Tags:        []string{"health"},
	}, dbHealthHandler)

	// Todo API エンドポイント
	huma.Register(api, huma.Operation{
		OperationID: "list-todos",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos",
		Summary:     "全てのTodoを取得",
		Description: "優先度や完了状況でフィルタリング可能",
		Tags:        []string{"todos"},
	}, todoHandler.GetAllTodos)

	huma.Register(api, huma.Operation{
		OperationID:   "create-todo",
		Method:        http.MethodPost,
		Path:          "/api/v1/todos",
		Summary:       "新しいTodoを作成",
		Tags:          []string{"todos"},
		DefaultStatus: 201,
	}, todoHandler.CreateTodo)

	huma.Register(api, huma.Operation{
		OperationID: "get-todo",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos/{id}",
		Summary:     "特定のTodoを取得",
		Tags:        []string{"todos"},
	}, todoHandler.GetTodoByID)

	huma.Register(api, huma.Operation{
		OperationID: "update-todo",
		Method:      http.MethodPut,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Todoを更新",
		Tags:        []string{"todos"},
	}, todoHandler.UpdateTodo)

	huma.Register(api, huma.Operation{
		OperationID: "delete-todo",
		Method:      http.MethodDelete,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Todoを削除",
		Tags:        []string{"todos"},
	}, todoHandler.DeleteTodo)

	// サーバーの起動
	port := ":8080"
	fmt.Printf("Todo API サーバーがポート%sで起動しています...\n", port)
	fmt.Println("利用可能なエンドポイント:")
	fmt.Println("  GET    /                    - ホームページ")
	fmt.Println("  GET    /health              - ヘルスチェック")
	fmt.Println("  GET    /health/db           - DBヘルスチェック")
	fmt.Println("  GET    /api/v1/todos        - 全Todoを取得")
	fmt.Println("  POST   /api/v1/todos        - 新しいTodoを作成")
	fmt.Println("  GET    /api/v1/todos/{id}   - 特定のTodoを取得")
	fmt.Println("  PUT    /api/v1/todos/{id}   - Todoを更新")
	fmt.Println("  DELETE /api/v1/todos/{id}   - Todoを削除")
	fmt.Println("  GET    /docs                - OpenAPI ドキュメント")

	// HTTPサーバーの設定
	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	// グレースフルシャットダウンの設定
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("サーバー起動エラー: %v", err)
		}
	}()

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("サーバーをシャットダウンしています...")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("サーバーシャットダウンエラー: %v", err)
	}

	// データベース接続を閉じる
	if err := db.Close(); err != nil {
		log.Printf("データベース接続の終了エラー: %v", err)
	}

	log.Println("サーバーがシャットダウンしました")
}
