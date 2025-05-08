package main

import (
	"fmt"

	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "myuser"
	password = "mypassword"
	dbname   = "mydatabase"
)

func auth(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	var jwtSecret = []byte("Test")
	if cookie == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Missing or invalid token")
	}

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return jwtSecret, nil
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token claims")
	}

	c.Locals("userID", claims.Subject)

	return c.Next()
}

func main() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Book{}, &User{})
	fmt.Println("Migrate Successfully")
	app := fiber.New()
	app.Use("/books", auth)
	app.Get("/books", func(c *fiber.Ctx) error {
		books, err := getBooks(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch books")
		}
		return c.JSON(books)
	})
	app.Get("/books/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		idInt, err := strconv.Atoi(idParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid book ID")
		}

		id := uint(idInt)
		book, err := getBook(db, id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Database error")
		}
		if book == nil {
			return c.Status(fiber.StatusNotFound).SendString("Book not found")
		}

		return c.JSON(book)
	})
	app.Post("/books", func(c *fiber.Ctx) error {
		var book Book
		if err := c.BodyParser(&book); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}

		err := createBook(db, &book)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to create book")
		}

		return c.Status(fiber.StatusCreated).JSON(book)
	})
	app.Put("/books/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		idInt, err := strconv.Atoi(idParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid book ID")
		}
		id := uint(idInt)

		var book Book
		if err := c.BodyParser(&book); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}
		book.ID = id // make sure the ID is set for GORM to update the correct record

		err = updateBook(db, &book)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to update book")
		}

		return c.JSON(book)
	})
	app.Delete("/books/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		idInt, err := strconv.Atoi(idParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid book ID")
		}
		id := uint(idInt)

		err = deleteBook(db, id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete book")
		}

		return c.SendString("Book deleted successfully")
	})

	app.Post("/register", func(c *fiber.Ctx) error {
		var user User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}

		err := createUser(db, &user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to create user")
		}

		user.Password = ""
		return c.Status(fiber.StatusCreated).JSON(user)
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		var user User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}

		token, err := loginUser(db, &user)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
		}
		c.Cookie(&fiber.Cookie{
			Name:     "jwt",
			Value:    token,
			Expires:  time.Now().Add(time.Hour * 24),
			HTTPOnly: true,
		})
		return c.JSON(fiber.Map{
			"token": token,
		})
	})

	app.Listen(":8000")
}
