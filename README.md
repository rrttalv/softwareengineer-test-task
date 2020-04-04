## Comments and thoughts

**Database query ideas and explanations in ./dbquerys.sql file**

#### Aggregate rating algorithm explained

Used the following formula: 
    ((x * y)/100)/z)
* x = Total rating which can range from 0 - 5
* y = The rating's weight in rating categories table
* z = Sum of all rating category weights
This algorithm was applied on the database level

#### Results

For testing in postman:

1. Aggregated category scores:
    yields a years worth of results in <200ms which is pretty good
    **URL**: http://localhost:8080/aggregated
    **JSON BODY**:
    ```json{
        "period_from": {"day": 25, "month": 2, "year": 2019},
        "period_to": {"day": 26, "month": 2, "year": 2020}
    }```
2. Scores by ticket:
    **URL**: http://localhost:8080/byticket
    **JSON BODY**:
    ```json{
        "period_from": {"day": 25, "month": 2, "year": 2019},
        "period_to": {"day": 26, "month": 2, "year": 2020}
    }```
3. Overal quality score:
    ***note***: Result almost always 20
    **URL**: http://localhost:8080/quality
    **JSON BODY**:
    ```json
    {
        "period_from": {"day": 25, "month": 2, "year": 2019},
        "period_to": {"day": 26, "month": 2, "year": 2020}
    }```
4. Period over period query:
    **note**: Not sure if it's the data, but 99% of the time the result is empty which means my algorithm yielded 0 as the result
    **URL**: http://localhost:8080/period
    **JSON BODY**:
    ```json{
    "selected_period": {
        "period_from": {"day": 1, "month": 2, "year": 2019},
        "period_to": {"day": 31, "month": 9, "year": 2019}
        },
    "previous_period": {
        "period_from": {"day": 1, "month": 10, "year": 2019},
        "period_to": {"day": 25, "month": 2, "year": 2020}
        }
    }```

# Software Engineer Test Task

As a test task for [Klaus](https://www.klausapp.com) software engineering position we ask our candidates to build a small [gRPC](https://grpc.io) service using language of their choice. Prefered language for new services in Klaus is [Go](https://golang.org).

The service should be using provided sample data from SQLite database (`database.db`).

Please fork this repository and share the link to your solution with us.

### Tasks

1. Come up with ticket score algorithm that accounts for rating category weights (available in `rating_categories` table). Ratings are given in a scale of 0 to 5. Score should be representable in percentages from 0 to 100. 

2. Build a service that can be queried using [gRPC](https://grpc.io/docs/tutorials/basic/go/) calls and can answer following questions:

    * **Aggregated category scores over a period of time**
    
        E.g. what have the daily ticket scores been for a past week or what were the scores between 1st and 31st of January.

        For periods longer than one month weekly aggregates should be returned instead of daily values.

        From the reponse the following UI representation should be possible:

        | Category | Ratings | Date 1 | Date 2 | ... | Score |
        |----|----|----|----|----|----|
        | Tone | 1 | 30% | N/A | N/A | X% |
        | Grammar | 2 | N/A | 90% | 100% | X% |
        | Random | 6 | 12% | 10% | 10% | X% |

    * **Scores by ticket**

        Aggregate scores for categories within defined period by ticket.

        E.g. what aggregate category scores tickets have within defined rating time range have.

        | Ticket ID | Category 1 | Category 2 |
        |----|----|----|
        | 1   |  100%  |  30%  |
        | 2   |  30%  |  80%  |

    * **Overal quality score**

        What is the overall aggregate score for a period.

        E.g. the overall score over past week has been 96%.

    * **Period over Period score change**

        What has been the change from selected period over previous period.

        E.g. current week vs. previous week or December vs. January change in percentages.


### Bonus

* How would you build and deploy the solution?

    At Klaus we make heavy use of containers and [Kubernetes](https://kubernetes.io).