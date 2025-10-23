# 💪 GoLv2 筋トレ記録アプリ

**Go言語Lv2評価基準を楽しく学ぶための筋トレ記録アプリ！** 🏋️‍♂️

このアプリは、Go言語のLv2評価基準（命名、変数、定数、関数、並行処理、型、構造体、インターフェース、エラーハンドリング、パッケージモジュール、パフォーマンス）を実践的に学ぶために作られた筋トレ記録アプリです。

## 🎯 特徴

- **💪 筋トレ記録**: ワークアウトの作成、更新、削除、一覧表示
- **🔥 難易度管理**: 初心者から野獣級まで4段階の難易度
- **🏋️ 筋肉群別管理**: 胸、背中、脚、腕、腹筋、肩など
- **📊 統計機能**: 総重量、完了数、筋肉群別統計
- **🎮 楽しいメッセージ**: 筋トレに励ましのメッセージ
- **🚀 gRPC API**: 高性能なRPCフレームワーク
- **🐳 Docker対応**: 簡単な環境構築
- **🗄️ MySQL/GORM**: データベース永続化

## 🏗️ アーキテクチャ

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   gRPC Client   │    │   gRPC Server   │    │   Repository    │
│   (Evans)       │◄──►│   (Workout)     │◄──►│   (GORM/MySQL)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 クイックスタート

### 1. 環境構築

```bash
# リポジトリをクローン
git clone <repository-url>
cd go_lv2

# プロトコルバッファーの生成
make proto

# Dockerで起動
make docker-up
```

### 2. Evansで接続

```bash
# EvansでgRPCサーバーに接続
evans -r repl -p 50051

# パッケージとサービスを選択
package workout
service WorkoutService
```

### 3. ワークアウトを作成

```bash
# ベンチプレスを作成
call CreateWorkout
{
  "name": "ベンチプレス",
  "description": "胸をバキバキに鍛える！",
  "difficulty": "DIFFICULTY_INTERMEDIATE",
  "muscle_group": "胸",
  "sets": 3,
  "reps": 10,
  "weight": 60.0,
  "notes": "今日は調子がいい！💪"
}
```

### 4. ワークアウト一覧を取得

```bash
# 全ワークアウトを取得
call ListWorkouts
{
  "status_filter": "WORKOUT_STATUS_PLANNED"
}
```

### 5. 統計を確認

```bash
# 今日の統計を取得
call GetWorkoutStats
{
  "period": "today"
}
```

## 📚 学習項目

### 🏷️ 命名規則
- **変数名**: `workoutManager`, `muscleGroup`, `totalWeight`
- **パッケージ名**: `workout`, `repository`, `server`
- **レシーバー名**: `wm *WorkoutManager`, `r *GORMRepository`

### 🔧 変数・定数
- **変数宣言**: `var repo repository.WorkoutRepository`
- **初期値**: `workout := &repository.Workout{...}`
- **定数**: `const (WorkoutStatusPlanned WorkoutStatus = iota)`

### ⚙️ 関数・メソッド
- **関数**: `NewGORMRepository(dsn string)`
- **メソッド**: `(wm *WorkoutManager) CreateWorkout(...)`
- **オプション引数**: `CreateWorkout(name string, opts ...WorkoutOption)`

### 🔄 並行処理
- **ゴルーチン**: 統計計算での並行処理
- **チャネル**: エラーハンドリングでのチャネル使用

### 🏗️ 型・構造体
- **カスタム型**: `type WorkoutID int64`
- **構造体**: `type Workout struct {...}`
- **ジェネリクス**: `Filter[T any](slice []T, predicate func(T) bool)`

### 🔌 インターフェース
- **インターフェース**: `type WorkoutRepository interface {...}`
- **実装**: `GORMRepository`と`MockWorkoutRepository`

### ❌ エラーハンドリング
- **カスタムエラー**: `TaskError`構造体
- **エラー比較**: `IsTaskNotFound(err error)`
- **エラーラッピング**: `fmt.Errorf("failed to create workout: %w", err)`

### 📦 パッケージ・モジュール
- **パッケージ分離**: `repository`, `server`, `proto`
- **依存性注入**: `NewWorkoutManagerWithRepository(repo)`

### ⚡ パフォーマンス
- **文字列結合**: `strings.Builder`の使用
- **スライス最適化**: 事前割り当てとキャパシティ指定

## 🛠️ 開発コマンド

```bash
# プロトコルバッファーの生成
make proto

# テスト実行
make test

# Docker起動
make docker-up

# Docker停止
make docker-down

# ログ確認
make docker-logs

# 環境リセット
make docker-reset
```

## 📊 サンプルデータ

アプリには以下の楽しいサンプルデータが含まれています：

- **ベンチプレス**: 胸をバキバキに鍛える！モテ男への第一歩 💪
- **スクワット**: 下半身を鍛えてモテ男に！ 😅
- **デッドリフト**: 背中を広くして逆三角形に！野獣級の重量に挑戦！ 🔥
- **腕立て伏せ**: 腕を太くして彼女にアピール 💀
- **プランク**: 腹筋を割ってビーチボディに！1分間が永遠に感じる ⏰

## 🎯 難易度レベル

1. **初心者**: 基本的なワークアウト
2. **中級者**: 少しハードなワークアウト
3. **上級者**: 本格的なワークアウト
4. **野獣級**: 超人的なワークアウト 🔥

## 📚 ドキュメント

詳細なドキュメントは `docs/` ディレクトリにあります：
- [📋 アーキテクチャ](./docs/ARCHITECTURE.md) - システム設計の詳細
- [📁 ディレクトリ構造](./docs/DIRECTORY_STRUCTURE.md) - プロジェクト構造の説明
- [🔌 APIリファレンス](./docs/API_REFERENCE.md) - gRPC APIの詳細仕様
- [🛠️ 開発ガイド](./docs/DEVELOPMENT_GUIDE.md) - 開発環境とフロー
- [🎯 学習目標](./docs/LEARNING_OBJECTIVES.md) - Go言語Lv2評価基準との対応

## 🎉 楽しく学ぼう！

このアプリでGo言語のLv2評価基準を楽しく学びながら、筋トレも一緒に記録しましょう！

**💪 今日も筋肉を鍛えましょう！** 🏋️‍♂️

---

*Made with 💪 and Go*
