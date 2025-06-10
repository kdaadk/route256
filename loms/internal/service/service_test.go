package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"os"
	"os/exec"
	"route256/loms/internal/kafka"
	"route256/loms/internal/model"
	"route256/loms/internal/repository"
	proto "route256/loms/proto"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type ServiceIntegrationTestSuite struct {
	suite.Suite
	db      *model.DBPools
	service *Service
	ctx     context.Context
}

func TestServiceIntegration(t *testing.T) {
	suite.Run(t, new(ServiceIntegrationTestSuite))
}

func (s *ServiceIntegrationTestSuite) SetupSuite() {
	var err error
	s.db, err = setupTestDB()
	if err != nil {
		s.FailNow("Failed to setup test database", err)
	}

	s.ctx = context.Background()
	txManager := repository.NewTxManager(s.db)
	ordersRepo := repository.NewOrdersRepository(s.db)
	stocksRepo := repository.NewStocksRepository(s.db)
	kafkaCli, err := kafka.NewKafkaProducer("loms.order-events")
	s.service = NewService(ordersRepo, stocksRepo, txManager, kafkaCli)
}

func (s *ServiceIntegrationTestSuite) TearDownTest() {
	_, err := s.db.Master.Exec(s.ctx, `
		TRUNCATE order_items, orders, stocks RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		s.T().Logf("Cleanup failed: %v", err)
	}
}

func setupTestDB() (*model.DBPools, error) {
	const (
		masterURL     = "postgres://loms:loms@localhost:5432/loms?sslmode=disable"
		replicaURL    = "postgres://loms:loms@localhost:5433/loms?sslmode=disable"
		migrationsDir = "../../migrations"
		gooseBinPath  = "../../bin/goose"
	)

	// Apply migrations to both databases
	if err := applyMigrations(gooseBinPath, migrationsDir, masterURL); err != nil {
		return nil, fmt.Errorf("master DB migration failed: %w", err)
	}
	if err := applyMigrations(gooseBinPath, migrationsDir, replicaURL); err != nil {
		return nil, fmt.Errorf("replica DB migration failed: %w", err)
	}

	// Create database connections
	masterDB, err := pgxpool.New(context.Background(), masterURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}

	replicaDB, err := pgxpool.New(context.Background(), replicaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to replica database: %w", err)
	}

	return &model.DBPools{Master: masterDB, Replica: replicaDB}, nil
}

func applyMigrations(goosePath, migrationsDir, dbURL string) error {
	cmd := exec.Command(goosePath, "-dir", migrationsDir, "postgres", dbURL, "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply migrations to %s: %w", dbURL, err)
	}
	return nil
}

func (s *ServiceIntegrationTestSuite) prepareTestOrder() (orderId int64, skuId int32, reserved int32, totalCount int32) {
	reserved = 1
	skuId = rand.Int31()
	totalCount = 100
	s.db.Master.Exec(s.ctx, "INSERT INTO stocks (sku, total_count, reserved) VALUES ($1, $2, 0);", skuId, totalCount)
	orderId, err := s.service.CreateOrder(&proto.CreateOrderRequest{
		UserId: rand.Int63(),
		Items:  []*proto.OrderItem{{Sku: skuId, Count: reserved}},
	})
	assert.NoError(s.T(), err)
	assert.Greater(s.T(), orderId, int64(0))

	return orderId, skuId, reserved, totalCount
}

func (s *ServiceIntegrationTestSuite) TestCreateOrder() {
	orderId, skuId, _, _ := s.prepareTestOrder()

	var status int64
	err := s.db.Replica.QueryRow(s.ctx,
		"SELECT status FROM orders WHERE id = $1", orderId).Scan(&status)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(model.OrderStatus_AwaitingPayment), status)

	var reserved int64
	err = s.db.Replica.QueryRow(s.ctx,
		"SELECT reserved FROM stocks WHERE sku = $1", skuId).Scan(&reserved)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), reserved)
}

func (s *ServiceIntegrationTestSuite) TestPayOrder() {
	orderId, skuId, payed, totalCount := s.prepareTestOrder()

	payReq := &proto.PayOrderRequest{OrderId: orderId}
	err := s.service.PayOrder(payReq)
	assert.NoError(s.T(), err)

	var status model.OrderStatus
	err = s.db.Replica.QueryRow(s.ctx,
		"SELECT status FROM orders WHERE id = $1", orderId).Scan(&status)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), model.OrderStatus_Paid, status)

	var reserved, total int64
	err = s.db.Replica.QueryRow(s.ctx,
		"SELECT reserved, total_count FROM stocks WHERE sku = $1", skuId).Scan(&reserved, &total)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), reserved)
	assert.Equal(s.T(), int64(totalCount-payed), total)
}

func (s *ServiceIntegrationTestSuite) TestCancelOrder() {
	orderId, skuId, _, _ := s.prepareTestOrder()

	cancelReq := &proto.CancelOrderRequest{OrderId: orderId}
	err := s.service.CancelOrder(cancelReq)
	assert.NoError(s.T(), err)

	var status model.OrderStatus
	err = s.db.Replica.QueryRow(s.ctx,
		"SELECT status FROM orders WHERE id = $1", orderId).Scan(&status)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), model.OrderStatus_Cancelled, status)

	var reserved int64
	err = s.db.Replica.QueryRow(s.ctx,
		"SELECT reserved FROM stocks WHERE sku = $1", skuId).Scan(&reserved)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), reserved)
}
