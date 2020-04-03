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

type clientServer struct {
	router *gin.Engine
	client service.TicketServiceClient
}

func main() {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := service.NewTicketServiceClient(conn)
	router := gin.Default()
	cs := clientServer{
		router: router,
		client: client,
	}
	router.POST("/aggregated", cs.GetAggregated)
	router.POST("/byticket", cs.GetScoresByTicket)
	if err := router.Run(port); err != nil {
		panic(err)
	}
}

func (cs *clientServer) GetAggregated(ctx *gin.Context) {
	dateRange := &service.DateRange{}
	b, _ := ioutil.ReadAll(ctx.Request.Body)
	err := json.Unmarshal(b, &dateRange)
	if err != nil {
		panic(err)
	}
	stream, err := cs.client.GetAggregatedCategory(context.Background(), dateRange)
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
}

func (cs *clientServer) GetScoresByTicket(ctx *gin.Context) {
	dateRange := &service.DateRange{}
	b, _ := ioutil.ReadAll(ctx.Request.Body)
	err := json.Unmarshal(b, &dateRange)
	if err != nil {
		panic(err)
	}
	stream, err := cs.client.GetScoresByTickets(context.Background(), dateRange)
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
}
