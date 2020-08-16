package calculator

import (
	"reflect"
	"strconv"
	"testing"
)

func TestExpectedOutput(t *testing.T) {

	var r RecipeStatsCalculator

	postcodeDeliveryTimeFilter := PostcodeDeliveryTimeFilter{
		Postcode: "10120",
		FromAM:   10,
		ToPM:     3,
	}
	expectedOutput := r.CalculateStats("../../resources/hf_test_calculation_fixtures_SMALL.json",
		postcodeDeliveryTimeFilter,
		[]string{"Potato", "Veggie", "Mushroom"},
	)

	// it should return instance of ExpectedOutput
	if reflect.TypeOf(expectedOutput).Name() != "ExpectedOutput" {
		t.Errorf("got %s, want %s", reflect.TypeOf(expectedOutput).Name(), "ExpectedOutput")
	}

	// 1. Count the number of unique recipe names
	if expectedOutput.UniqueRecipeCount != 4 {
		t.Errorf("got %d, want %d", expectedOutput.UniqueRecipeCount, 4)
	}

	// 2. Count the number of occurrences for each unique recipe name (alphabetically ordered by recipe name)
	countPerRecipe := CountPerRecipe{
		"A5 Balsamic Veggie Chops",
		1,
	}

	// checking only first recipe
	if expectedOutput.SortedRecipeCountData[0] != countPerRecipe {
		t.Errorf("got %v, want %v", expectedOutput.SortedRecipeCountData[0], countPerRecipe)
	}

	// 3. Find the postcode with most delivered recipes.
	busiestPostcode := BusiestPostcode{
		Postcode:      "10120",
		DeliveryCount: 3,
	}

	if expectedOutput.BusiestPostcode != busiestPostcode {
		t.Errorf("got %v, want %v", expectedOutput.BusiestPostcode, busiestPostcode)
	}

	// 4. Count the number of deliveries to postcode `10120` that lie within the delivery time between `10AM` and `3PM`
	countPerPostcodeAndTime := CountPerPostcodeAndTime{
		Postcode:      postcodeDeliveryTimeFilter.Postcode,
		FromAM:        strconv.Itoa(postcodeDeliveryTimeFilter.FromAM) + "AM",
		ToPM:          strconv.Itoa(postcodeDeliveryTimeFilter.ToPM) + "PM",
		DeliveryCount: 2,
	}

	if expectedOutput.CountPerPostcodeAndTime != countPerPostcodeAndTime {
		t.Errorf("got %v, want %v", expectedOutput.CountPerPostcodeAndTime, countPerPostcodeAndTime)
	}

	// 5.  List the recipe names (alphabetically ordered) that contain in their name one of the following words:
	if len(expectedOutput.SortedRecipeNames) != 3 {
		t.Errorf("got %v, want %v", len(expectedOutput.SortedRecipeNames), 3)
	}
}
