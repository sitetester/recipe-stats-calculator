package main

import "github.com/sitetester/recipe-stats-calculator/service/calculator"

func main() {

	var r calculator.RecipeStatsCalculator
	r.CalculateStats("./tmp/hf_test_calculation_fixtures_SMALL.json",
		calculator.PostcodeDeliveryTimeFilter{
			Postcode: "10120",
			FromAM:   10,
			ToPM:     3,
		},
		[]string{"Potato", "Veggie", "Mushroom"},
	)

}
