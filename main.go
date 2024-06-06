package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CAGR struct {
	OneYear   float64 `json:"1_year" bson:"1_year"`
	ThreeYear float64 `json:"3_year" bson:"3_year"`
	FiveYear  float64 `json:"5_year" bson:"5_year"`
}

type Fund struct {
	Name     string `json:"name" bson:"name"`
	Category string `json:"category" bson:"category"`
	CAGR     []CAGR `json:"cagr" bson:"cagr"`
	Rating   int    `json:"rating" bson:"rating"`
}

var collection *mongo.Collection

func main() {
	router := gin.Default()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("mutual_funds").Collection("funds")

	router.GET("/getAllFunds", getAllFunds)
	router.POST("/addFund", addFund)

	router.Run()
}

func addFund(c *gin.Context) {
	var fund Fund

	if err := c.ShouldBindJSON(&fund); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error()})
		return
	}

	_, err := collection.InsertOne(context.TODO(), fund)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"result": "success"})
}

func getAllFunds(c *gin.Context) {

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.TODO())

	var funds []Fund
	for cursor.Next(context.TODO()) {
		var fund Fund
		if err := cursor.Decode(&fund); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		funds = append(funds, fund)
	}

	c.JSON(200, funds)
}
