package main

import (
	"Order_PaymentPlatform/internal/repository"
	httpTransport "Order_PaymentPlatform/internal/transport/http"
	"Order_PaymentPlatform/internal/usecase"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	dsn := "host=localhost port=5432 user=postgres password=0000 dbname=payment_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("database not reachable: ", err)
	}

	log.Println("Connected to PostgreSQL")

	paymentRepo := repository.NewPaymentRepository(db)
	paymentUC := usecase.NewPaymentUsecase(paymentRepo)
	paymentHandler := httpTransport.NewPaymentHandler(paymentUC)

	r := gin.Default()

	r.POST("/payments", paymentHandler.CreatePayment)
	r.GET("/payments/:order_id", paymentHandler.GetPayment)

	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
