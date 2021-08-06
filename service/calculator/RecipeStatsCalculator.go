package calculator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type RecipeStatsCalculator struct {
	customPostcodeDeliveryTime CustomPostcodeDeliveryTime
	customRecipeNames          []string
}

type CustomPostcodeDeliveryTime struct {
	Postcode string
	From     int
	To       int
}

type ExpectedOutput struct {
	UniqueRecipeCount       int                     `json:"unique_recipe_count"`
	SortedRecipesCount      []CountPerRecipe        `json:"count_per_recipe"`
	BusiestPostcode         BusiestPostcode         `json:"busiest_postcode"`
	CountPerPostcodeAndTime CountPerPostcodeAndTime `json:"count_per_postcode_and_time"`
	SortedRecipeNames       []string                `json:"match_by_name"`
	TotalObjects			int64					 `json:"total_json_objects"`
}

type BusiestPostcode struct {
	Postcode      string `json:"postcode"`
	DeliveryCount int    `json:"delivery_count"`
}

type CountPerRecipe struct {
	Recipe string `json:"recipe"`
	Count  int    `json:"count"`
}

type RecipeData struct {
	Postcode string
	Recipe   string
	Delivery string
}

type CountPerPostcodeAndTime struct {
	Postcode      string `json:"postcode"`
	FromAM        string `json:"from"`
	ToPM          string `json:"to"`
	DeliveryCount int    `json:"delivery_count"`
}

type CountPerPostcode struct {
	Postcode string
	Count    int
}

func toLower(customRecipeNames []string) []string {

	lowerCased := make([]string, 0, len(customRecipeNames))

	for _, v := range customRecipeNames {
		lowerCased = append(lowerCased, strings.ToLower(v))
	}

	return lowerCased
}

// filter criteria is passed as params, so that we couldn't miss it
func (calc *RecipeStatsCalculator) CalculateStats(
	filePath string,
	customPostcodeDeliveryTime CustomPostcodeDeliveryTime,
	customRecipeNames []string) ExpectedOutput {

	calc.customPostcodeDeliveryTime = customPostcodeDeliveryTime
	calc.customRecipeNames = toLower(customRecipeNames)

	file, err := os.Open(filePath)
	if err != nil {
		toStdErr(err)
		return ExpectedOutput{}
	}
	defer closeFile(file)

	countPerRecipe := make(map[string]int)
	countPerPostcode := make(map[string]int)
	deliveriesCountPerPostcode := make(map[string]int)
	var filteredRecipeNames []string

	totalObjects := 0
	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)

	decoder.Token()
	for decoder.More() {
		recipeData := &RecipeData{}
		err := decoder.Decode(recipeData)
		if err != nil {
			toStdErr(err)
		}

		totalObjects += 1
		calc.calculateCountPerRecipe(recipeData.Recipe, countPerRecipe)
		calc.calculateCountPerPostcode(recipeData.Postcode, countPerPostcode)
		calc.calculateDeliveriesCountPerPostcode(recipeData, deliveriesCountPerPostcode)
		calc.filterRecipeName(recipeData, &filteredRecipeNames)
	}

	return calc.getExpectedOutput(countPerRecipe, countPerPostcode, deliveriesCountPerPostcode, filteredRecipeNames, int64(totalObjects))
}

func closeFile(f *os.File) {
	err := f.Close()
	if err != nil {
		toStdErr(err)
	}
}

func toStdErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}

func (calc *RecipeStatsCalculator) calculateCountPerRecipe(recipe string, countPerRecipe map[string]int) {

	_, ok := countPerRecipe[recipe]

	if !ok {
		countPerRecipe[recipe] = 1
	} else {
		countPerRecipe[recipe] += 1
	}
}

func (calc *RecipeStatsCalculator) calculateCountPerPostcode(postcode string, countPerPostcode map[string]int) {

	_, ok := countPerPostcode[postcode]

	if !ok {
		countPerPostcode[postcode] = 1
	} else {
		countPerPostcode[postcode] += 1
	}
}

func (calc *RecipeStatsCalculator) calculateDeliveriesCountPerPostcode(recipeData *RecipeData, deliveriesCountPerPostcode map[string]int) {

	postcode := recipeData.Postcode

	if postcode == calc.customPostcodeDeliveryTime.Postcode && calc.isWithinDeliveryTime(recipeData.Delivery) {
		deliveriesCountPerPostcode[postcode] += 1
	}
}

