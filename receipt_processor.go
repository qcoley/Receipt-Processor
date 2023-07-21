package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func point_calculator(retailer string, purchaseDate string, purchaseTime string, total string, items []interface{}) int {

	var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	retail_new := nonAlphanumericRegex.ReplaceAllString(retailer, "")

	fmt.Println(len(strings.TrimSpace(retail_new)))
	fmt.Println(strings.ReplaceAll(retail_new, " ", ""))

	points := 0
	points += len(strings.ReplaceAll(retail_new, " ", ""))
	cents := total[strings.IndexByte(total, '.')+1:]

	if cents == "00" {
		// whole dollar, + 50 points, also multiple of .25, +25 points
		points += 50
		points += 25
	} else {
		// not a whole dollar, check if multiple of .25
		if cents == "25" || cents == "50" || cents == "75" {
			points += 25
		}
	}
	// add 5 points for each multiple of 2 items
	if len(items) > 2 {
		points += (len(items) / 2) * 5
	}

	// loop through items
	for i := 0; i < len(items); i++ {
		// if item description length is a multiple of 3
		if len(strings.TrimSpace(items[i].(map[string]interface{})["shortDescription"].(string)))%3 == 0 {
			// parse price string from json to float32
			price, err := strconv.ParseFloat(items[i].(map[string]interface{})["price"].(string), 32)
			if err != nil {
				log.Fatal(err)
				return 0
			}
			// multiply price, round to nearest digit and add to points
			points += int((price * 0.2) + 1)
		}
	}

	// parse days from purchase date into an integer representing the day
	month_days := purchaseDate[strings.IndexByte(purchaseDate, '-')+1:]
	days := month_days[strings.IndexByte(month_days, '-')+1:]
	days_int, err_days := strconv.Atoi(days)
	if err_days != nil {
		log.Fatal(err_days)
		return 0
	}
	// if date was odd, +6 to points
	if days_int%2 != 0 {
		points += 6
	}

	// parse hours from purchase time into integer representing the hour
	hour := purchaseTime[:strings.IndexByte(purchaseTime, ':')]
	hour_int, err_hour := strconv.Atoi(hour)
	if err_days != nil {
		log.Fatal(err_hour)
		return 0
	}
	// if hour is between 2pm and 4pm, +10 to points
	if hour_int >= 14 && hour_int <= 16 {
		points += 10
	}

	return points
}

func main() {

	fileContent, err_read := os.Open("simple-receipt.json")

	if err_read != nil {
		log.Fatal(err_read)
		return
	}
	defer fileContent.Close()

	byteResult, _ := io.ReadAll(fileContent)
	var res map[string]interface{}

	json.Unmarshal([]byte(byteResult), &res)

	// create a processed receipt with unique id and points calculated from function
	receipt_gen := map[string]interface{}{
		"id":     uuid.New(),
		"points": point_calculator(res["retailer"].(string), res["purchaseDate"].(string), res["purchaseTime"].(string), res["total"].(string), (res["items"].([]interface{}))),
	}

	receipt_json, err_marshal := json.Marshal(receipt_gen)
	if err_marshal != nil {
		log.Fatal(err_marshal)
		return
	}
	fmt.Printf("json data: %s\n", receipt_json)
}
