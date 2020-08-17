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
	From     int
	To       int
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

	countPerRecipe := make(map[string]int)
	countPerPostcode := make(map[string]int)
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

		calc.calculateCountPerRecipe(recipeData.Recipe, countPerRecipe)
		calc.calculateCountPerPostcode(recipeData.Postcode, countPerPostcode)
		calc.calculateDeliveriesCountPerPostcode(recipeData, deliveriesCountPerPostcode)
		calc.filterRecipeName(recipeData, &filteredRecipeNames)
	}

	return calc.getExpectedOutput(countPerRecipe, countPerPostcode, deliveriesCountPerPostcode, filteredRecipeNames)
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
		if strings.Contains(recipe, customRecipeName) && !alreadyFiltered(recipe, *filteredRecipeNames) {
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

// count the number of unique recipe names
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

// count the number of occurrences for each unique recipe name (alphabetically ordered by recipe name)
func (expectedOutput *ExpectedOutput) setSortedRecipeCount(countPerRecipe map[string]int) *ExpectedOutput {

	recipes := make([]string, 0, len(countPerRecipe))

	for recipe := range countPerRecipe {
		recipes = append(recipes, recipe)
	}

	sort.Strings(recipes)

	for _, recipe := range recipes {
		expectedOutput.SortedRecipesCount = append(expectedOutput.SortedRecipesCount, CountPerRecipe{
			Recipe: recipe, Count: countPerRecipe[recipe],
		})
	}

	return expectedOutput
}

// find the postcode with most delivered recipes
func (expectedOutput *ExpectedOutput) setBusiestPostcode(countPerPostcode map[string]int) *ExpectedOutput {

	type PostcodeCount struct {
		Key   string
		Value int
	}

	var postcodeCounts []PostcodeCount
	for k, v := range countPerPostcode {
		postcodeCounts = append(postcodeCounts, PostcodeCount{k, v})
	}

	sort.Slice(postcodeCounts, func(i, j int) bool {
		return postcodeCounts[i].Value > postcodeCounts[j].Value
	})

	expectedOutput.BusiestPostcode = BusiestPostcode{
		Postcode: postcodeCounts[0].Key, DeliveryCount: postcodeCounts[0].Value,
	}

	return expectedOutput
}

func (expectedOutput *ExpectedOutput) setDeliveriesCountForPostCode(
	postcode string,
	deliveriesCountPerPostcode map[string]int,
	customPostcodeDeliveryTime CustomPostcodeDeliveryTime) *ExpectedOutput {

	expectedOutput.CountPerPostcodeAndTime = CountPerPostcodeAndTime{
		Postcode:      customPostcodeDeliveryTime.Postcode,
		FromAM:        strconv.Itoa(customPostcodeDeliveryTime.From) + "AM",
		ToPM:          strconv.Itoa(customPostcodeDeliveryTime.To) + "PM",
		DeliveryCount: deliveriesCountPerPostcode[postcode],
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
	filteredRecipeNames []string) ExpectedOutput {

	var expectedOutput ExpectedOutput

	expectedOutput.
		setUniqueRecipeCount(countPerRecipe).
		setSortedRecipeCount(countPerRecipe).
		setBusiestPostcode(countPerPostcode).
		setDeliveriesCountForPostCode(calc.customPostcodeDeliveryTime.Postcode, deliveriesCountPerPostcode, calc.customPostcodeDeliveryTime).
		setSortedRecipeNames(filteredRecipeNames)

	return expectedOutput
}
