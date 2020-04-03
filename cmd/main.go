package main

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	service "github.com/softwareengineer-test-task/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

const (
	port = ":8080"
	addr = "localhost:8000"
)

var router *gin.Engine

func main() {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := service.NewTicketServiceClient(conn)
	router := gin.Default()
	router.POST("/aggregated", func(ctx *gin.Context) {
		dateRange := &service.DateRange{}
		b, _ := ioutil.ReadAll(ctx.Request.Body)
		err := json.Unmarshal(b, &dateRange)
		if err != nil {
			panic(err)
		}
		stream, err := client.GetAggregatedCategory(context.Background(), dateRange)
		if err != nil {
			panic(err)
		}
		var results []*service.Categories
		for {
			cat, err := stream.Recv()
			if err == io.EOF {
				log.Println("end")
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			results = append(results, cat)
		}
		ctx.JSON(http.StatusOK, gin.H{
			"results": results,
		})
		return
	})
	router.POST("/byticket", func(ctx *gin.Context) {
		dateRange := &service.DateRange{}
		b, _ := ioutil.ReadAll(ctx.Request.Body)
		err := json.Unmarshal(b, &dateRange)
		if err != nil {
			panic(err)
		}
		stream, err := client.GetScoresByTickets(context.Background(), dateRange)
		if err != nil {
			panic(err)
		}
		var results []*service.TicketScores
		for {
			tik, err := stream.Recv()
			if err == io.EOF {
				log.Println("end")
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			results = append(results, tik)
		}
		ctx.JSON(http.StatusOK, gin.H{
			"results": results,
		})
		return
	})

	if err := router.Run(port); err != nil {
		panic(err)
	}
}
