package seeder

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"test-golang-project/internal/entity"

	"gorm.io/gorm"
)

type categoryTodoSeed struct {
	CategoryName string
	Todos        []struct {
		Title       string
		Description string
		IsCompleted bool
	}
}

type userTodoSeed struct {
	UserEmail     string
	CategoryTodos []categoryTodoSeed
}

// SeedTodos creates 2 todos for every category belonging to the 4 active users
func SeedTodos(db *gorm.DB) error {
	data := []userTodoSeed{
		{
			UserEmail: "admin@example.com",
			CategoryTodos: []categoryTodoSeed{
				{
					CategoryName: "Elektronik & Gadget",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Beli mouse wireless kantor",
							Description: "Pilih mouse ergonomis untuk tim developer",
							IsCompleted: true,
						},
						{
							Title:       "Upgrade RAM server development",
							Description: "Tambah 32GB RAM DDR5 untuk server CI/CD",
							IsCompleted: false,
						},
					},
				},
				{
					CategoryName: "Peralatan Kantor",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Restock kertas HVS A4",
							Description: "Order 5 rim kertas untuk printer HR",
							IsCompleted: true,
						},
						{
							Title:       "Servis AC ruang meeting",
							Description: "Jadwalkan teknisi untuk maintenance AC ruang rapat lantai 2",
							IsCompleted: false,
						},
					},
				},
				{
					CategoryName: "Software & Lisensi",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Perpanjang lisensi JetBrains",
							Description: "Renew subscription tahunan untuk tim backend",
							IsCompleted: true,
						},
						{
							Title:       "Audit lisensi cloud AWS",
							Description: "Cek penggunaan resource EC2 dan S3 yang tidak aktif",
							IsCompleted: false,
						},
					},
				},
			},
		},
		{
			UserEmail: "user@example.com",
			CategoryTodos: []categoryTodoSeed{
				{
					CategoryName: "Pakaian & Fashion",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Beli kemeja putih lengan panjang",
							Description: "Untuk persiapan acara onboarding minggu depan",
							IsCompleted: true,
						},
						{
							Title:       "Jahit celana chino",
							Description: "Perbaiki ukuran panjang celana di tailor langganan",
							IsCompleted: false,
						},
					},
				},
				{
					CategoryName: "Sepatu & Tas",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Cuci sepatu sneakers",
							Description: "Bawa sepatu lari ke laundry sepatu sebelum akhir pekan",
							IsCompleted: true,
						},
						{
							Title:       "Beli ransel laptop waterproof",
							Description: "Cari ransel 20L dengan slot laptop 15.6 inch untuk kerja komuter",
							IsCompleted: false,
						},
					},
				},
				{
					CategoryName: "Aksesoris & Jam Tangan",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Ganti baterai jam tangan",
							Description: "Bawa ke counter servis resmi di mall",
							IsCompleted: true,
						},
						{
							Title:       "Beli kacamata anti radiasi",
							Description: "Pilih lensa blueray untuk kerja di depan monitor",
							IsCompleted: false,
						},
					},
				},
			},
		},
		{
			UserEmail: "nao@example.com",
			CategoryTodos: []categoryTodoSeed{
				{
					CategoryName: "Buku & Referensi",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Selesaikan baca buku The Go Programming Language",
							Description: "Pelajari Bab 8 dan 9 tentang Goroutines dan Channels",
							IsCompleted: true,
						},
						{
							Title:       "Beli buku Designing Data-Intensive Applications",
							Description: "Edisi terbaru untuk referensi arsitektur sistem terdistribusi",
							IsCompleted: false,
						},
					},
				},
				{
					CategoryName: "Hobi & Koleksi",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Rakit Gunpla RG RX-78-2",
							Description: "Selesaikan perakitan part kaki dan backpack",
							IsCompleted: true,
						},
						{
							Title:       "Beli cat akrilik dan kuas detail",
							Description: "Warna titanium gold dan matte black untuk detailing model kit",
							IsCompleted: false,
						},
					},
				},
				{
					CategoryName: "Fotografi & Video",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Bersihkan sensor kamera mirrorless",
							Description: "Gunakan cleaning kit sensor swab",
							IsCompleted: true,
						},
						{
							Title:       "Beli tripod carbon fiber",
							Description: "Cari yang ringkas untuk travel dan kokoh menahan beban 5kg",
							IsCompleted: false,
						},
					},
				},
			},
		},
		{
			UserEmail: "shin@example.com",
			CategoryTodos: []categoryTodoSeed{
				{
					CategoryName: "Makanan & Minuman",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Beli biji kopi arabika Gayo",
							Description: "Pilih roast profile medium-to-dark 500 gram",
							IsCompleted: true,
						},
						{
							Title:       "Belanja mingguan bahan makanan",
							Description: "Ayam fillet, telur, sayur brokoli, dan buah apel",
							IsCompleted: false,
						},
					},
				},
				{
					CategoryName: "Kesehatan & Nutrisi",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Beli multivitamin harian",
							Description: "Stok vitamin C 500mg dan vitamin D3",
							IsCompleted: true,
						},
						{
							Title:       "Restock kotak P3K rumah",
							Description: "Beli perban, alkohol 70%, dan plester luka",
							IsCompleted: false,
						},
					},
				},
				{
					CategoryName: "Olahraga & Fitness",
					Todos: []struct {
						Title       string
						Description string
						IsCompleted bool
					}{
						{
							Title:       "Lari pagi 5 km di stadion",
							Description: "Target pace di bawah 6 menit per kilometer",
							IsCompleted: true,
						},
						{
							Title:       "Beli resistance band set",
							Description: "Pilih yang ada 5 level beban (10-50 lbs) untuk latihan di rumah",
							IsCompleted: false,
						},
					},
				},
			},
		},
	}

	for _, item := range data {
		var user entity.User
		fallbackEmail := strings.Replace(item.UserEmail, "@example.com", "@mail.com", 1)
		if err := db.Where("(email = ? OR email = ?) AND deleted_at IS NULL", item.UserEmail, fallbackEmail).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("  -> [TodoSeeder] User with email '%s' not found or soft-deleted, skipping...\n", item.UserEmail)
				continue
			}
			return err
		}

		for _, catSeed := range item.CategoryTodos {
			var category entity.Category
			if err := db.Where("user_id = ? AND name = ? AND deleted_at IS NULL", user.ID, catSeed.CategoryName).First(&category).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Printf("  -> [TodoSeeder] Category '%s' for user '%s' not found, skipping...\n", catSeed.CategoryName, user.Email)
					continue
				}
				return err
			}

			categoryID := category.ID
			for _, td := range catSeed.Todos {
				var existing entity.Todo
				err := db.Where("user_id = ? AND category_id = ? AND title = ?", user.ID, categoryID, td.Title).First(&existing).Error
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						desc := td.Description
						newTodo := entity.Todo{
							UserID:      user.ID,
							CategoryID:  &categoryID,
							Title:       td.Title,
							Description: &desc,
							IsCompleted: td.IsCompleted,
						}
						if err := db.Create(&newTodo).Error; err != nil {
							return fmt.Errorf("failed to seed todo '%s' for user %s: %w", td.Title, user.Email, err)
						}
						log.Printf("  -> [TodoSeeder] Inserted todo '%s' in category '%s' for user '%s' (ID: %d)\n", td.Title, category.Name, user.Name, user.ID)
					} else {
						return err
					}
				} else {
					log.Printf("  -> [TodoSeeder] Todo '%s' already exists for user '%s', skipping\n", td.Title, user.Name)
				}
			}
		}
	}

	return nil
}
