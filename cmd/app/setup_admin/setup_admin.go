package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"vinyl-vault/internal/config"
	"vinyl-vault/internal/repositories"
	"vinyl-vault/internal/services"

	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	fmt.Println("******* Vinyl Vault Admin Setup *******")

	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	userRepo := repositories.NewGormUserRepository(db)
	userService := services.NewUserService(userRepo)

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter admin username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("enter admin email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Enter admin password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("failed to read password")
	}

	password := string(passwordBytes)
	fmt.Println()

	fmt.Print("Confirm password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Failed to read password:", err)
	}
	confirm := string(confirmBytes)
	fmt.Println()

	if password != confirm {
		log.Fatal("Passwords do not match")
	}

	// create admin user
	ctx := context.Background()
	user, err := userService.Register(ctx, username, email, password)
	if err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	// set user as admin and save in db
	user.IsAdmin = true
	if err = userRepo.Save(ctx, user); err != nil {
		log.Fatal("Failed to set admin status", err)
	}

	fmt.Printf("\n Admin user '%s' created successfully!\n", username)
	fmt.Println("You can now start the server and generate registration keys.")

}
