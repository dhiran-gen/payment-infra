package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/payment-infra/payment-api/internal"
)

func main() {
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Load config from env
	pgHost := os.Getenv("POSTGRES_HOST")
	pgPort := os.Getenv("POSTGRES_PORT")
	pgUser := os.Getenv("POSTGRES_USER")
	pgPass := os.Getenv("POSTGRES_PASSWORD")
	pgDB := os.Getenv("POSTGRES_DB")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPass, pgDB)
	db, err := internal.NewDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	dynamoEndpoint := os.Getenv("DYNAMO_ENDPOINT")
	dynamoTable := os.Getenv("DYNAMO_TABLE")
	ctx := context.Background()
	dynamo, err := internal.NewDynamoCache(ctx, dynamoEndpoint, dynamoTable)
	if err != nil {
		log.Fatalf("failed to connect to dynamodb: %v", err)
	}

	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	producer, err := internal.NewKafkaProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()

	// Configurable table/topic names
	tableName := os.Getenv("PAYMENT_TABLE")
	if tableName == "" {
		tableName = "payments"
	}
	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (id TEXT PRIMARY KEY, amount INT)`, tableName))
	if err != nil {
		log.Fatalf("failed to create payments table: %v", err)
	}

	dynamoTable = os.Getenv("DYNAMO_TABLE")
	_, err = dynamo.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: &dynamoTable,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: "S"},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: "HASH"},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{ReadCapacityUnits: aws.Int64(1), WriteCapacityUnits: aws.Int64(1)},
	})
	// Ignore error if table already exists

	// CRUD endpoints
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "payments.commands"
	}
	r.POST("/payments", func(c *gin.Context) {
		var req struct {
			ID     string `json:"id"`
			Amount int    `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p := internal.Payment{ID: req.ID, Amount: req.Amount}
		if err := db.CreatePayment(p); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		_ = dynamo.PutPayment(p.ID, p.Amount) // cache write-through
		_ = producer.SendMessage(topic, []byte(p.ID), []byte(fmt.Sprintf("%s:%d", p.ID, p.Amount)))
		c.JSON(http.StatusOK, gin.H{"status": "payment created"})
	})

	r.GET("/payments/:id", func(c *gin.Context) {
		id := c.Param("id")
		// Try cache first
		amount, err := dynamo.GetPayment(id)
		if err == nil && amount != 0 {
			c.JSON(http.StatusOK, gin.H{"id": id, "amount": amount, "cache": true})
			return
		}
		// Fallback to DB
		query := fmt.Sprintf("SELECT id, amount FROM %s WHERE id = $1", tableName)
		row := db.QueryRow(query, id)
		var p internal.Payment
		if err := row.Scan(&p.ID, &p.Amount); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		_ = dynamo.PutPayment(p.ID, p.Amount) // cache fill
		c.JSON(http.StatusOK, gin.H{"id": p.ID, "amount": p.Amount, "cache": false})
	})

	r.GET("/payments", func(c *gin.Context) {
		query := fmt.Sprintf("SELECT id, amount FROM %s", tableName)
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer rows.Close()
		var payments []internal.Payment
		for rows.Next() {
			var p internal.Payment
			if err := rows.Scan(&p.ID, &p.Amount); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
				return
			}
			payments = append(payments, p)
		}
		c.JSON(http.StatusOK, payments)
	})

	r.PUT("/payments/:id", func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Amount int `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Update in DB
		query := fmt.Sprintf("UPDATE %s SET amount = $1 WHERE id = $2", tableName)
		_, err := db.Exec(query, req.Amount, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if err := dynamo.PutPayment(id, req.Amount); err != nil {
			log.Printf("dynamo cache update failed: %v", err)
		}
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	})

	r.DELETE("/payments/:id", func(c *gin.Context) {
		id := c.Param("id")
		query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)
		_, err := db.Exec(query, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		// No delete in DynamoDB cache for simplicity
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}

}
