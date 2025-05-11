package repository

import (
	"fmt"
	"route256/cart/internal/model"
	"strconv"
	"testing"
)

func BenchmarkCartRepo_AddItemToCart(b *testing.B) {
	repo := NewCartRepository()
	item := model.Item{
		SkuId: 123,
		Count: 1,
		Name:  "Test Product",
		Price: 1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := int64(i % 100)
		item.SkuId = int64(i)
		item.Name = "Product " + strconv.Itoa(i)
		_ = repo.AddItemToCart(userID, &item)
	}
}

func BenchmarkCartRepo_GetItems(b *testing.B) {
	repo := NewCartRepository()
	item := model.Item{
		SkuId: 123,
		Count: 1,
		Name:  "Test Product",
		Price: 1000,
	}

	for i := 0; i < 100; i++ {
		userId := int64(i)
		for j := 0; j < 10; j++ {
			item.SkuId = int64(j)
			item.Name = fmt.Sprintf("Product %d", j)
			_ = repo.AddItemToCart(userId, &item)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userId := int64(i % 100)
		_, _ = repo.GetItems(userId)
	}
}

func BenchmarkCartRepo_DeleteAllFromCart(b *testing.B) {
	repo := NewCartRepository()
	item := model.Item{
		SkuId: 123,
		Count: 1,
		Name:  "Test Product",
		Price: 1000,
	}

	for i := 0; i < 100; i++ {
		userId := int64(i)
		for j := 0; j < 10; j++ {
			item.SkuId = int64(j)
			item.Name = fmt.Sprintf("Product %d", j)
			_ = repo.AddItemToCart(userId, &item)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userId := int64(i % 100)
		_ = repo.DeleteAllFromCart(userId)
	}
}

func BenchmarkCartRepo_DeleteFromCart(b *testing.B) {
	repo := NewCartRepository()
	item := model.Item{
		SkuId: 123,
		Count: 1,
		Name:  "Test Product",
		Price: 1000,
	}

	for i := 0; i < 100; i++ {
		userId := int64(i)
		for j := 0; j < 10; j++ {
			item.SkuId = int64(j)
			item.Name = fmt.Sprintf("Product %d", j)
			_ = repo.AddItemToCart(userId, &item)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userId := int64(i % 100)
		_ = repo.DeleteFromCart(userId, item.SkuId)
	}
}
