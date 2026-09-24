// Package main はクリーンアーキテクチャの構成例を実行します。
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"

	"sample/application"
	"sample/infrastructure/persistence/query"
	"sample/infrastructure/persistence/repository"
	"sample/infrastructure/txmanager"
)

func main() {
	// DB初期化（例: PostgreSQLやMySQLなど）
	db, err := sql.Open("sqlite", "file:memdb1?mode=memory&cache=shared")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// 最初にテーブルを確実に作成
	setupTables(db)

	// 1. インフラ層の生成
	txManager := txmanager.NewSqlTxManager(db)
	userRepo := repository.NewSQLiteUserRepository(db)
	orderRepo := repository.NewSQLiteOrderRepository(db)
	queryService := query.NewSQLiteUserQueryService(db)

	// 2. ユースケース層の生成（DI）
	userUseCase := application.NewUserUseCase(txManager, userRepo, orderRepo, queryService)

	ctx := context.Background()

	// 3. 【更新処理】User登録 ＋ 初期Order登録（1トランザクション）
	fmt.Println("--- 更新処理実行 ---")
	err = userUseCase.RegisterUserWithInitialOrder(ctx, "Alice", "alice@example.com", 5000)
	if err != nil {
		fmt.Printf("Register Failed: %v\n", err)
	} else {
		fmt.Println("User & Initial Order registered successfully!")
	}

	err = userUseCase.RegisterOrder(ctx, 1, 3000)
	if err != nil {
		fmt.Printf("Register Failed: %v\n", err)
	} else {
		fmt.Println("User & Initial Order registered successfully!")
	}

	// 4. 【参照系実行】JOIN集計データの高速取得（CQRS）
	summary, err := userUseCase.GetSummary(ctx, 1)
	if err != nil {
		fmt.Printf("GetSummary Failed: %v\n", err)
	} else {
		fmt.Printf("Summary: %+v\n", summary)
	}
}

// setupTables はアプリケーションが利用するテーブルを作成します。
func setupTables(db *sql.DB) {
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		amount INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(createTablesSQL); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}
