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
	router.POST("/quality", cs.GetOveralQuality)
	router.POST("/period", cs.GetPeriodOverPeriod)
	if err := router.Run(port); err != nil {
		panic(err)
	}
}

func (cs *clientServer) GetAggregated(ctx *gin.Context) {
	dateRange := &service.DateRange{}
	b, _ := ioutil.ReadAll(ctx.Request.Body)
	err := json.Unmarshal(b, &dateRange)
	if err != nil {
		sendErr(ctx, err)
		return
	}
	stream, err := cs.client.GetAggregatedCategory(context.Background(), dateRange)
	if err != nil {
		sendErr(ctx, err)
		return
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

func (cs *clientServer) GetOveralQuality(ctx *gin.Context) {
	dateRange := &service.DateRange{}
	b, _ := ioutil.ReadAll(ctx.Request.Body)
	err := json.Unmarshal(b, &dateRange)
	if err != nil {
		sendErr(ctx, err)
		return
	}
	if result, err := cs.client.GetOveralQuality(context.Background(), dateRange); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Error": err.Error(),
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"Result": result,
		})
	}
	return
}

func (cs *clientServer) GetScoresByTicket(ctx *gin.Context) {
	dateRange := &service.DateRange{}
	b, _ := ioutil.ReadAll(ctx.Request.Body)
	err := json.Unmarshal(b, &dateRange)
	if err != nil {
		sendErr(ctx, err)
		return
	}
	stream, err := cs.client.GetScoresByTickets(context.Background(), dateRange)
	if err != nil {
		sendErr(ctx, err)
		return
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

func (cs *clientServer) GetPeriodOverPeriod(ctx *gin.Context) {
	dateRange := &service.DoubleDateRange{}
	b, _ := ioutil.ReadAll(ctx.Request.Body)
	err := json.Unmarshal(b, &dateRange)
	if err != nil {
		sendErr(ctx, err)
		return
	}
	if result, err := cs.client.GetPeriodOverPeriod(context.Background(), dateRange); err != nil {
		sendErr(ctx, err)
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"result": result,
		})
	}
	return
}

func sendErr(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
