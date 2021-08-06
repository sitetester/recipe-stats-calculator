package main

import (
	"encoding/json"
	"fmt"
	"github.com/sitetester/recipe-stats-calculator/service/calculator"
	"time"
)

func main() {
	start := time.Now()
	var calc calculator.RecipeStatsCalculator

	expectedOutput := calc.CalculateStats("./resources/hf_test_calculation_fixtures_SMALL.json",
		calculator.CustomPostcodeDeliveryTime{
			Postcode: "10224",
			From:     10,
			To:       3,
		},
		[]string{"Potato", "Veggie", "Mushroom"},
	)

	println(prettyPrint(expectedOutput))

	duration := time.Since(start)
	fmt.Println(fmt.Sprintf("%s: %f", "Time took in seconds", duration.Seconds()))
}

func prettyPrint(i interface{}) string {

	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
