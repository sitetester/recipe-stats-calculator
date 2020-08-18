package calculator

import (
	"reflect"
	"strconv"
	"testing"
)

func TestExpectedOutput(t *testing.T) {

	var r RecipeStatsCalculator

	customPostcodeDeliveryTime := CustomPostcodeDeliveryTime{
		Postcode: "10120",
		From:     10,
		To:       3,
	}
	expectedOutput := r.CalculateStats("../../resources/hf_test_calculation_fixtures_SMALL.json",
		customPostcodeDeliveryTime,
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
	t.Run("SortedRecipesCount", func(t *testing.T) {
		countPerRecipeAtIndex0 := CountPerRecipe{
			"A5 Balsamic Veggie Chops",
			1,
		}

		// first recipe in sorted order
		if expectedOutput.SortedRecipesCount[0] != countPerRecipeAtIndex0 {
			t.Errorf("got %v, want %v", expectedOutput.SortedRecipesCount[0], countPerRecipeAtIndex0)
		}

		// it should cover all recipes
		if len(expectedOutput.SortedRecipesCount) != 5 {
			t.Errorf("got %v, want %v", len(expectedOutput.SortedRecipesCount), 5)
		}

		// `Creamy Dill Chicken` has two counts
		countPerRecipeAtIndex3 := CountPerRecipe{
			"Creamy Dill Chicken",
			2,
		}

		if expectedOutput.SortedRecipesCount[3] != countPerRecipeAtIndex3 {
			t.Errorf("got %v, want %v", expectedOutput.SortedRecipesCount[3], countPerRecipeAtIndex3)
		}
	})

	// 3. Find the postcode with most delivered recipes.
	t.Run("BusiestPostcode", func(t *testing.T) {
		busiestPostcode := BusiestPostcode{
			Postcode:      "10120",
			DeliveryCount: 3,
		}

		if expectedOutput.BusiestPostcode != busiestPostcode {
			t.Errorf("got %v, want %v", expectedOutput.BusiestPostcode, busiestPostcode)
		}
	})

	// 4. Count the number of deliveries to postcode `10120` that lie within the delivery time between `10AM` and `3PM`
	t.Run("CountPerPostcodeAndTime", func(t *testing.T) {
		countPerPostcodeAndTime := CountPerPostcodeAndTime{
			Postcode:      customPostcodeDeliveryTime.Postcode,
			FromAM:        strconv.Itoa(customPostcodeDeliveryTime.From) + "AM",
			ToPM:          strconv.Itoa(customPostcodeDeliveryTime.To) + "PM",
			DeliveryCount: 2,
		}

		if expectedOutput.CountPerPostcodeAndTime != countPerPostcodeAndTime {
			t.Errorf("got %v, want %v", expectedOutput.CountPerPostcodeAndTime, countPerPostcodeAndTime)
		}
	})

	// 5.  List the recipe names (alphabetically ordered) that contain in their name one of the following words:
	t.Run("SortedRecipeNames", func(t *testing.T) {
		if len(expectedOutput.SortedRecipeNames) != 3 {
			t.Errorf("got %v, want %v", len(expectedOutput.SortedRecipeNames), 3)
		}
	})
}
