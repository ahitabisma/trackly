package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	token := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	hashHex := hex.EncodeToString(hash[:])

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL env var required")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	err = db.Exec(
		`INSERT INTO api_tokens (token_hash, role, label) VALUES (?, ?, ?)`,
		hashHex, "admin", "Hermes Telegram Bot",
	).Error
	if err != nil {
		panic(err)
	}

	fmt.Println("Token admin (SIMPAN SEKARANG, tidak akan ditampilkan lagi):")
	fmt.Println(token)
	fmt.Println("Pasang di config Hermes sebagai header: Authorization: Bearer " + token)
}
