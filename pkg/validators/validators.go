package validators

import "fireynis/velocity_checker/pkg/models"

type ILoadValidator interface {
	LessThanThreeLoadsDaily([]*models.Load) bool
	LessThanFiveThousandLoadedDaily([]*models.Load, *models.Load) bool
	LessThanTwentyThousandLoadedWeekly([]*models.Load, *models.Load) bool
}
