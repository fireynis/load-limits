package validators

import (
	"fireynis/velocity_checker/pkg/models"
)

type LoadValidator struct{}

//LessThanThreeLoadsDaily takes in an array of models that should all be from the same day. Basically it gets the
//len of the models passed in
func (l *LoadValidator) LessThanThreeLoadsDaily(loads []*models.Load) bool {
	if len(loads) >= 3 {
		return false
	}
	return true
}

//LessThanFiveThousandLoadedDaily sums the accepted loads to determine if they exceed the daily 5k limit
func (l *LoadValidator) LessThanFiveThousandLoadedDaily(loads []*models.Load, load *models.Load) bool {
	return l.sumLessThanMax(loads, load, int64(500000))
}

//LessThanTwentyThousandLoadedWeekly sums the accepted loads
func (l *LoadValidator) LessThanTwentyThousandLoadedWeekly(loads []*models.Load, load *models.Load) bool {
	return l.sumLessThanMax(loads, load, int64(2000000))
}

func (l *LoadValidator) sumLessThanMax(loads []*models.Load, load *models.Load, maxAmount int64) bool {
	//Short circuit if the cur value is higher than the daily limit. Don't need to waste the computation.
	if load.Amount > maxAmount {
		return false
	}
	if load.Amount > maxAmount {
		return false
	}
	amount := int64(0)
	for _, load := range loads {
		amount += load.Amount
	}
	amount += load.Amount
	if amount > maxAmount {
		return false
	}
	return true
}
