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

func (recipeStatsCalculator *RecipeStatsCalculator) CalculateStats(filePath string) {

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

		recipeStatsCalculator.calculateCountPerRecipe(recipeData, recipeCountMap)
		recipeStatsCalculator.calculateCountPerPostcode(recipeData, postcodeCountMap)
	}

	expectedOutput := recipeStatsCalculator.getExpectedOutput(recipeCountMap, postcodeCountMap, deliveriesCountPerPostcode, filteredRecipeNames)
	println(prettyPrint(expectedOutput))
}

func prettyPrint(i interface{}) string {
	s, _ := json2.MarshalIndent(i, "", "\t")
	return string(s)
}

func (recipeStatsCalculator *RecipeStatsCalculator) decodeJson(json string) RecipeData {

	text := []byte(json)
	var recipeData RecipeData

	err := json2.Unmarshal(text, &recipeData)
	if err != nil {
		panic(err)
	}

	return recipeData
}

func (recipeStatsCalculator *RecipeStatsCalculator) calculateCountPerRecipe(recipeData *RecipeData, recipeCountMap map[string]int) {

	recipe := recipeData.Recipe

	count, ok := recipeCountMap[recipe]
	if !ok {
		recipeCountMap[recipe] = 1
	} else {
		recipeCountMap[recipe] = count + 1
	}
}

func (recipeStatsCalculator *RecipeStatsCalculator) calculateCountPerPostcode(data *RecipeData, postcodeCountMap map[string]int) {

	postcode := data.Postcode

	count, ok := postcodeCountMap[postcode]
	if !ok {
		postcodeCountMap[postcode] = 1
	} else {
		postcodeCountMap[postcode] = count + 1
	}
}

func (recipeStatsCalculator *RecipeStatsCalculator) calculateDeliveriesCountPerPostcode(data RecipeData, deliveriesCountPerPostcode map[string]int) {

	postcode := data.Postcode
	if postcode == recipeStatsCalculator.PostcodeDeliveryTimeFilter.Postcode && recipeStatsCalculator.isWithinDeliveryTime(data.Delivery, recipeStatsCalculator.PostcodeDeliveryTimeFilter) {
		count := deliveriesCountPerPostcode[postcode]
		deliveriesCountPerPostcode[postcode] = count + 1
	}
}

func (recipeStatsCalculator *RecipeStatsCalculator) filterRecipeName(data RecipeData, filteredRecipeNames *[]string) {

	recipe := data.Recipe
	if strings.Contains(recipe, "Potato") || strings.Contains(recipe, "Veggie") || strings.Contains(recipe, "Mushroom") {
		*filteredRecipeNames = append(*filteredRecipeNames, recipe)
	}
}

// `"delivery"` always has the following format: "{weekday} {h}AM - {h}PM", i.e. "Monday 9AM - 5PM"
func (recipeStatsCalculator *RecipeStatsCalculator) isWithinDeliveryTime(delivery string, postcodeDeliveryTimeFilter PostcodeDeliveryTimeFilter) bool {

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
func (recipeStatsCalculator *RecipeStatsCalculator) setUniqueRecipeCount(recipeCountMap map[string]int, expectedOutput *ExpectedOutput) {

	uniqueRecipeCount := 0
	for _, count := range recipeCountMap {
		if count == 1 {
			uniqueRecipeCount += 1
		}
	}

	expectedOutput.UniqueRecipeCount = uniqueRecipeCount
}

// count the number of occurrences for each unique recipe name (alphabetically ordered by recipe name)
func (recipeStatsCalculator *RecipeStatsCalculator) setSortedRecipeCount(recipeCountMap map[string]int, expectedOutput *ExpectedOutput) {

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
func (recipeStatsCalculator *RecipeStatsCalculator) setBusiestPostcode(postcodeCountMap map[string]int, expectedOutput *ExpectedOutput) {

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

func (recipeStatsCalculator *RecipeStatsCalculator) setDeliveriesCountForPostCode(
	postcode string,
	deliveriesCountPerPostcode map[string]int,
	expectedOutput *ExpectedOutput) {

	expectedOutput.CountPerPostcodeAndTime = CountPerPostcodeAndTime{
		Postcode:      recipeStatsCalculator.PostcodeDeliveryTimeFilter.Postcode,
		From:          strconv.Itoa(recipeStatsCalculator.PostcodeDeliveryTimeFilter.FromAM) + "AM",
		To:            strconv.Itoa(recipeStatsCalculator.PostcodeDeliveryTimeFilter.ToPM) + "PM",
		DeliveryCount: deliveriesCountPerPostcode[postcode],
	}
}

func (recipeStatsCalculator *RecipeStatsCalculator) setSortedRecipeNames(filteredRecipeNames []string, expectedOutput *ExpectedOutput) {

	sort.Strings(filteredRecipeNames)
	expectedOutput.SortedRecipeNames = filteredRecipeNames
}

func (recipeStatsCalculator *RecipeStatsCalculator) getExpectedOutput(
	recipeCountMap map[string]int,
	postcodeCountMap map[string]int,
	deliveriesCountPerPostcode map[string]int,
	filteredRecipeNames []string) ExpectedOutput {

	var expectedOutput ExpectedOutput

	recipeStatsCalculator.setUniqueRecipeCount(recipeCountMap, &expectedOutput)
	recipeStatsCalculator.setSortedRecipeCount(recipeCountMap, &expectedOutput)
	recipeStatsCalculator.setBusiestPostcode(postcodeCountMap, &expectedOutput)
	recipeStatsCalculator.setDeliveriesCountForPostCode(recipeStatsCalculator.PostcodeDeliveryTimeFilter.Postcode, deliveriesCountPerPostcode, &expectedOutput)
	recipeStatsCalculator.setSortedRecipeNames(filteredRecipeNames, &expectedOutput)

	return expectedOutput
}
