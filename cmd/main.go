package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"

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
		fmt.Println(dateRange.GetPeriodFrom())
		stream, err := client.GetAggregatedCategory(context.Background(), dateRange)
		if err != nil {
			panic(err)
		}
		var toUser string = ""
		for {
			cat, err := stream.Recv()
			if err == io.EOF {
				log.Println("endd")
				log.Println(cat)
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			toUser += cat.String()
			fmt.Println(toUser)
		}
		fmt.Println(toUser)
	})

	if err := router.Run(port); err != nil {
		panic(err)
	}
}
