package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

	if err != nil {
		log.Fatal("Connection error:", err)
	}
	// testing of conneciton
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("error of ping:", err)
	}
	fmt.Println("succesfully conneciton")
	return client

}

// colleciton = table
func GetColleciton(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("taxihub").Collection(collectionName)
}
