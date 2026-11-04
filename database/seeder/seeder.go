package seeder

import (
	"log"

	"gorm.io/gorm"
)

type SeederFunc func(db *gorm.DB) error

// RunAll executes all registered database seeders sequentially
func RunAll(db *gorm.DB) error {
	seeders := []struct {
		name string
		fn   SeederFunc
	}{
		{name: "UserSeeder", fn: SeedUsers},
		// Daftarkan seeder baru lainnya di sini nantinya, contoh:
		// {name: "RoleSeeder", fn: SeedRoles},
		// {name: "ProductSeeder", fn: SeedProducts},
	}

	for _, s := range seeders {
		log.Printf("==> Running %s...\n", s.name)
		if err := s.fn(db); err != nil {
			return err
		}
	}

	log.Println("==> All database seeders completed successfully!")
	return nil
}
