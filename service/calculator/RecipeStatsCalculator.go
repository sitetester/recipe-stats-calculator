package calculator

import (
	"bufio"
	json2 "encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type RecipeStatsCalculator struct {
	PostcodeDeliveryTimeFilter PostcodeDeliveryTimeFilter
}

type PostcodeDeliveryTimeFilter struct {
	Postcode string
	FromAM   int
	ToPM     int
}

type ExpectedOutput struct {
	UniqueRecipeCount       int
	SortedRecipeCountData   []CountPerRecipe
	BusiestPostcode         BusiestPostcode
	CountPerPostcodeAndTime CountPerPostcodeAndTime
	SortedRecipeNames       []string
}

type BusiestPostcode struct {
	Postcode      string
	DeliveryCount int
}

type CountPerRecipe struct {
	Recipe string
	Count  int
}

type RecipeData struct {
	Postcode string
	Recipe   string
	Delivery string
}

type CountPerPostcodeAndTime struct {
	Postcode      string
	From          string
	To            string
	DeliveryCount int
}

func (calc *RecipeStatsCalculator) CalculateStats(filePath string) {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	defer file.Close()

	recipeCountMap := make(map[string]int)
	postcodeCountMap := make(map[string]int)
	deliveriesCountPerPostcode := make(map[string]int)
	var filteredRecipeNames []string

	r := bufio.NewReader(file)
	d := json2.NewDecoder(r)

	d.Token()
	for d.More() {
		recipeData := &RecipeData{}
		d.Decode(recipeData)

		calc.calculateCountPerRecipe(recipeData, recipeCountMap)
		calc.calculateCountPerPostcode(recipeData, postcodeCountMap)
	}

	expectedOutput := calc.getExpectedOutput(recipeCountMap, postcodeCountMap, deliveriesCountPerPostcode, filteredRecipeNames)
	println(prettyPrint(expectedOutput))
}

func prettyPrint(i interface{}) string {
	s, _ := json2.MarshalIndent(i, "", "\t")
	return string(s)
}

func (calc *RecipeStatsCalculator) decodeJson(json string) RecipeData {

	text := []byte(json)
	var recipeData RecipeData

	err := json2.Unmarshal(text, &recipeData)
	if err != nil {
		panic(err)
	}

	return recipeData
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

func (calc *RecipeStatsCalculator) calculateCountPerPostcode(data *RecipeData, postcodeCountMap map[string]int) {

	postcode := data.Postcode

	count, ok := postcodeCountMap[postcode]
	if !ok {
		postcodeCountMap[postcode] = 1
	} else {
		postcodeCountMap[postcode] = count + 1
	}
}

func (calc *RecipeStatsCalculator) calculateDeliveriesCountPerPostcode(data RecipeData, deliveriesCountPerPostcode map[string]int) {

	postcode := data.Postcode
	if postcode == calc.PostcodeDeliveryTimeFilter.Postcode && calc.isWithinDeliveryTime(data.Delivery, calc.PostcodeDeliveryTimeFilter) {
		count := deliveriesCountPerPostcode[postcode]
		deliveriesCountPerPostcode[postcode] = count + 1
	}
}

func (calc *RecipeStatsCalculator) filterRecipeName(data RecipeData, filteredRecipeNames *[]string) {

	recipe := data.Recipe
	if strings.Contains(recipe, "Potato") || strings.Contains(recipe, "Veggie") || strings.Contains(recipe, "Mushroom") {
		*filteredRecipeNames = append(*filteredRecipeNames, recipe)
	}
}

// `"delivery"` always has the following format: "{weekday} {h}AM - {h}PM", i.e. "Monday 9AM - 5PM"
func (calc *RecipeStatsCalculator) isWithinDeliveryTime(delivery string, postcodeDeliveryTimeFilter PostcodeDeliveryTimeFilter) bool {

	r := regexp.MustCompile(`[a-zA-Z]+\s(\d{0,2})AM\s-\s(\d{0,2})PM`)
	matches := r.FindStringSubmatch(delivery)

	i, errI := strconv.Atoi(matches[1])
	if errI != nil {
		log.Println(errI)
	}

	j, errJ := strconv.Atoi(matches[2])
	if errI != nil {
		log.Println(errJ)
	}

	return i >= postcodeDeliveryTimeFilter.FromAM && j <= postcodeDeliveryTimeFilter.ToPM
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
		expectedOutput.SortedRecipeCountData = append(expectedOutput.SortedRecipeCountData, CountPerRecipe{
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
		Postcode:      calc.PostcodeDeliveryTimeFilter.Postcode,
		From:          strconv.Itoa(calc.PostcodeDeliveryTimeFilter.FromAM) + "AM",
		To:            strconv.Itoa(calc.PostcodeDeliveryTimeFilter.ToPM) + "PM",
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
	calc.setDeliveriesCountForPostCode(calc.PostcodeDeliveryTimeFilter.Postcode, deliveriesCountPerPostcode, &expectedOutput)
	calc.setSortedRecipeNames(filteredRecipeNames, &expectedOutput)

	return expectedOutput
}
