package main

import (
	"context"
	"log"
	"time"

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
type User struct {
	UserID      string    `json:"user_id" bson:"user_id"`
	Username    string    `json:"username" bson:"username"`
	Email       string    `json:"email" bson:"email"`
	Password    string    `json:"-" bson:"password"`
	FirstName   string    `json:"first_name" bson:"first_name"`
	LastName    string    `json:"last_name" bson:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth" bson:"date_of_birth"`
	PhoneNumber string    `json:"phone_number" bson:"phone_number"`

	LastLoginAt time.Time `json:"last_login_at" bson:"last_login_at"`
	MutualFunds []Fund    `json:"mutual_funds" bson:"mutual_funds"`
}

var collection *mongo.Collection
var userCollection *mongo.Collection

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
	userCollection = client.Database("mutual_funds").Collection("users")

	router.GET("/getAllFunds", getAllFunds)
	router.POST("/addFund", addFund)
	router.GET("/user/:userID", getUser)
	router.POST("/addUser", addUser)
	router.DELETE("/deleteUser", deleteUser)

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

func getUser(c *gin.Context) {
	userID := c.Param("userID")

	var user User
	err := userCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(404, gin.H{"error": "User not found"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(200, user)
}

func addUser(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Generate a unique user ID (you may want to use a more robust method in production)
	user.UserID = generateUniqueUserID()

	// Set the last login time to the current time
	user.LastLoginAt = time.Now()

	_, err := userCollection.InsertOne(context.TODO(), user)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"result": "success", "user_id": user.UserID})
}

func deleteUser(c *gin.Context) {
	userID := c.Param("userID")

	result, err := userCollection.DeleteOne(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	c.JSON(200, gin.H{"result": "success", "message": "User deleted successfully"})
}

func generateUniqueUserID() string {
	return time.Now().Format("20060102150405")
}
