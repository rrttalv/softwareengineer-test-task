package helper

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	service "github.com/softwareengineer-test-task/proto"
)

type Helper struct {
}

func (h *Helper) ParseDateFromFilter(f *service.DateRange) (string, string) {
	monF, monT, dayF, dayT := f.PeriodFrom.GetMonth(), f.PeriodTo.GetMonth(), f.PeriodFrom.GetDay(), f.PeriodTo.GetDay()+1
	smonF, smonT, sdayF, sdayT := h.ParseDate(monF), h.ParseDate(monT), h.ParseDate(dayF), h.ParseDate(dayT)
	return fmt.Sprintf("%v-%v-%v", f.PeriodFrom.GetYear(), smonF, sdayF), fmt.Sprintf("%v-%v-%v", f.PeriodTo.GetYear(), smonT, sdayT)
}

func (h *Helper) ParseDate(date int32) string {
	if date < 10 {
		return fmt.Sprintf("0%v", date)
	}
	return fmt.Sprintf("%v", date)
}

func (h *Helper) GenerateClientDate(date string) *service.Period {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	da := strings.FieldsFunc(date, f)
	//We dont need to check for errors because we know the DB field structure
	//Still should not be done in production though..
	year, _ := strconv.Atoi(da[0])
	month, _ := strconv.Atoi(da[1])
	day, _ := strconv.Atoi(da[2])
	return &service.Period{
		Year:  int32(year),
		Day:   int32(day),
		Month: int32(month),
	}
}
