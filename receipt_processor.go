// created by Quinton Coley https://github.com/qcoley https://www.linkedin.com/in/quinton-coley/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// structure for posting receipts on the server
type Receipt struct {
	ID     string `json:"ID"`
	Points int    `json:"Points"`
}

// start server and get json data points if it matches id
func get_handle(id string) {
	http.HandleFunc("/receipts/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// build receipt from struct and decode
		receipt_there := &Receipt{}
		err := json.NewDecoder(r.Body).Decode(receipt_there)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// look for receipt matching id, if found print points. else, not found
		if receipt_there.ID == id {
			fmt.Println("\nFound receipt, points:", receipt_there.Points)
			w.WriteHeader(http.StatusCreated)
			return

		} else {
			fmt.Println("\nReceipt not found, ID: ", id)
		}
	})

	if err := http.ListenAndServe(":8080", nil); err != http.ErrServerClosed {
		panic(err)
	}
}

// post json to the local server
func post_handle(ID string, Points int) error {
	receipt_here := &Receipt{
		ID:     ID,
		Points: Points}

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(receipt_here)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/receipts/process", "application/json", b)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	fmt.Println(resp.Status)
	return nil
}

func point_calculator(retailer string, purchaseDate string, purchaseTime string, total string, items []interface{}) int {

	var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	retail_new := nonAlphanumericRegex.ReplaceAllString(retailer, "")
	points := 0
	points += len(strings.ReplaceAll(retail_new, " ", ""))
	cents := total[strings.IndexByte(total, '.')+1:]

	if cents == "00" {
		// whole dollar, + 50 points, also multiple of .25, +25 points
		points += 50
		points += 25

		// not a whole dollar, check if multiple of .25
	} else {
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

	// read local json file and parse to interface for use in subsequent post and get
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
		"id":     uuid.New().String(),
		"points": point_calculator(res["retailer"].(string), res["purchaseDate"].(string), res["purchaseTime"].(string), res["total"].(string), (res["items"].([]interface{})))}

	// server start, look up receipt by id
	go get_handle((receipt_gen["id"].(string)))

	// wait a second to give server time
	time.Sleep(time.Second)

	// make a post request with local json file data
	if err := post_handle(receipt_gen["id"].(string), receipt_gen["points"].(int)); err != nil {
		fmt.Println(err)
	}
}
