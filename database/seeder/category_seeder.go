package seeder

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"test-golang-project/internal/entity"

	"gorm.io/gorm"
)

type userCategorySeed struct {
	UserEmail  string
	Categories []struct {
		Name        string
		Description string
	}
}

// SeedCategories creates 3 categories for each of the 4 active users
func SeedCategories(db *gorm.DB) error {
	data := []userCategorySeed{
		{
			UserEmail: "admin@example.com",
			Categories: []struct {
				Name        string
				Description string
			}{
				{Name: "Elektronik & Gadget", Description: "Perangkat elektronik, laptop, smartphone, dan aksesoris kantor"},
				{Name: "Peralatan Kantor", Description: "Kebutuhan inventaris, alat tulis kantor, dan perabot kerja"},
				{Name: "Software & Lisensi", Description: "Lisensi aplikasi, langganan cloud, dan sistem operasi"},
			},
		},
		{
			UserEmail: "user@example.com",
			Categories: []struct {
				Name        string
				Description string
			}{
				{Name: "Pakaian & Fashion", Description: "Koleksi pakaian kasual, formal, dan pakaian luar"},
				{Name: "Sepatu & Tas", Description: "Sepatu sneakers, formal, ransel, dan dompet"},
				{Name: "Aksesoris & Jam Tangan", Description: "Kacamata, topi, jam tangan, dan perhiasan"},
			},
		},
		{
			UserEmail: "nao@mail.com",
			Categories: []struct {
				Name        string
				Description string
			}{
				{Name: "Buku & Referensi", Description: "Buku fiksi, sains, teknologi, dan komik"},
				{Name: "Hobi & Koleksi", Description: "Model kit, board game, dan pernak-pernik koleksi"},
				{Name: "Fotografi & Video", Description: "Kamera, lensa, gimbal, dan pencahayaan studio"},
			},
		},
		{
			UserEmail: "shin@mail.com",
			Categories: []struct {
				Name        string
				Description string
			}{
				{Name: "Makanan & Minuman", Description: "Kudapan ringan, kopi, teh, dan bahan masakan"},
				{Name: "Kesehatan & Nutrisi", Description: "Suplemen vitamin, obat herbal, dan masker"},
				{Name: "Olahraga & Fitness", Description: "Matras yoga, dumbbell, pakaian olahraga, dan botol minum"},
			},
		},
	}

	for _, item := range data {
		var user entity.User
		fallbackEmail := strings.Replace(item.UserEmail, "@example.com", "@mail.com", 1)
		if err := db.Where("(email = ? OR email = ?) AND deleted_at IS NULL", item.UserEmail, fallbackEmail).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("  -> [CategorySeeder] User with email '%s' not found or soft-deleted, skipping...\n", item.UserEmail)
				continue
			}
			return err
		}

		for _, cat := range item.Categories {
			var existing entity.Category
			err := db.Where("user_id = ? AND name = ?", user.ID, cat.Name).First(&existing).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					newCategory := entity.Category{
						UserID:      user.ID,
						Name:        cat.Name,
						Description: cat.Description,
					}
					if err := db.Create(&newCategory).Error; err != nil {
						return fmt.Errorf("failed to seed category %s for user %s: %w", cat.Name, user.Email, err)
					}
					log.Printf("  -> [CategorySeeder] Inserted category '%s' for user '%s' (ID: %d)\n", cat.Name, user.Name, user.ID)
				} else {
					return err
				}
			} else {
				log.Printf("  -> [CategorySeeder] Category '%s' already exists for user '%s', skipping\n", cat.Name, user.Name)
			}
		}
	}

	return nil
}
