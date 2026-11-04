package seeder

import (
	"errors"
	"log"

	"test-golang-project/internal/entity"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) error {
	defaultPassword := "password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []entity.User{
		{
			Name:     "Administrator",
			Email:    "admin@example.com",
			Password: string(hashedPassword),
			IsAdmin:  true,
		},
		{
			Name:     "Test User",
			Email:    "user@example.com",
			Password: string(hashedPassword),
			IsAdmin:  false,
		},
	}

	for _, user := range users {
		var existing entity.User
		// Idempotent: check if email already exists so it doesn't duplicate
		err := db.Where("email = ?", user.Email).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&user).Error; err != nil {
					return err
				}
				log.Printf("  -> [UserSeeder] Inserted user: %s (%s)\n", user.Name, user.Email)
			} else {
				return err
			}
		} else {
			log.Printf("  -> [UserSeeder] User already exists, skipping: %s (%s)\n", user.Name, user.Email)
		}
	}

	return nil
}
