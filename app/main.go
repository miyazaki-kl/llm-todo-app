package main

import (
	"encoding/json"
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

	"github.com/gorilla/mux"
)

// APIレスポンス用の構造体
type Response struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// ヘルスチェック用のハンドラー
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message:   "Goサーバーが正常に動作しています",
		Timestamp: time.Now(),
		Status:    "healthy",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ホームページ用のハンドラー
func homeHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message:   "Todo API サーバーへようこそ！",
		Timestamp: time.Now(),
		Status:    "success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// データベース接続状態チェック用のハンドラー
func dbHealthHandler(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()

	if database == nil {
		response := Response{
			Message:   "データベース接続が初期化されていません",
			Timestamp: time.Now(),
			Status:    "error",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}

	sqlDB, err := database.DB()
	if err != nil || sqlDB.Ping() != nil {
		response := Response{
			Message:   "データベース接続に問題があります",
			Timestamp: time.Now(),
			Status:    "error",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := Response{
		Message:   "データベース接続は正常です",
		Timestamp: time.Now(),
		Status:    "healthy",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ログミドルウェア
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
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
	todoHandler := handler.NewTodoHandler(todoService)

	// ルーターの設定
	r := mux.NewRouter()

	// ミドルウェアの追加
	r.Use(loggingMiddleware)

	// 基本エンドポイント
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/health/db", dbHealthHandler).Methods("GET")

	// Todo API エンドポイント
	todoRouter := r.PathPrefix("/api/v1/todos").Subrouter()
	todoRouter.HandleFunc("", todoHandler.GetAllTodos).Methods("GET")
	todoRouter.HandleFunc("", todoHandler.CreateTodo).Methods("POST")
	todoRouter.HandleFunc("/{id:[0-9]+}", todoHandler.GetTodoByID).Methods("GET")
	todoRouter.HandleFunc("/{id:[0-9]+}", todoHandler.UpdateTodo).Methods("PUT")
	todoRouter.HandleFunc("/{id:[0-9]+}", todoHandler.DeleteTodo).Methods("DELETE")

	// CORSの設定
	r.Use(func(next http.Handler) http.Handler {
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

	// グレースフルシャットダウンの設定
	go func() {
		log.Fatal(http.ListenAndServe(port, r))
	}()

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("サーバーをシャットダウンしています...")

	// データベース接続を閉じる
	if err := db.Close(); err != nil {
		log.Printf("データベース接続の終了エラー: %v", err)
	}

	log.Println("サーバーがシャットダウンしました")
}
