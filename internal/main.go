package main

import (
	"context"
	"database/sql"
	"errors"
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

const (
	dbErrorString = "Error while querying database: "
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
		panic(serr)
	}
}

func (s *server) GetAggregatedCategory(filter *service.DateRange, stream service.TicketService_GetAggregatedCategoryServer) error {
	from, to, isWeekly := s.Helper.ParseDateFromFilter(filter)
	selectDateRange := `DATE(ratings.created_at, "weekday 0")`
	if !isWeekly {
		selectDateRange = "substr(ratings.created_at, 1, 10)"
	}
	sqlGet := `SELECT ` + selectDateRange + ` AS SRAD,
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
		log.Printf("%v: %v", dbErrorString, err.Error())
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
			return e
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
			/*
				Stream data back
				Total in range is used to divide the dailyResult and calculate a precentage
				Date generation is multithreaded, because string manipulation is CPU heavy
			*/
			dch := make(chan *service.Period)
			go s.Helper.GenerateClientDate(prevStrDate, dch)
			cat := make([]*service.CategoryResult, 0)
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
			result := &service.Categories{
				Result: cat,
				Date:   <-dch,
			}
			//Clear up memory
			delete(dailyResult, prevStrDate)
			if err := stream.Send(result); err != nil {
				return err
			}
			dailyCount = make(map[string]int)
		}
		prevStrDate = formattedDate
	}
	return errors.New("Database is empty or entered date range is invalid")
}

func (s *server) GetScoresByTickets(filter *service.DateRange, stream service.TicketService_GetScoresByTicketsServer) error {
	from, to, _ := s.Helper.ParseDateFromFilter(filter)
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
		log.Printf("%v: %v", dbErrorString, err.Error())
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
			return e
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
	from, to, _ := s.Helper.ParseDateFromFilter(request)
	fmt.Println(from)
	fmt.Println("========")
	fmt.Println(to)
	sqlGet := `
	SELECT COUNT(*) as TCount, SUM(rtg) as TSum FROM (
		SELECT (((ratings.rating * rating_categories.weight)*100)/(` + fmt.Sprintf("%v", s.Weight) + `)) as rtg
		from ratings
		LEFT JOIN rating_categories ON ratings.rating_category_id = rating_categories.id
		WHERE (
			ratings.created_at BETWEEN DATE("` + fmt.Sprintf("%v", from) + `") 
			AND DATE("` + fmt.Sprintf("%v", to) + `")
			AND rating_categories.weight > 0 
			AND rtg NOT NULL
			AND ratings.rating > 0
		)
	)`
	rows, err := s.Database.Query(sqlGet)
	if err != nil {
		log.Printf("%v: %v", dbErrorString, err.Error())
		return nil, err
	}
	defer rows.Close()
	var result *service.Quality
	for rows.Next() {
		var count sql.NullInt32
		var sum sql.NullFloat64
		rows.Scan(&count, &sum)
		s := sum.Float64
		c := count.Int32
		if c == 0 || s == 0 {
			return nil, errors.New("Missing values in database or invalid date range")
		}
		p := (int32(s) / c)
		log.Println()
		result = &service.Quality{
			Precentage: p,
		}
	}
	return result, nil
}

func (s *server) GetPeriodOverPeriod(ctx context.Context, request *service.DateRange) (*service.PeriodChange, error) {
	return nil, nil
}
