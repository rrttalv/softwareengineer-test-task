/*1. GET TOTAL RATINGS FOR PERIOD
Using substring for date parsing, because it is faster than the strftime function
The variable here is SRAD, the value of SRAD can also be: DATE(ratings.created_at, "weekday 0")
if grouped by weeks
*/
SELECT substr(ratings.created_at, 1, 10) as SRAD,
	rating_categories.name as Category,
	ROUND((((ratings.rating * rating_categories.weight)*100)/(14.5))) as rtg
	from ratings 
	LEFT JOIN rating_categories ON 
	ratings.rating_category_id = rating_categories.id 
	WHERE (
		ratings.created_at BETWEEN DATE("2019-01-25") 
		AND DATE("2020-05-27")
		AND rating_categories.weight > 0 
		AND rtg NOT NULL
		AND ratings.rating > 0
		)
	GROUP BY Category, ratings.created_at
	ORDER BY ratings.created_at

/*
	2. Get ratings by ticket within time range

	I user the rating created at for the time range constraint.
	The constraint could be replaced with tickets.created_at to get ratings for tickets
	in the defined time range. Wording in the README was cluncky.

	Secondly, I removed "ORDER BY TKTID" from the statement. For some reason the
	statement slowed down the query. Funnily enough SQLITE returns IDs in groups which allowed me to
	build an algorithm which does not use much RAM and goes through the tickets in order.
	Ref - internal/main.go: func GetScoresByTickets..
*/
SELECT rating_categories.name as Category,
tickets.id AS TKTID,
ROUND((((ratings.rating * rating_categories.weight)*100)/(rating_categories.weight*5))) as SCORE	
from ratings
LEFT JOIN rating_categories ON 
ratings.rating_category_id = rating_categories.id
LEFT JOIN tickets ON ratings.ticket_id = tickets.id
	WHERE ( 
		ratings.created_at BETWEEN DATE("2019-02-25") 
		AND DATE("2020-02-26")
		AND rating_categories.weight > 0 
		AND SCORE NOT NULL
		AND ratings.rating > 0
		AND rating_categories.name = "GDPR"
	)

/*
	3. Overal quality score select

	Simple Count and Sum select which returns a single row.

	Weirdly enough if I do TSum/TCount the result almost always = 20 +-1
	I hope it's just the data :D
*/
SELECT COUNT(*) as TCount, SUM(rtg) as TSum FROM (
	SELECT (((ratings.rating * rating_categories.weight)*100)/(14.5)) as rtg
	from ratings
	LEFT JOIN rating_categories ON ratings.rating_category_id = rating_categories.id
	WHERE (
		ratings.created_at BETWEEN DATE("2019-03-01") 
		AND DATE("2019-03-31")
		AND rating_categories.weight > 0 
		AND rtg NOT NULL
		AND ratings.rating > 0
	)
)

/*

	4. Period over period change

	Pretty much the same query as before, just executed twice.

	Almost always the answer is 0 and result is empty if querying through the client.

*/

SELECT COUNT(*) as TCount, SUM(rtg) as TSum FROM (
		SELECT (((ratings.rating * rating_categories.weight)*100)/(14.5)) as rtg
		from ratings
		LEFT JOIN rating_categories ON ratings.rating_category_id = rating_categories.id
		WHERE (
			ratings.created_at BETWEEN DATE("2019-03-01") 
			AND DATE("2019-03-02")
			AND rating_categories.weight > 0 
			AND rtg NOT NULL
			AND ratings.rating > 0
		)
	)