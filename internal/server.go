package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/mxk/go-sqlite/sqlite3"

	"github.com/golang/protobuf/proto"

	service "github.com/softwareengineer-test-task/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	Database *sqlite3.Conn
}

type Rating struct {
	Weight float64
	Value  int
}

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	conn, dberr := sqlite3.Open("database.db")
	if dberr != nil {
		log.Println("Can't connect to database!")
		panic(err)
	}
	server := &server{}
	server.Database = conn
	srv := grpc.NewServer()
	service.RegisterTicketServiceServer(srv, server)
	reflection.Register(srv)
	if serr := srv.Serve(listener); serr != nil {
		panic(err)
	}
}

func (s *server) GetAggregatedCategory(filter *service.DateRange, stream service.TicketService_GetAggregatedCategoryServer) error {
	pf := filter.PeriodFrom
	pt := filter.PeriodTo
	var weekly = false
	//This sould run in go routines
	//dateFrom := dateToString(pf.GetMonth(), pf.GetDay(), pf.GetYear())
	//dateTo := dateToString(pt.GetMonth(), pt.GetDay(), pt.GetYear())
	if pf.GetDay()-pt.GetDay() > 30 {
		weekly = true
	}
	fmt.Println(weekly)
	return nil
}

func (s *server) GetScoresByTickets(tickets *service.Tickets, stream service.TicketService_GetScoresByTicketsServer) error {
	//ticketIDS := request.GetIds()
	i := proto.Int32(1)
	fmt.Print(i)
	return nil
}

func (s *server) GetOveralQuality(ctx context.Context, request *service.DateRange) (*service.Quality, error) {
	return nil, nil
}

func (s *server) GetPeriodOverPeriod(ctx context.Context, request *service.DateRange) (*service.PeriodChange, error) {
	return nil, nil
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

func dateToString(m int32, d int32, y int32) string {
	var res string
	res += fmt.Sprintf("%v-", y)
	res += fmt.Sprintf("%v-", m)
	res += fmt.Sprintf("%v", d)
	return res
}
