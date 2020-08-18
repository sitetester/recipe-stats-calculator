package main

import (
	json2 "encoding/json"
	"github.com/sitetester/recipe-stats-calculator/service/calculator"
)

func main() {

	var calc calculator.RecipeStatsCalculator

	expectedOutput := calc.CalculateStats("./resources/hf_test_calculation_fixtures_SMALL.json",
		calculator.CustomPostcodeDeliveryTime{
			Postcode: "10120",
			From:     10,
			To:       3,
		},
		[]string{"Potato", "Veggie", "Mushroom"},
	)

	println(prettyPrint(expectedOutput))
}

func prettyPrint(i interface{}) string {

	s, _ := json2.MarshalIndent(i, "", "\t")

	return string(s)
}
