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

func (h *Helper) ParseDateFromFilter(f *service.DateRange) (string, string, bool) {
	monF, monT, dayF, dayT, yearF, yearT := f.PeriodFrom.GetMonth(), f.PeriodTo.GetMonth(), f.PeriodFrom.GetDay(),
		f.PeriodTo.GetDay()+1, f.PeriodFrom.GetYear(), f.PeriodTo.GetYear()
	smonF, smonT, sdayF, sdayT := h.ParseDate(monF), h.ParseDate(monT), h.ParseDate(dayF), h.ParseDate(dayT)
	pfrom, pto := fmt.Sprintf("%v-%v-%v", yearF, smonF, sdayF), fmt.Sprintf("%v-%v-%v", yearT, smonT, sdayT)
	if yearT > yearF || monT-monF >= 2 {
		return pfrom, pto, true
	}
	/*
		If monF minus month to is equal to 1 and the day range is greater than zero
		then weekly is true
	*/
	if monT-monF == 1 && dayT-dayF > 0 {
		return pfrom, pto, true
	}
	return pfrom, pto, false
}

func (h *Helper) ParseDate(date int32) string {
	if date < 10 {
		return fmt.Sprintf("0%v", date)
	}
	return fmt.Sprintf("%v", date)
}

func (h *Helper) GenerateClientDate(date string, ch chan *service.Period) {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	da := strings.FieldsFunc(date, f)
	//We dont need to check for errors because we know the DB field structure
	//Still should not be done in production though..
	year, _ := strconv.Atoi(da[0])
	month, _ := strconv.Atoi(da[1])
	day, _ := strconv.Atoi(da[2])
	ch <- &service.Period{
		Day:   int32(day),
		Month: int32(month),
		Year:  int32(year),
	}
}
