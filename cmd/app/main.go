package main

import (
	"log"

	"github.com/Cork-Holdings/gp_payment_orchestration/cmd"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Failed to load env: %v", err.Error())
	}

	global.New()
	cmd.Execute()
}