// `"delivery"` always has the following format: "{weekday} {h}AM - {h}PM", i.e. "Monday 9AM - 5PM"
func (calc *RecipeStatsCalculator) isWithinDeliveryTime(delivery string) bool {

	re := regexp.MustCompile(`(\d{0,2})AM\s-\s(\d{0,2})PM`)
	matches := re.FindStringSubmatch(delivery)

	from, err := strconv.Atoi(matches[1])
	if err != nil {
		toStdErr(err)
	}

	to, err := strconv.Atoi(matches[2])
	if err != nil {
		toStdErr(err)
	}

	return from >= calc.customPostcodeDeliveryTime.From && to <= calc.customPostcodeDeliveryTime.To
}

func (calc *RecipeStatsCalculator) filterRecipeName(recipeData *RecipeData, filteredRecipeNames *[]string) {

	recipe := recipeData.Recipe

	for _, customRecipeName := range calc.customRecipeNames {
		if strings.Contains(strings.ToLower(recipe), customRecipeName) && !alreadyFiltered(recipe, *filteredRecipeNames) {
			*filteredRecipeNames = append(*filteredRecipeNames, recipe)
			break
		}
	}
}

func alreadyFiltered(recipe string, filteredRecipeNames []string) bool {

	for _, filteredRecipeName := range filteredRecipeNames {
		if filteredRecipeName == recipe {
			return true
		}
	}

	return false
}

// counts the number of unique recipe names
func (expectedOutput *ExpectedOutput) setUniqueRecipeCount(countPerRecipe map[string]int) *ExpectedOutput {

	uniqueRecipeCount := 0

	for _, count := range countPerRecipe {
		if count == 1 {
			uniqueRecipeCount += 1
		}
	}

	expectedOutput.UniqueRecipeCount = uniqueRecipeCount

	return expectedOutput
}

// counts the number of occurrences for each unique recipe name (alphabetically ordered by recipe name)
func (expectedOutput *ExpectedOutput) setSortedRecipeCount(countPerRecipe map[string]int) *ExpectedOutput {

	recipes := make([]string, 0, len(countPerRecipe))

	for recipe := range countPerRecipe {
		recipes = append(recipes, recipe)
	}

	sort.Strings(recipes)

	for _, recipe := range recipes {
		expectedOutput.SortedRecipesCount = append(expectedOutput.SortedRecipesCount, CountPerRecipe{
			Recipe: recipe,
			Count:  countPerRecipe[recipe],
		})
	}

	return expectedOutput
}

// finds the postcode with most delivered recipes
func (expectedOutput *ExpectedOutput) setBusiestPostcode(countPerPostcode map[string]int) *ExpectedOutput {

	var countPerPostcodeList []CountPerPostcode

	for postcode, count := range countPerPostcode {
		countPerPostcodeList = append(countPerPostcodeList, CountPerPostcode{
			Postcode: postcode,
			Count:    count,
		})
	}

	sort.Slice(countPerPostcodeList, func(i, j int) bool {
		return countPerPostcodeList[i].Count > countPerPostcodeList[j].Count
	})

	expectedOutput.BusiestPostcode = BusiestPostcode{
		Postcode:      countPerPostcodeList[0].Postcode,
		DeliveryCount: countPerPostcodeList[0].Count,
	}

	return expectedOutput
}

// counts the number of deliveries to postcode `10120` that lie within the delivery time between `10AM` and `3PM`
func (expectedOutput *ExpectedOutput) setDeliveriesCountForPostCode(
	customPostcodeDeliveryTime CustomPostcodeDeliveryTime,
	deliveryCount int) *ExpectedOutput {

	expectedOutput.CountPerPostcodeAndTime = CountPerPostcodeAndTime{
		Postcode:      customPostcodeDeliveryTime.Postcode,
		FromAM:        strconv.Itoa(customPostcodeDeliveryTime.From) + "AM",
		ToPM:          strconv.Itoa(customPostcodeDeliveryTime.To) + "PM",
		DeliveryCount: deliveryCount,
	}

	return expectedOutput
}

func (expectedOutput *ExpectedOutput) setSortedRecipeNames(filteredRecipeNames []string) {

	sort.Strings(filteredRecipeNames)
	expectedOutput.SortedRecipeNames = filteredRecipeNames
}

func (calc *RecipeStatsCalculator) getExpectedOutput(
	countPerRecipe map[string]int,
	countPerPostcode map[string]int,
	deliveriesCountPerPostcode map[string]int,
	filteredRecipeNames []string,
	totalObjects int64,
	) ExpectedOutput {

	var expectedOutput ExpectedOutput

	expectedOutput.
		setUniqueRecipeCount(countPerRecipe).
		setSortedRecipeCount(countPerRecipe).
		setBusiestPostcode(countPerPostcode).
		setDeliveriesCountForPostCode(calc.customPostcodeDeliveryTime, deliveriesCountPerPostcode[calc.customPostcodeDeliveryTime.Postcode]).
		setSortedRecipeNames(filteredRecipeNames)

	expectedOutput.TotalObjects = totalObjects

	return expectedOutput
}
