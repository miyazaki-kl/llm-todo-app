#!/bin/bash

# カラーコードの定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ログ関数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# テスト結果を格納する変数
TESTS_PASSED=0
TESTS_FAILED=0

# テスト関数
run_test() {
    local test_name="$1"
    local url="$2"
    local expected_status="$3"
    
    log_info "テスト実行中: $test_name"
    
    # HTTPステータスコードを取得
    status_code=$(curl -s -o /dev/null -w "%{http_code}" "$url")
    
    if [ "$status_code" = "$expected_status" ]; then
        log_success "$test_name - ステータスコード: $status_code"
        ((TESTS_PASSED++))
        
        # レスポンスボディも表示
        response=$(curl -s "$url")
        echo "  レスポンス: $response"
    else
        log_error "$test_name - 期待値: $expected_status, 実際: $status_code"
        ((TESTS_FAILED++))
    fi
    echo
}

# メイン実行
main() {
    echo "=================================================="
    echo "     Goサーバー 動作確認テスト"
    echo "=================================================="
    echo
    
    # サーバーの基本情報
    log_info "サーバーURL: http://localhost:8080"
    log_info "開始時刻: $(date)"
    echo
    
    # Docker Composeの状態確認
    log_info "Docker Composeサービスの状態確認"
    docker compose ps
    echo
    
    # サーバーが起動するまで少し待機
    log_info "サーバー起動の待機中..."
    sleep 3
    echo
    
    # テスト実行
    log_info "APIエンドポイントのテスト開始"
    echo
    
    # ホームページエンドポイントのテスト
    run_test "ホームページ (GET /)" "http://localhost:8080/" "200"
    
    # ヘルスチェックエンドポイントのテスト
    run_test "ヘルスチェック (GET /health)" "http://localhost:8080/health" "200"
    
    # データベースヘルスチェックエンドポイントのテスト
    run_test "DBヘルスチェック (GET /health/db)" "http://localhost:8080/health/db" "200"
    
    # 存在しないエンドポイントのテスト（404エラーが期待される）
    run_test "存在しないエンドポイント (GET /notfound)" "http://localhost:8080/notfound" "404"
    
    # Todo CRUD API のテスト
    log_info "Todo CRUD APIのテスト開始"
    echo
    
    # 初期状態のTodoリスト取得
    run_test "Todoリスト取得 (GET /api/v1/todos)" "http://localhost:8080/api/v1/todos" "200"
    
    # 新しいTodoを作成
    log_info "新しいTodoを作成中..."
    create_response=$(curl -s -X POST http://localhost:8080/api/v1/todos \
        -H "Content-Type: application/json" \
        -d '{"title":"テストTodo","description":"テスト用のTodo","priority":"high"}')
    
    if echo "$create_response" | grep -q "Todoを作成しました"; then
        log_success "Todo作成 - 成功"
        echo "  レスポンス: $create_response"
        
        # 作成されたTodoのIDを抽出
        todo_id=$(echo "$create_response" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        
        if [ -n "$todo_id" ]; then
            log_info "作成されたTodoのID: $todo_id"
            
            # 特定のTodoを取得
            run_test "特定のTodo取得 (GET /api/v1/todos/$todo_id)" "http://localhost:8080/api/v1/todos/$todo_id" "200"
            
            # Todoを更新
            log_info "Todoを更新中..."
            update_response=$(curl -s -X PUT http://localhost:8080/api/v1/todos/$todo_id \
                -H "Content-Type: application/json" \
                -d '{"completed":true,"priority":"low"}')
            
            if echo "$update_response" | grep -q "Todoを更新しました"; then
                log_success "Todo更新 - 成功"
                echo "  レスポンス: $update_response"
                ((TESTS_PASSED++))
            else
                log_error "Todo更新 - 失敗"
                echo "  レスポンス: $update_response"
                ((TESTS_FAILED++))
            fi
            echo
            
            # Todoを削除
            log_info "Todoを削除中..."
            delete_response=$(curl -s -X DELETE http://localhost:8080/api/v1/todos/$todo_id)
            
            if echo "$delete_response" | grep -q "Todoを削除しました"; then
                log_success "Todo削除 - 成功"
                echo "  レスポンス: $delete_response"
                ((TESTS_PASSED++))
            else
                log_error "Todo削除 - 失敗"
                echo "  レスポンス: $delete_response"
                ((TESTS_FAILED++))
            fi
            echo
            
        else
            log_error "作成されたTodoのIDを取得できませんでした"
            ((TESTS_FAILED++))
        fi
    else
        log_error "Todo作成 - 失敗"
        echo "  レスポンス: $create_response"
        ((TESTS_FAILED++))
    fi
    echo
    
    # レスポンス時間の測定
    log_info "レスポンス時間の測定"
    response_time=$(curl -s -o /dev/null -w "%{time_total}" "http://localhost:8080/")
    echo "  ホームページのレスポンス時間: ${response_time}秒"
    echo
    
    # サーバーログの確認
    log_info "最新のサーバーログ (最後の10行)"
    docker compose logs --tail=10 app
    echo
    
    # テスト結果のサマリー
    echo "=================================================="
    echo "              テスト結果サマリー"
    echo "=================================================="
    log_success "成功: $TESTS_PASSED テスト"
    if [ $TESTS_FAILED -gt 0 ]; then
        log_error "失敗: $TESTS_FAILED テスト"
    else
        log_success "失敗: $TESTS_FAILED テスト"
    fi
    echo
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "すべてのテストが成功しました！✅"
        exit 0
    else
        log_error "一部のテストが失敗しました ❌"
        exit 1
    fi
}

# 使用方法の表示
show_usage() {
    echo "使用方法: $0 [オプション]"
    echo
    echo "オプション:"
    echo "  -h, --help     このヘルプメッセージを表示"
    echo "  -q, --quiet    詳細な出力を抑制"
    echo "  -v, --verbose  詳細な出力を表示"
    echo
    echo "例:"
    echo "  $0              # 通常の動作確認テストを実行"
    echo "  $0 --quiet      # 簡潔な出力でテストを実行"
}

# コマンドライン引数の処理
case "$1" in
    -h|--help)
        show_usage
        exit 0
        ;;
    -q|--quiet)
        exec > /dev/null 2>&1
        main
        ;;
    -v|--verbose)
        set -x
        main
        ;;
    "")
        main
        ;;
    *)
        log_error "不明なオプション: $1"
        show_usage
        exit 1
        ;;
esac
