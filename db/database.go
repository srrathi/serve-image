package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// To connect to mongodb
func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// To get mongodb connection URI from .env file
	uri := os.Getenv("MONGOURI")
	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Println("Error: " + err.Error())
	}

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Println("Error: " + err.Error())
	}
	log.Println("Connected to MongoDB")
	return client
}
