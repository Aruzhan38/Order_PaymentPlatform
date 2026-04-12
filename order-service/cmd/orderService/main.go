package main

import (
	"Order_PaymentPlatform/internal/app"
	"Order_PaymentPlatform/internal/repository"
	httpTransport "Order_PaymentPlatform/internal/transport/http"
	"Order_PaymentPlatform/internal/usecase"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
)

func main() {
	dsn := "host=localhost port=5432 user=postgres password=0000 dbname=order_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("database not reachable:", err)
	}

	log.Println("Connected to PostgreSQL")

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	orderRepo := repository.NewOrderRepository(db)

	paymentClient := app.NewPaymentHTTPClient(
		"http://localhost:8081",
		httpClient,
	)

	orderUC := usecase.NewOrderUsecase(orderRepo, paymentClient)

	orderHandler := httpTransport.NewOrderHandler(orderUC)

	r := gin.Default()

	r.POST("/orders", orderHandler.CreateOrder)
	r.GET("/orders/:id", orderHandler.GetOrder)
	r.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
