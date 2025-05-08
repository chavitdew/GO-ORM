package main

import (
	"errors"

	"gorm.io/gorm"
)

type Book struct {
	gorm.Model
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Price       uint   `json:"price"`
}

func createBook(db *gorm.DB, book *Book) error {
	result := db.Create(book)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func getBook(db *gorm.DB, id uint) (*Book, error) {
	var book Book
	result := db.First(&book, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // not found, return nil without error
		}
		return nil, result.Error // actual DB error
	}
	return &book, nil
}

func getBooks(db *gorm.DB) ([]Book, error) {
	var books []Book
	result := db.Find(&books)
	if result.Error != nil {
		return nil, result.Error
	}
	return books, nil
}

func updateBook(db *gorm.DB, book *Book) error {
	result := db.Model(&book).Updates(book)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func deleteBook(db *gorm.DB, id uint) error {
	result := db.Delete(&Book{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func searchBook(db *gorm.DB, bookName string) ([]Book, error) {
	var books []Book
	result := db.Where("name = ?", bookName).Order("price").Find(&books)
	if result.Error != nil {
		return nil, result.Error
	}
	return books, nil
}
