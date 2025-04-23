package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

// Todo merepresentasikan struktur tugas
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Category  string    `json:"category"`
	DueDate   time.Time `json:"dueDate"`
}

// Database sementara
var todos = []Todo{
	{ID: 1, Title: "Belajar Go", Completed: false, Category: "belajar", DueDate: time.Now().Add(24 * time.Hour)},
	{ID: 2, Title: "Belajar Fiber", Completed: false, Category: "belajar", DueDate: time.Now().Add(48 * time.Hour)},
	{ID: 3, Title: "Belajar Tailwind", Completed: false, Category: "belajar", DueDate: time.Now().Add(72 * time.Hour)},
}

// Kategori yang tersedia
var categories = []string{"belajar", "kerja", "pribadi", "lainnya"}

// lastID untuk penomoran ID baru
var lastID = 3

func main() {
	// Inisialisasi template engine
	engine := html.New("./views", ".html")

	// Inisialisasi aplikasi Fiber
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Middleware
	app.Use(logger.New())

	// Static files
	app.Static("/", "./public")

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Todos":      todos,
			"Categories": categories,
		})
	})

	// API Routes
	api := app.Group("/api")

	// Mendapatkan semua todo
	api.Get("/todos", func(c *fiber.Ctx) error {
		return c.JSON(todos)
	})

	// Mendapatkan semua kategori
	api.Get("/categories", func(c *fiber.Ctx) error {
		return c.JSON(categories)
	})

	// Menambahkan todo baru
	api.Post("/todos", func(c *fiber.Ctx) error {
		todo := new(Todo)
		if err := c.BodyParser(todo); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tidak dapat memproses permintaan",
			})
		}

		lastID++
		todo.ID = lastID
		todo.Completed = false

		// Jika tidak ada kategori, set ke 'lainnya'
		if todo.Category == "" {
			todo.Category = "lainnya"
		}

		// Jika tanggal tenggat waktu kosong, set ke 7 hari dari sekarang
		if todo.DueDate.IsZero() {
			todo.DueDate = time.Now().Add(7 * 24 * time.Hour)
		}

		todos = append(todos, *todo)
		return c.Status(fiber.StatusCreated).JSON(todo)
	})

	// Mengubah status todo
	api.Put("/todos/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "ID tidak valid",
			})
		}

		todo := new(Todo)
		if err := c.BodyParser(todo); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tidak dapat memproses permintaan",
			})
		}

		for i, t := range todos {
			if t.ID == id {
				todos[i].Completed = todo.Completed
				todos[i].Title = todo.Title

				// Update kategori jika ada
				if todo.Category != "" {
					todos[i].Category = todo.Category
				}

				// Update tanggal tenggat waktu jika ada
				if !todo.DueDate.IsZero() {
					todos[i].DueDate = todo.DueDate
				}

				return c.JSON(todos[i])
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Todo tidak ditemukan",
		})
	})

	// Menghapus todo
	api.Delete("/todos/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "ID tidak valid",
			})
		}

		for i, t := range todos {
			if t.ID == id {
				todos = append(todos[:i], todos[i+1:]...)
				return c.SendStatus(fiber.StatusNoContent)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Todo tidak ditemukan",
		})
	})

	// Menambahkan kategori baru
	api.Post("/categories", func(c *fiber.Ctx) error {
		var data map[string]string
		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tidak dapat memproses permintaan",
			})
		}

		category := data["category"]
		if category == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Kategori tidak boleh kosong",
			})
		}

		// Cek jika kategori sudah ada
		for _, cat := range categories {
			if cat == category {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Kategori sudah ada",
				})
			}
		}

		categories = append(categories, category)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"category": category,
		})
	})

	// Menjalankan server
	log.Fatal(app.Listen(":3000"))
}
