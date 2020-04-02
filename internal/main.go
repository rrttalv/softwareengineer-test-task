package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
	_ "github.com/mattn/go-sqlite3"
	service "github.com/softwareengineer-test-task/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	Database *sql.DB
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
	database, dberr := sql.Open("sqlite3", "database.db")
	if dberr != nil {
		log.Println("Can't connect to database!")
		panic(err)
	}
	server := &server{
		Database: database,
	}
	_, ser := database.Exec(`DROP TABLE IF EXISTS vartable`)
	if ser != nil {
		panic(ser)
	}
	_, cerr := database.Exec(`CREATE TABLE vartable (name TEXT PRIMARY KEY, val INT)`)
	if cerr != nil {
		panic(cerr)
	}
	srv := grpc.NewServer()
	service.RegisterTicketServiceServer(srv, server)
	reflection.Register(srv)
	if serr := srv.Serve(listener); serr != nil {
		panic(err)
	}
}

func parseDate(date int32) string {
	if date < 10 {
		return fmt.Sprintf("0%v", date)
	}
	return fmt.Sprintf("%v", date)
}

func (s *server) GetAggregatedCategory(filter *service.DateRange, stream service.TicketService_GetAggregatedCategoryServer) error {
	_, ierr := s.Database.Exec(`INSERT OR REPLACE INTO vartable SELECT 'tw' AS name, sum(rating_categories.weight*5) AS val FROM rating_categories`)
	if ierr != nil {
		log.Println("here!!")
		return ierr
	}
	monF, monT, dayF, dayT := filter.PeriodFrom.GetMonth(), filter.PeriodTo.GetMonth(), filter.PeriodFrom.GetDay(), filter.PeriodTo.GetDay()
	smonF, smonT, sdayF, sdayT := parseDate(monF), parseDate(monT), parseDate(dayF), parseDate(dayT)
	sqlGet := `SELECT ratings.id as rid,
	ratings.created_at as RAD,
	substr(ratings.created_at, 1, 10) as SRAD,
	rating_categories.name as Category,
	ROUND((((ratings.rating * rating_categories.weight)*100)/(
		SELECT val FROM vartable WHERE name = "tw"
	))) as rtg,
	ratings.rating,
	rating_categories.weight
	from ratings 
	LEFT JOIN rating_categories ON 
	ratings.rating_category_id = rating_categories.id 
	WHERE (
		RAD BETWEEN DATE("` + fmt.Sprintf("%v-%v-%v", filter.PeriodFrom.GetYear(), smonF, sdayF) + `") 
		AND DATE("` + fmt.Sprintf("%v-%v-%v", filter.PeriodTo.GetYear(), smonT, sdayT) + `")
		AND rating_categories.weight > 0 
		AND rtg NOT NULL
		AND ratings.rating > 0
		)
	GROUP BY Category, RAD
	ORDER BY RAD`
	rows, err := s.Database.Query(sqlGet)
	if err != nil {
		log.Println("here")
		return err
	}
	var prevStrDate string = ""
	var totalInRange = 0
	cols, _ := rows.Columns()
	log.Println(cols)
	var dailyResult = make(map[string]map[string]int32)
	for rows.Next() {
		var SRAD sql.NullString
		var Category sql.NullString
		var rtg sql.NullInt32
		var RAD sql.NullTime
		var rid sql.NullInt32
		var rating sql.NullInt32
		var weight sql.NullFloat64
		e := rows.Scan(&rid, &RAD, &SRAD, &Category, &rtg, &rating, &weight)
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
			g[categoryString] += rtg.Int32
		} else {
			g[categoryString] = rtg.Int32
		}
		fmt.Println(dailyResult[prevStrDate])
		if prevStrDate != formattedDate && len(dailyResult[prevStrDate]) > 0 {
			fmt.Println(prevStrDate)
			fmt.Println("=======")
			log.Println(dailyResult[prevStrDate])
			//Stream data back
			//Total in range is used to divide the dailyResult and calucalte a precentage
			cat := make([]*service.CategoryResult, 0)
			//log.Println(dailyResult[srad])
			for k, v := range dailyResult[prevStrDate] {
				tot := int32(totalInRange)
				calc := int32(v / tot)
				r := service.CategoryResult{
					CategoryName: k,
					Rating:       tot,
					Score:        calc,
				}
				cat = append(cat, &r)
			}
			result := service.Categories{
				Result: cat,
			}
			if err := stream.Send(&result); err != nil {
				return err
			}
			totalInRange = 0
		}
		totalInRange++
		prevStrDate = formattedDate
	}
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
	take the rating value and multiply it with it's weight, then divide the result by the
	weights with maximum values which is 5 in our case.
	finally return an int32 as precentage
*/
func calculateRatingPrecentage(totalWeight int, ratings map[int]Rating) int {
	var finalRatingPrecentage = 0
	for _, r := range ratings {
		fv := r.Value
		var ratingVal = int(r.Weight * float64(fv))
		finalRatingPrecentage += (ratingVal * 100) / int(totalWeight*5)
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
