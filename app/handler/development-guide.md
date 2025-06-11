# Huma Framework ハンドラー開発ガイド

このガイドでは、Humaフレームワークを使用して新しいAPIハンドラーを追加する手順を説明します。

## 目次

1. [Humaフレームワークについて](#humaフレームワークについて)
2. [ハンドラー追加の基本手順](#ハンドラー追加の基本手順)
3. [リクエスト・レスポンス構造体の定義](#リクエストレスポンス構造体の定義)
4. [ハンドラー関数の実装](#ハンドラー関数の実装)
5. [ルート登録](#ルート登録)
6. [バリデーション](#バリデーション)
7. [エラーハンドリング](#エラーハンドリング)
8. [OpenAPIドキュメント](#openapi-ドキュメント)
9. [ベストプラクティス](#ベストプラクティス)

## Humaフレームワークについて

Humaは、Go言語でREST APIを構築するためのモダンなフレームワークです。

### 主な特徴
- **型安全性**: 構造体ベースのリクエスト/レスポンス定義
- **自動バリデーション**: 構造体タグによる入力値検証
- **OpenAPI自動生成**: コードからSwagger/OpenAPIドキュメントを自動生成
- **高パフォーマンス**: Chi routerをベースとした高速ルーティング

## ハンドラー追加の基本手順

新しいAPIエンドポイントを追加する際は、以下の手順に従ってください：

### 1. リクエスト・レスポンス構造体の定義
### 2. ハンドラー関数の実装
### 3. main.goでのルート登録
### 4. テストの実装

## リクエスト・レスポンス構造体の定義

### 基本パターン

```go
// リクエスト構造体（パスパラメータ）
type UserIDRequest struct {
    ID int `path:"id" doc:"ユーザーID" minimum:"1"`
}

// リクエスト構造体（クエリパラメータ）
type UserQueryRequest struct {
    Name   string `query:"name" doc:"ユーザー名でフィルタリング"`
    Active string `query:"active" doc:"アクティブ状態でフィルタリング"`
}

// リクエスト構造体（ボディ）
type UserCreateRequest struct {
    Body UserCreateData `doc:"作成するユーザー情報"`
}

// レスポンス構造体
type UserResponse struct {
    Body struct {
        Data    *UserData `json:"data" doc:"ユーザー情報"`
        Message string    `json:"message" doc:"レスポンスメッセージ"`
    }
}
```

### 構造体タグの説明

| タグ | 用途 | 例 |
|------|------|-----|
| `path` | パスパラメータ | `path:"id"` |
| `query` | クエリパラメータ | `query:"name"` |
| `header` | HTTPヘッダー | `header:"Authorization"` |
| `json` | JSONフィールド | `json:"user_name"` |
| `doc` | ドキュメント | `doc:"ユーザー名"` |
| `minimum` | 最小値 | `minimum:"1"` |
| `maximum` | 最大値 | `maximum:"100"` |
| `enum` | 列挙値 | `enum:"low,medium,high"` |

### 重要なポイント

1. **ポインター型の使用**: パスパラメータ、クエリパラメータではポインター型を使用できません
2. **omitemptyタグ**: 更新リクエストでオプショナルフィールドには `json:",omitempty"` を使用
3. **構造体の入れ子**: Body部分は無名構造体として定義することが多い

## ハンドラー関数の実装

### 基本構造

```go
func (h *YourHandler) YourMethod(ctx context.Context, input *YourRequest) (*YourResponse, error) {
    // 1. ビジネスロジックの実行
    data, err := h.service.DoSomething(input.ID)
    if err != nil {
        // 2. エラーハンドリング
        return nil, huma.Error500InternalServerError(err.Error())
    }

    // 3. レスポンスの構築
    return &YourResponse{
        Body: struct {
            Data    *YourData `json:"data" doc:"データ"`
            Message string    `json:"message" doc:"メッセージ"`
        }{
            Data:    data,
            Message: "処理が完了しました",
        },
    }, nil
}
```

### ハンドラーメソッドの命名規則

| HTTP Method | 関数名の例 | 説明 |
|-------------|------------|------|
| GET (list) | `GetAllUsers` | 一覧取得 |
| GET (single) | `GetUserByID` | 単体取得 |
| POST | `CreateUser` | 作成 |
| PUT | `UpdateUser` | 更新 |
| DELETE | `DeleteUser` | 削除 |

## ルート登録

main.goでのルート登録例：

```go
// ユーザー管理API
huma.Register(api, huma.Operation{
    OperationID: "list-users",          // OpenAPIのoperationId
    Method:      http.MethodGet,        // HTTPメソッド
    Path:        "/api/v1/users",       // パス
    Summary:     "全てのユーザーを取得", // 概要
    Description: "ページネーション対応", // 詳細説明
    Tags:        []string{"users"},     // タグ（グループ化）
}, userHandler.GetAllUsers)

huma.Register(api, huma.Operation{
    OperationID:   "create-user",
    Method:        http.MethodPost,
    Path:          "/api/v1/users",
    Summary:       "新しいユーザーを作成",
    Tags:          []string{"users"},
    DefaultStatus: 201,                 // デフォルトステータスコード
}, userHandler.CreateUser)

huma.Register(api, huma.Operation{
    OperationID: "get-user",
    Method:      http.MethodGet,
    Path:        "/api/v1/users/{id}",  // パスパラメータ
    Summary:     "特定のユーザーを取得",
    Tags:        []string{"users"},
}, userHandler.GetUserByID)
```

### ルート登録時の注意点

1. **OperationID**: 一意である必要があります
2. **Tags**: 関連するエンドポイントをグループ化するために使用
3. **Path**: パスパラメータは `{param}` 形式で記述
4. **DefaultStatus**: POST操作では通常201を指定

## バリデーション

### 構造体タグによるバリデーション

```go
type UserCreateData struct {
    Name     string `json:"name" minLength:"1" maxLength:"100" doc:"ユーザー名"`
    Email    string `json:"email" format:"email" doc:"メールアドレス"`
    Age      int    `json:"age" minimum:"0" maximum:"150" doc:"年齢"`
    Status   string `json:"status" enum:"active,inactive" doc:"ステータス"`
    Tags     []string `json:"tags" maxItems:"10" doc:"タグリスト"`
}
```

### 使用可能なバリデーションタグ

| タグ | 説明 | 例 |
|------|------|-----|
| `minLength` | 最小文字数 | `minLength:"1"` |
| `maxLength` | 最大文字数 | `maxLength:"100"` |
| `minimum` | 最小値 | `minimum:"0"` |
| `maximum` | 最大値 | `maximum:"150"` |
| `format` | フォーマット | `format:"email"` |
| `enum` | 列挙値 | `enum:"active,inactive"` |
| `pattern` | 正規表現 | `pattern:"^[a-zA-Z]+$"` |
| `minItems` | 配列最小要素数 | `minItems:"1"` |
| `maxItems` | 配列最大要素数 | `maxItems:"10"` |

## エラーハンドリング

### 標準エラーレスポンス

```go
// 400 Bad Request
return nil, huma.Error400BadRequest("無効なリクエストです")

// 401 Unauthorized
return nil, huma.Error401Unauthorized("認証が必要です")

// 403 Forbidden
return nil, huma.Error403Forbidden("アクセス権限がありません")

// 404 Not Found
return nil, huma.Error404NotFound("リソースが見つかりません")

// 409 Conflict
return nil, huma.Error409Conflict("リソースが競合しています")

// 422 Unprocessable Entity
return nil, huma.Error422UnprocessableEntity("処理できないエンティティです")

// 500 Internal Server Error
return nil, huma.Error500InternalServerError("内部サーバーエラーです")

// 503 Service Unavailable
return nil, huma.Error503ServiceUnavailable("サービスが利用できません")
```

### エラーハンドリングのパターン

```go
func (h *UserHandler) GetUserByID(ctx context.Context, input *UserIDRequest) (*UserResponse, error) {
    user, err := h.service.GetUserByID(uint(input.ID))
    if err != nil {
        // エラーの種類に応じて適切なHTTPステータスを返す
        if strings.Contains(err.Error(), "見つかりません") {
            return nil, huma.Error404NotFound(err.Error())
        }
        return nil, huma.Error500InternalServerError(err.Error())
    }

    return &UserResponse{
        Body: struct {
            Data    *UserData `json:"data" doc:"ユーザー情報"`
            Message string    `json:"message" doc:"メッセージ"`
        }{
            Data:    user.ToResponse(),
            Message: "ユーザーを取得しました",
        },
    }, nil
}
```

## OpenAPI ドキュメント

### 自動生成される情報

Humaは以下の情報を自動でOpenAPIドキュメントに含めます：

- **リクエスト・レスポンススキーマ**: 構造体から自動生成
- **バリデーション情報**: 構造体タグから自動生成
- **エンドポイント情報**: `huma.Operation`から自動生成

### アクセス方法

- **ドキュメントページ**: `http://localhost:8080/docs`
- **OpenAPI仕様書**: `http://localhost:8080/openapi.json`

### ドキュメントの品質向上

```go
// 詳細な説明とサンプルを含む構造体
type UserCreateData struct {
    Name  string `json:"name" doc:"ユーザー名" example:"田中太郎" minLength:"1" maxLength:"100"`
    Email string `json:"email" doc:"メールアドレス" example:"tanaka@example.com" format:"email"`
    Age   int    `json:"age" doc:"年齢" example:"25" minimum:"0" maximum:"150"`
}
```

## ベストプラクティス

### 1. 構造体の設計

```go
// Good: 明確で一貫性のある命名
type TodoCreateRequest struct {
    Body TodoCreateData `doc:"作成するTodo情報"`
}

type TodoListResponse struct {
    Body struct {
        Data    []*TodoResponse `json:"data" doc:"Todoリスト"`
        Message string          `json:"message" doc:"メッセージ"`
        Count   int             `json:"count" doc:"総数"`
    }
}
```

### 2. エラーハンドリング

```go
// Good: エラーの種類に応じた適切なステータスコード
if errors.Is(err, ErrNotFound) {
    return nil, huma.Error404NotFound("リソースが見つかりません")
}
if errors.Is(err, ErrValidation) {
    return nil, huma.Error400BadRequest(err.Error())
}
return nil, huma.Error500InternalServerError("内部エラーが発生しました")
```

### 3. レスポンス構造の統一

```go
// Good: 一貫したレスポンス構造
type APIResponse struct {
    Data    interface{} `json:"data" doc:"レスポンスデータ"`
    Message string      `json:"message" doc:"メッセージ"`
    Count   *int        `json:"count,omitempty" doc:"総数（リスト取得時のみ）"`
}
```

### 4. ドキュメント品質の向上

```go
// Good: わかりやすい説明とサンプルデータ
type TodoPriority string `enum:"low,medium,high,urgent" doc:"Todo の優先度" example:"high"`
```

### 5. ファイル構成

```
app/
├── handler/
│   ├── todo_handler.go          # Todoハンドラー
│   ├── user_handler.go          # ユーザーハンドラー
│   ├── huma_todo_handler.go     # Huma版Todoハンドラー
│   └── development-guide.md     # このガイド
├── service/
├── db/
└── main.go
```

## まとめ

Humaフレームワークを使用することで：

1. **型安全性**: コンパイル時にAPIの整合性をチェック
2. **自動ドキュメント生成**: 手動でのドキュメント管理が不要
3. **バリデーション**: 構造体タグによる自動バリデーション
4. **開発効率向上**: ボイラープレートコードの削減

新しいAPIエンドポイントを追加する際は、この手順に従って一貫性のあるAPIを構築してください。
