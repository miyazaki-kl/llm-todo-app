# Goサーバー Docker プロジェクト

このプロジェクトは、Docker ComposeでGoサーバーを構築するためのテンプレートです。

## 構成

- **Goアプリケーション**: ポート8080で動作するWebサーバー
- **PostgreSQL**: データベース（ポート5432）

## 起動方法

### 1. Dockerコンテナの起動

```bash
docker compose up -d
```

### 2. ログの確認

```bash
docker compose logs -f app
```

### 3. APIのテスト

```bash
# ホームページ
curl http://localhost:8080/

# ヘルスチェック
curl http://localhost:8080/health
```

## 開発時の使用方法

### Goモジュールの初期化（コンテナ内で実行される）

コンテナ起動時に自動的に`go mod init myapp`が実行されます。

### 依存関係の追加

```bash
docker compose exec app go get <パッケージ名>
```

### アプリケーションの再起動

```bash
docker compose restart app
```

## ファイル構成

```
.
├── app/                 # アプリケーションディレクトリ
│   ├── main.go         # メインのGoアプリケーション
│   ├── go.mod          # Go modules設定
│   └── db/             # データベース関連
├── compose.yaml         # Docker Compose設定
├── test.sh             # 動作確認スクリプト
└── README.md           # このファイル
```

## API エンドポイント

### 基本エンドポイント
- `GET /` - ホームページ
- `GET /health` - アプリケーションヘルスチェック
- `GET /health/db` - データベース接続ヘルスチェック

### Todo API (RESTful)
- `GET /api/v1/todos` - 全てのTodoを取得
  - クエリパラメータ: `?priority=high&completed=false`
- `POST /api/v1/todos` - 新しいTodoを作成
- `GET /api/v1/todos/{id}` - 特定のTodoを取得
- `PUT /api/v1/todos/{id}` - Todoを更新
- `DELETE /api/v1/todos/{id}` - Todoを削除

### Todo リクエスト例

**Todo作成 (POST /api/v1/todos)**
```json
{
  "title": "重要なタスク",
  "description": "明日までに完了する必要があります",
  "priority": "high",
  "due_date": "2025-06-12T15:00:00Z"
}
```

**Todo更新 (PUT /api/v1/todos/1)**
```json
{
  "completed": true,
  "priority": "low"
}
```

## 環境変数

- `GO_ENV`: 実行環境（development/production）
- `CGO_ENABLED`: CGOの有効/無効
- `GOOS`: ターゲットOS
- `GOARCH`: ターゲットアーキテクチャ

## トラブルシューティング

### コンテナが起動しない場合

```bash
docker compose down
docker compose up --build
```

### ポートが使用中の場合

`compose.yaml`のポート設定を変更してください。

```yaml
ports:
  - "8081:8080"  # ホストポートを8081に変更
``` 