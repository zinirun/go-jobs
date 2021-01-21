package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	title    string
	location string
	salary   string
	summary  string
}

var baseURL string = "https://kr.indeed.com/jobs?q=javascript&limit=50"
var viewJobURL string = "https://kr.indeed.com/viewjob?jk="

func main() {
	fmt.Println("Start Exctracting..")
	c := make(chan []extractedJob)
	var jobs []extractedJob
	totalPages := getPages()
	for i := 0; i < totalPages; i++ {
		go getPage(i, c)
	}
	for i := 0; i < totalPages; i++ {
		jobs = append(jobs, <-c...)
	}
	writeJobs(jobs)
}

func getPage(page int, c chan<- []extractedJob) {
	c2 := make(chan extractedJob)
	var jobs []extractedJob
	pageURL := baseURL + "&start=" + strconv.Itoa(page*50) // Int to Text
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c2)
	})

	for i := 0; i < searchCards.Length(); i++ {
		jobs = append(jobs, <-c2)
	}

	c <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := cleanString(card.Find(".title>a").Text()) //innerText
	location := cleanString(card.Find(".sjcl").Text())
	salary := cleanString(card.Find(".salaryText").Text())
	summary := cleanString(card.Find(".summary").Text())
	c <- extractedJob{id: id, title: title, location: location, salary: salary, summary: summary}
}

func getPages() int {
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func writeJobs(jobs []extractedJob) {
	fmt.Println("Make csv file..")
	c := make(chan []string)
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush() // 함수가 끝날때 데이터를 입력

	headers := []string{"Link", "Title", "Location", "Salary", "Summary"}
	hwErr := w.Write(headers)
	checkErr(hwErr)

	for _, job := range jobs {
		go makeJobSlice(job, c)
	}

	for range jobs {
		jwErr := w.Write(<-c)
		checkErr(jwErr)
	}

	fmt.Println("Done. Extracted ", len(jobs))
}

func makeJobSlice(job extractedJob, c chan<- []string) {
	c <- []string{viewJobURL + job.id, job.title, job.location, job.salary, job.summary}
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status ", res.StatusCode)
	}
}
