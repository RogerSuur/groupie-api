package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
)

func main() {
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./assets")))) //	http.HandleFunc("/concerts", concerts) // handler for result site
	http.HandleFunc("/", handler)                                                          // handler for main page on site
	http.HandleFunc("/query", query)                                                       // handler for query results
	http.HandleFunc("/search", search)                                                     // handler for search bar
	fmt.Println("Starting server at localhost:8000")
	http.ListenAndServe(":8000", nil) // start web server on port 8000
}

func handler(w http.ResponseWriter, r *http.Request) { // creates main site using templates
	templ, err := template.ParseFiles("assets/index.html") // function to show html template on page
	if err != nil {
		http.Error(w, "500 Internal Server ERROR", http.StatusInternalServerError)
		return
	}
	if r.URL.Path != "/" {
		http.Error(w, "404 address NOT FOUND", http.StatusNotFound)
		return
	}
	getData(w, r)

	err = templ.ExecuteTemplate(w, "index.html", finaldata)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}
}

func search(w http.ResponseWriter, r *http.Request) { // creates search bar site using templates
	templ, err := template.ParseFiles("assets/search.html") // function to show html template on page
	if err != nil {
		http.Error(w, "500 Internal Server ERROR", http.StatusInternalServerError)
		return
	}
	if r.URL.Path != "/search" {
		http.Error(w, "Error 404\nPage not found!", 404)
		return
	}
	getData(w, r)

	err = templ.ExecuteTemplate(w, "search.html", finaldata)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}
}

func query(w http.ResponseWriter, r *http.Request) { // createsquery results site using templates
	templ, err := template.ParseFiles("assets/query.html") // function to show html template on page
	if err != nil {
		http.Error(w, "500 Internal Server ERROR", http.StatusInternalServerError)
		return
	}
	if r.URL.Path != "/query" {
		http.Error(w, "Error 404\nPage not found!", 404)
		return
	}
	getData(w, r)

	rquery := r.FormValue("mybands")
	rquery = strings.Title(rquery)
	query := strings.Split(rquery, " - ")
	intquery, _ := strconv.Atoi((query[0]))
	var oneartistData []allBands
	if len(query) > 1 { // checks is there search combination "word - word"
		switch query[1] {
		case "Band":
			for i := range artistData {
				if artistData[i].Name == query[0] {
					oneartistData = append(oneartistData, artistData[i])
				}
			}
		case "Creation date":
			for i := range artistData {
				if artistData[i].CreationDate == intquery {
					oneartistData = append(oneartistData, artistData[i])
				}
			}
		case "First album":
			for i := range artistData {
				if artistData[i].FirstAlbum == query[0] {
					oneartistData = append(oneartistData, artistData[i])
				}
			}
		case "Members":
			for i := range artistData {
				for j := range artistData[i].Members {
					if artistData[i].Members[j] == query[0] {
						oneartistData = append(oneartistData, artistData[i])
					}
				}
			}
		case "Locations":
			for i := range artistData {
				for j := range artistData[i].DatesLocations {
					if j == strings.ToLower(query[0]) {
						oneartistData = append(oneartistData, artistData[i])
					}
				}
			}
		}
	} else { // all other cases when there is just regular word on search bar
		for k := range artistData {

			if artistData[k].Name == query[0] {
				oneartistData = append(oneartistData, artistData[k])
			}
			if artistData[k].FirstAlbum == query[0] {
				oneartistData = append(oneartistData, artistData[k])
			}

			if artistData[k].CreationDate == intquery {
				oneartistData = append(oneartistData, artistData[k])
			}
			for l := range artistData[k].Members {
				if artistData[k].Members[l] == query[0] && artistData[k].Members[l] != artistData[k].Name {
					oneartistData = append(oneartistData, artistData[k])
				}
			}
			for j := range artistData[k].DatesLocations {
				if j == strings.ToLower(query[0]) {
					oneartistData = append(oneartistData, artistData[k])
				}
			}

		}
	}
	err = templ.ExecuteTemplate(w, "query.html", oneartistData) // shows only data according to search results
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}
}

func getData(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get("https://groupietrackers.herokuapp.com/api/artists") // takes artists data from API
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	bandData, err := ioutil.ReadAll(res.Body)
	if err = json.Unmarshal(bandData, &artistData); err != nil {
		log.Printf("Body parse error, %v", err)
		w.WriteHeader(500) // Return 500 Bad Request.
		return
	}
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/relation") // takes relations data from API
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err = json.Unmarshal(bandData, &artistData); err != nil {
		log.Printf("Body parse error, %v", err)
		w.WriteHeader(500) // Return 500 Bad Request.
		return
	}
	var concertData relationIndex

	json.Unmarshal(responseData, &concertData)
	relationData = concertData.Index

	for i, element := range relationData {
		artistData[i].DatesLocations = element.DatesLocations // replaces empty DatesLocations map with relations API data
	}
	tempmap := make(map[string][]int)
	tempbool := make(map[string]bool)

	for k := range concertData.Index {
		for i := range concertData.Index[k].DatesLocations {
			if tempbool[i] == false {
				// e2 := strings.Split(i, e)
				// strings.SplitAfter()
				e := strings.Title(i)
				tempmap[e] = append(tempmap[e], k)
			
				tempbool[i] = true
			}
		}
	}
	finaldata.Map = tempmap
	finaldata.ArtistData = artistData
}

type allBands struct {
	ID             int
	Image          string
	Name           string
	Members        []string
	CreationDate   int
	FirstAlbum     string
	Locations      string
	ConcertDates   string
	Relations      string
	DatesLocations map[string][]string
}

type relationIndex struct {
	Index []struct {
		Id             int
		DatesLocations map[string][]string
	}
}

var (
	artistData    []allBands
	oneartistData []allBands
	relationData  []struct {
		Id             int
		DatesLocations map[string][]string
	}
)

var UptDatesLocations map[string][]int

type manybands struct {
	Map        map[string][]int
	ArtistData []allBands
}

var finaldata manybands
