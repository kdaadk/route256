package repository

import (
	"context"
	"route256/cart/internal/model"
	"testing"
)

func BenchmarkCartRepo_AddItemToCart(b *testing.B) {
	repo := NewCartRepository()
	item := model.DtoItem{
		SkuId: 123,
		Count: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := int64(i % 100)
		item.SkuId = int64(i)
		_ = repo.AddItemToCart(context.Background(), userID, &item)
	}
}

func BenchmarkCartRepo_GetItems(b *testing.B) {
	repo := NewCartRepository()
	item := model.DtoItem{
		SkuId: 123,
		Count: 1,
	}

	for i := 0; i < 100; i++ {
		userId := int64(i)
		for j := 0; j < 10; j++ {
			item.SkuId = int64(j)
			_ = repo.AddItemToCart(context.Background(), userId, &item)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userId := int64(i % 100)
		_, _ = repo.GetItems(context.Background(), userId)
	}
}

func BenchmarkCartRepo_DeleteAllFromCart(b *testing.B) {
	repo := NewCartRepository()
	item := model.DtoItem{
		SkuId: 123,
		Count: 1,
	}

	for i := 0; i < 100; i++ {
		userId := int64(i)
		for j := 0; j < 10; j++ {
			item.SkuId = int64(j)
			_ = repo.AddItemToCart(context.Background(), userId, &item)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userId := int64(i % 100)
		_ = repo.DeleteAllFromCart(context.Background(), userId)
	}
}

func BenchmarkCartRepo_DeleteFromCart(b *testing.B) {
	repo := NewCartRepository()
	item := model.DtoItem{
		SkuId: 123,
		Count: 1,
	}

	for i := 0; i < 100; i++ {
		userId := int64(i)
		for j := 0; j < 10; j++ {
			item.SkuId = int64(j)
			_ = repo.AddItemToCart(context.Background(), userId, &item)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userId := int64(i % 100)
		_ = repo.DeleteFromCart(context.Background(), userId, item.SkuId)
	}
}
