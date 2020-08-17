package calculator

import (
	"bufio"
	json2 "encoding/json"
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
	FromAM   int
	ToPM     int
}

type ExpectedOutput struct {
	UniqueRecipeCount       int                     `json:"unique_recipe_count"`
	SortedRecipesCount      []CountPerRecipe        `json:"count_per_recipe"`
	BusiestPostcode         BusiestPostcode         `json:"busiest_postcode"`
	CountPerPostcodeAndTime CountPerPostcodeAndTime `json:"count_per_postcode_and_time"`
	SortedRecipeNames       []string                `json:"match_by_name"`
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

// filter criteria is passed as params, so that we couldn't miss it
func (calc *RecipeStatsCalculator) CalculateStats(
	filePath string,
	customPostcodeDeliveryTime CustomPostcodeDeliveryTime,
	customRecipeNames []string) ExpectedOutput {

	calc.customPostcodeDeliveryTime = customPostcodeDeliveryTime
	calc.customRecipeNames = customRecipeNames

	file, err := os.Open(filePath)
	if err != nil {
		toStdErr(err)
	}
	defer closeFile(file)

	recipeCountMap := make(map[string]int)
	postcodeCountMap := make(map[string]int)
	deliveriesCountPerPostcode := make(map[string]int)
	var filteredRecipeNames []string

	r := bufio.NewReader(file)
	d := json2.NewDecoder(r)

	d.Token()
	for d.More() {
		recipeData := &RecipeData{}
		err := d.Decode(recipeData)
		if err != nil {
			toStdErr(err)
		}

		calc.calculateCountPerRecipe(recipeData, recipeCountMap)
		calc.calculateCountPerPostcode(recipeData, postcodeCountMap)
		calc.calculateDeliveriesCountPerPostcode(recipeData, deliveriesCountPerPostcode)
		calc.filterRecipeName(recipeData, &filteredRecipeNames)
	}

	return calc.getExpectedOutput(recipeCountMap, postcodeCountMap, deliveriesCountPerPostcode, filteredRecipeNames)
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

func (calc *RecipeStatsCalculator) calculateCountPerRecipe(recipeData *RecipeData, recipeCountMap map[string]int) {

	recipe := recipeData.Recipe

	count, ok := recipeCountMap[recipe]
	if !ok {
		recipeCountMap[recipe] = 1
	} else {
		recipeCountMap[recipe] = count + 1
	}
}

func (calc *RecipeStatsCalculator) calculateCountPerPostcode(recipeData *RecipeData, postcodeCountMap map[string]int) {

	postcode := recipeData.Postcode

	count, ok := postcodeCountMap[postcode]
	if !ok {
		postcodeCountMap[postcode] = 1
	} else {
		postcodeCountMap[postcode] = count + 1
	}
}

func (calc *RecipeStatsCalculator) calculateDeliveriesCountPerPostcode(recipeData *RecipeData, deliveriesCountPerPostcode map[string]int) {

	postcode := recipeData.Postcode

	if postcode == calc.customPostcodeDeliveryTime.Postcode && calc.isWithinDeliveryTime(recipeData.Delivery) {
		deliveriesCountPerPostcode[postcode] += 1
	}
}

func keyExists(key string, customRecipeNames []string) bool {

	for _, v := range customRecipeNames {
		if v == key {
			return true
		}
	}

	return false
}

func (calc *RecipeStatsCalculator) filterRecipeName(recipeData *RecipeData, filteredRecipeNames *[]string) {

	recipe := recipeData.Recipe

	for _, v := range calc.customRecipeNames {
		if strings.Contains(recipe, v) && !keyExists(recipe, *filteredRecipeNames) {
			*filteredRecipeNames = append(*filteredRecipeNames, recipe)
			break
		}
	}
}

// `"delivery"` always has the following format: "{weekday} {h}AM - {h}PM", i.e. "Monday 9AM - 5PM"
func (calc *RecipeStatsCalculator) isWithinDeliveryTime(delivery string) bool {

	re := regexp.MustCompile(`[a-zA-Z]+\s(\d{0,2})AM\s-\s(\d{0,2})PM`)
	matches := re.FindStringSubmatch(delivery)

	from, err := strconv.Atoi(matches[1])
	if err != nil {
		toStdErr(err)
	}

	to, err := strconv.Atoi(matches[2])
	if err != nil {
		toStdErr(err)
	}

	return from >= calc.customPostcodeDeliveryTime.FromAM && to <= calc.customPostcodeDeliveryTime.ToPM
}

// count the number of unique recipe names
func (calc *RecipeStatsCalculator) setUniqueRecipeCount(recipeCountMap map[string]int, expectedOutput *ExpectedOutput) {

	uniqueRecipeCount := 0

	for _, count := range recipeCountMap {
		if count == 1 {
			uniqueRecipeCount += 1
		}
	}

	expectedOutput.UniqueRecipeCount = uniqueRecipeCount
}

// count the number of occurrences for each unique recipe name (alphabetically ordered by recipe name)
func (calc *RecipeStatsCalculator) setSortedRecipeCount(recipeCountMap map[string]int, expectedOutput *ExpectedOutput) {

	keys := make([]string, 0, len(recipeCountMap))

	for k := range recipeCountMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		expectedOutput.SortedRecipesCount = append(expectedOutput.SortedRecipesCount, CountPerRecipe{
			Recipe: k, Count: recipeCountMap[k],
		})
	}
}

// find the postcode with most delivered recipes
func (calc *RecipeStatsCalculator) setBusiestPostcode(postcodeCountMap map[string]int, expectedOutput *ExpectedOutput) {

	type PostcodeCount struct {
		Key   string
		Value int
	}

	var postcodeCounts []PostcodeCount
	for k, v := range postcodeCountMap {
		postcodeCounts = append(postcodeCounts, PostcodeCount{k, v})
	}

	sort.Slice(postcodeCounts, func(i, j int) bool {
		return postcodeCounts[i].Value > postcodeCounts[j].Value
	})

	expectedOutput.BusiestPostcode = BusiestPostcode{
		Postcode: postcodeCounts[0].Key, DeliveryCount: postcodeCounts[0].Value,
	}
}

func (calc *RecipeStatsCalculator) setDeliveriesCountForPostCode(
	postcode string,
	deliveriesCountPerPostcode map[string]int,
	expectedOutput *ExpectedOutput) {

	expectedOutput.CountPerPostcodeAndTime = CountPerPostcodeAndTime{
		Postcode:      calc.customPostcodeDeliveryTime.Postcode,
		FromAM:        strconv.Itoa(calc.customPostcodeDeliveryTime.FromAM) + "AM",
		ToPM:          strconv.Itoa(calc.customPostcodeDeliveryTime.ToPM) + "PM",
		DeliveryCount: deliveriesCountPerPostcode[postcode],
	}
}

func (calc *RecipeStatsCalculator) setSortedRecipeNames(filteredRecipeNames []string, expectedOutput *ExpectedOutput) {

	sort.Strings(filteredRecipeNames)
	expectedOutput.SortedRecipeNames = filteredRecipeNames
}

func (calc *RecipeStatsCalculator) getExpectedOutput(
	recipeCountMap map[string]int,
	postcodeCountMap map[string]int,
	deliveriesCountPerPostcode map[string]int,
	filteredRecipeNames []string) ExpectedOutput {

	var expectedOutput ExpectedOutput

	calc.setUniqueRecipeCount(recipeCountMap, &expectedOutput)
	calc.setSortedRecipeCount(recipeCountMap, &expectedOutput)
	calc.setBusiestPostcode(postcodeCountMap, &expectedOutput)
	calc.setDeliveriesCountForPostCode(calc.customPostcodeDeliveryTime.Postcode, deliveriesCountPerPostcode, &expectedOutput)
	calc.setSortedRecipeNames(filteredRecipeNames, &expectedOutput)

	return expectedOutput
}
