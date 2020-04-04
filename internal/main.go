package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/softwareengineer-test-task/internal/helper"
	service "github.com/softwareengineer-test-task/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	Database *sql.DB
	Weight   int32
	Helper   *helper.Helper
}

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	database, dberr := sql.Open("sqlite3", "database.db")
	if dberr != nil {
		log.Println("Can't connect to database!")
		panic(err)
	}
	server := &server{
		Database: database,
		Helper:   &helper.Helper{},
	}
	if r, dberr := database.Query(`SELECT sum(rating_categories.weight*5) FROM rating_categories`); dberr != nil {
		panic(dberr)
	} else {
		var SM sql.NullString
		if r.Next() {
			r.Scan(&SM)
			wSum, _ := strconv.ParseFloat(SM.String, 1)
			server.Weight = int32(wSum)
		}
	}
	srv := grpc.NewServer()
	service.RegisterTicketServiceServer(srv, server)
	reflection.Register(srv)
	if serr := srv.Serve(listener); serr != nil {
		panic(err)
	}
}

func (s *server) GetAggregatedCategory(filter *service.DateRange, stream service.TicketService_GetAggregatedCategoryServer) error {
	from, to := s.Helper.ParseDateFromFilter(filter)
	sqlGet := `SELECT substr(ratings.created_at, 1, 10) as SRAD,
	rating_categories.name as Category,
	ROUND((((ratings.rating * rating_categories.weight)*100)/(` + fmt.Sprintf("%v", s.Weight) + `))) as rtg
	from ratings 
	LEFT JOIN rating_categories ON 
	ratings.rating_category_id = rating_categories.id 
	WHERE (
		ratings.created_at BETWEEN DATE("` + from + `") 
		AND DATE("` + to + `")
		AND rating_categories.weight > 0 
		AND rtg NOT NULL
		AND ratings.rating > 0
		)
	GROUP BY Category, ratings.created_at
	ORDER BY ratings.created_at`
	rows, err := s.Database.Query(sqlGet)
	if err != nil {
		log.Println("here")
		return err
	}
	defer rows.Close()
	var prevStrDate string = ""
	//GO maps are hashmaps so the lookup time is O(1)
	var dailyResult = make(map[string]map[string]int32)
	var dailyCount = make(map[string]int)
	for rows.Next() {
		var SRAD sql.NullString
		var Category sql.NullString
		var rating sql.NullInt32
		e := rows.Scan(&SRAD, &Category, &rating)
		if e != nil {
			log.Fatal(e)
		}
		categoryString := Category.String
		formattedDate := SRAD.String
		if prevStrDate == "" {
			prevStrDate = formattedDate
		}
		if dailyResult[formattedDate] == nil {
			dailyResult[formattedDate] = make(map[string]int32)
		}
		g := dailyResult[formattedDate]
		if g[categoryString] > 0 {
			g[categoryString] += rating.Int32
			dailyCount[categoryString]++
		} else {
			g[categoryString] = rating.Int32
			if prevStrDate == formattedDate {
				dailyCount[categoryString] = 1
			}
		}
		if prevStrDate != formattedDate && len(dailyResult[prevStrDate]) > 0 {
			//Stream data back
			//Total in range is used to divide the dailyResult and calculate a precentage
			cat := make([]*service.CategoryResult, 0)
			//log.Println(dailyResult[srad])
			for k, v := range dailyResult[prevStrDate] {
				tot := int32(dailyCount[k])
				calc := int32(v / tot)
				r := service.CategoryResult{
					CategoryName: k,
					Ratings:      tot,
					Score:        calc,
				}
				cat = append(cat, &r)
			}
			date := s.Helper.GenerateClientDate(prevStrDate)
			result := service.Categories{
				Result: cat,
				Date:   date,
			}
			delete(dailyResult, prevStrDate)
			if err := stream.Send(&result); err != nil {
				return err
			}
			dailyCount = make(map[string]int)
		}
		prevStrDate = formattedDate
	}
	return nil
}

func (s *server) GetScoresByTickets(filter *service.DateRange, stream service.TicketService_GetScoresByTicketsServer) error {
	from, to := s.Helper.ParseDateFromFilter(filter)
	sqlGet := `SELECT rating_categories.name as Category,
	tickets.id AS TKTID,
	ROUND((((ratings.rating * rating_categories.weight)*100)/(rating_categories.weight*5))) as SCORE	
	from ratings
	LEFT JOIN rating_categories ON 
	ratings.rating_category_id = rating_categories.id
	LEFT JOIN tickets ON ratings.ticket_id = tickets.id
		WHERE ( 
			ratings.created_at BETWEEN DATE("` + from + `") 
			AND DATE("` + to + `")
			AND rating_categories.weight > 0 
			AND SCORE NOT NULL
			AND ratings.rating > 0
		)`
	rows, err := s.Database.Query(sqlGet)
	if err != nil {
		log.Println("here")
		return err
	}
	defer rows.Close()
	var prevTicketID int32 = 0
	var ticketResult = make(map[int32]map[string]int32)
	for rows.Next() {
		var sc sql.NullInt32
		var tkid sql.NullInt32
		var ctg sql.NullString
		e := rows.Scan(&ctg, &tkid, &sc)
		if e != nil {
			log.Fatal(e)
		}
		category := ctg.String
		ticketID := tkid.Int32
		score := sc.Int32
		if prevTicketID == 0 {
			prevTicketID = ticketID
		}
		if ticketResult[ticketID] == nil {
			ticketResult[ticketID] = make(map[string]int32)
		}
		//Add score precentage to the category
		ticketResult[ticketID][category] = score
		if prevTicketID != ticketID {
			res := service.TicketScores{
				Id:     prevTicketID,
				Result: make([]*service.CategoryResult, 0),
			}
			for k, v := range ticketResult[prevTicketID] {
				ctRes := &service.CategoryResult{
					Score:        v,
					CategoryName: k,
				}
				res.Result = append(res.Result, ctRes)
			}
			delete(ticketResult, prevTicketID)
			if err := stream.Send(&res); err != nil {
				log.Println(err)
				return err
			}
		}
		prevTicketID = ticketID
	}
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
	take the rating value and multiply it with it's weight, then divide the result by the
	weights with maximum values which is 5 in our case.
	finally return an int32 as precentage

func calculateRatingPrecentage(totalWeight int, ratings map[int]Rating) int {
	var finalRatingPrecentage = 0
	for _, r := range ratings {
		fv := r.Value
		var ratingVal = int(r.Weight * float64(fv))
		finalRatingPrecentage += (ratingVal * 100) / int(totalWeight*5)
	}
	return finalRatingPrecentage
}
*/

func dateToString(m int32, d int32, y int32) string {
	var res string
	res += fmt.Sprintf("%v-", y)
	res += fmt.Sprintf("%v-", m)
	res += fmt.Sprintf("%v", d)
	return res
}
