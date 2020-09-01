package validators

import (
	"fireynis/velocity_checker/pkg/models"
)

type LoadValidator struct{}

//LessThanThreeLoadsDaily takes in an array of models that should all be from the same day
func (l *LoadValidator) LessThanThreeLoadsDaily(loads []*models.Load) bool {
	if len(loads) >= 3 {
		return false
	}
	return true
}

func (l *LoadValidator) LessThanFiveThousandLoadedDaily(loads []*models.Load, load *models.Load) bool {
	amount := int64(0)
	for _, load := range loads {
		amount += load.Amount
	}
	amount += load.Amount
	if amount > 500000 {
		return false
	}
	return true
}

func (l *LoadValidator) LessThanTwentyThousandLoadedWeekly(loads []*models.Load, load *models.Load) bool {
	amount := int64(0)
	for _, load := range loads {
		amount += load.Amount
	}
	amount += load.Amount
	if amount > 2000000 {
		return false
	}
	return true
}
