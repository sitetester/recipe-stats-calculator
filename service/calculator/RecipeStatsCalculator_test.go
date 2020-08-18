package calculator

import (
	"reflect"
	"strconv"
	"testing"
)

func TestExpectedOutput(t *testing.T) {

	var r RecipeStatsCalculator

	postcodeDeliveryTimeFilter := CustomPostcodeDeliveryTime{
		Postcode: "10120",
		From:     10,
		To:       3,
	}
	expectedOutput := r.CalculateStats("../../resources/hf_test_calculation_fixtures_SMALL.json",
		postcodeDeliveryTimeFilter,
		[]string{"Potato", "Veggie", "Mushroom"},
	)

	// expectedOutput must be instance of ExpectedOutput
	t.Run("instanceCheck", func(t *testing.T) {
		if reflect.TypeOf(expectedOutput).Name() != "ExpectedOutput" {
			t.Errorf("got %s, want %s", reflect.TypeOf(expectedOutput).Name(), "ExpectedOutput")
		}
	})

	// 1. Count the number of unique recipe names
	t.Run("UniqueRecipeCount", func(t *testing.T) {
		if expectedOutput.UniqueRecipeCount != 4 {
			t.Errorf("got %d, want %d", expectedOutput.UniqueRecipeCount, 4)
		}
	})

	// 2. Count the number of occurrences for each unique recipe name (alphabetically ordered by recipe name)
	t.Run("UniqueRecipeCount", func(t *testing.T) {
		countPerRecipe := CountPerRecipe{
			"A5 Balsamic Veggie Chops",
			1,
		}

		// checking only first recipe
		if expectedOutput.SortedRecipesCount[0] != countPerRecipe {
			t.Errorf("got %v, want %v", expectedOutput.SortedRecipesCount[0], countPerRecipe)
		}
	})

	// 3. Find the postcode with most delivered recipes.
	t.Run("UniqueRecipeCount", func(t *testing.T) {
		busiestPostcode := BusiestPostcode{
			Postcode:      "10120",
			DeliveryCount: 3,
		}

		if expectedOutput.BusiestPostcode != busiestPostcode {
			t.Errorf("got %v, want %v", expectedOutput.BusiestPostcode, busiestPostcode)
		}
	})

	// 4. Count the number of deliveries to postcode `10120` that lie within the delivery time between `10AM` and `3PM`
	t.Run("UniqueRecipeCount", func(t *testing.T) {
		countPerPostcodeAndTime := CountPerPostcodeAndTime{
			Postcode:      postcodeDeliveryTimeFilter.Postcode,
			FromAM:        strconv.Itoa(postcodeDeliveryTimeFilter.From) + "AM",
			ToPM:          strconv.Itoa(postcodeDeliveryTimeFilter.To) + "PM",
			DeliveryCount: 2,
		}

		if expectedOutput.CountPerPostcodeAndTime != countPerPostcodeAndTime {
			t.Errorf("got %v, want %v", expectedOutput.CountPerPostcodeAndTime, countPerPostcodeAndTime)
		}
	})

	// 5.  List the recipe names (alphabetically ordered) that contain in their name one of the following words:
	t.Run("UniqueRecipeCount", func(t *testing.T) {
		if len(expectedOutput.SortedRecipeNames) != 3 {
			t.Errorf("got %v, want %v", len(expectedOutput.SortedRecipeNames), 3)
		}
	})
}
