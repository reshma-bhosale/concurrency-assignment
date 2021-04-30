package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/gin-gonic/gin"
	"log"
	"time"
	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

)

type Result struct {
	Url    string `json:"url"` //Bind result to display json
	Length int    `json:"response_size"`
}

func main() {
	r := gin.Default()
	r.POST("/check", handle)
	r.Run(":8000")
}

func handle(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	var result Result
	c.Bind(&result)
	ch := make(chan int)
	url  := "https://www." + result.Url
	go responseSize(url, ch)
	result.Url = url
	result.Length = <-ch

	response := client.Database("go-training").Collection("responses")
	res, err := response.InsertOne(context.Background(), bson.D{{"Url" , result.Url}, {"Response size" , result.Length}})
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(200, result)
	fmt.Println("\nInserted ID: ", *res)

}

func responseSize(url string, channel chan int) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Something Bad Happened!!", err)
		return
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	channel <- len(body)

}
