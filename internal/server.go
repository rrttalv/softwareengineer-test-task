package server

import (
	"context"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"

	service "github.com/softwareengineer-test-task/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
}

type Rating struct {
	Weight float64
	Value  int
}

func main() {
	listener, err := net.Listen("tcp", ":4040")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	service.RegisterTicketServiceServer(srv, &server{})
	reflection.Register(srv)
	if serr := srv.Serve(listener); serr != nil {
		panic(err)
	}
}

func (s *server) GetAggregatedCategory(filter *service.DateRange, stream service.TicketService_GetAggregatedCategoryServer) error {
	return nil
}

func (s *server) GetOveralQuality(ctx context.Context, request *service.DateRange) (*service.Quality, error) {
	return nil, nil
}

func (s *server) GetPeriodOverPeriod(ctx context.Context, request *service.DateRange) (*service.PeriodChange, error) {
	return nil, nil
}

func (s *server) GetScoresByTickets(tickets *service.Tickets, stream service.TicketService_GetScoresByTicketsServer) error {
	//ticketIDS := request.GetIds()
	i := proto.Int32(1)
	fmt.Print(i)
	return nil
}

/*
	map[int]map[int]int is this: INT - ID of rating, INT - RatingValue = INT - RatingWeight
	Ratings are calculated in the following manner:
	take the rating value and multiply it with it's weight
	finally return a integer precentage
*/
func calculateRatingPrecentage(totalWeight int, ratings map[int]Rating) int {
	var finalRatingPrecentage = 0
	for _, r := range ratings {
		fv := r.Value
		var ratingVal int
		if r.Weight == 0 {
			//Does not make a signifigant difference, but still give the rating very little weight
			ratingVal = int(0.01 * float64(fv))
		} else {
			ratingVal = int(r.Weight * float64(fv))
		}
		finalRatingPrecentage += (ratingVal * 100) / int(totalWeight)
	}
	return finalRatingPrecentage
}
