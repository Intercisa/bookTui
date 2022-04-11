package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
    "time"
    "strings"
    "github.com/jedib0t/go-pretty/v6/table"
    "os"
	"golang.org/x/net/html"

)
   // "github.com/jedib0t/go-pretty/v6/list"
   // "github.com/jedib0t/go-pretty/v6/progress"
   // "github.com/jedib0t/go-pretty/v6/text"

type Response struct {
    Start string
    End string
    Id string
    Card_html string
    Featured_event string
}

const booked string = "BOOKED"
const cBooked string = "Cancel booking"
const notBooked string = "NOT_BOOKED"
const trainer string = "Bodnár László"
const exerciseType string = "Cross"
var notBookedList []Response


func main() {
	pf := fmt.Printf
	var input string
	printInitMessage()
	for input != "q" {
		fmt.Scanln(&input)
		switch input {
		case "t":
			responses := getClasses()
    		printTable(responses)
			if len(notBookedList) != 0 {
				pf("There is %d not booked events!", len(notBookedList))
			}	
		}
	}
}

func getCurrentDate() string {
	t := time.Now().Local()
	return t.Format("2006-01-02")
}

func formatDateTime(dateTime string) string {
	layout := "2006-01-02T15:04:05.000Z"
	t, err := time.Parse(layout, dateTime)
	if err != nil {
		fmt.Println(err)
	}
	return t.Format("2006-01-02 15:04")
}

func printTable(responses []Response) {
    t := table.NewWriter()
    t.SetOutputMirror(os.Stdout)
    t.AppendHeader(table.Row{"#", "Start", "End", "Trainer", "Type", "Id", "Status"})

    for i := 0; i < len(responses); i++ {
        t.AppendRows([]table.Row{
            {i + 1,
			formatDateTime(responses[i].Start),
			formatDateTime(responses[i].End),
			trainer,
			exerciseType,
			responses[i].Id,
			bookHTMLParser(responses[i].Card_html, &responses[i])},
        })
    }
    t.AppendSeparator()
    t.AppendFooter(table.Row{"", "", "1.0", "Version"})
    t.Render()
}

func bookHTMLParser(text string, response *Response) (data string) {
    tkn := html.NewTokenizer(strings.NewReader(text))
	previousStartTokenTest := tkn.Token()
	loopDomTest:
	for {
		tt := tkn.Next()
		switch {
		case tt == html.ErrorToken:
			break loopDomTest // End of the document,  done
		case tt == html.StartTagToken:
			previousStartTokenTest = tkn.Token()
		case tt == html.TextToken:
			if previousStartTokenTest.Data == "script" {
				continue
			}
			TxtContent := strings.TrimSpace(html.UnescapeString(string(tkn.Text())))
			if len(TxtContent) > 0 && TxtContent == cBooked {
				return booked
			}
		}
	}
	notBookedList = append(notBookedList, *response)
 	return notBooked 
}

func getClasses() []Response {
	p := fmt.Println

	url := "https://www.motibro.com/musers/explore_get_events?date="+getCurrentDate()+"&member_id=122212&length_days=35&ts=1649588036192&event_ids=1735380&trainer_ids=414&location_ids=&premise_ids="
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		p(err)
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:97.0) Gecko/20100101 Firefox/97.0")
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Referer", "https://www.motibro.com/musers/explore")
	req.Header.Add("X-CSRF-Token", "avB6blLSzs34onRGmC6hSy5WBLoEOv/Ggd9TMmCTHTv+PJuCr8AsS2zA60Le9zXdep6AsyxX6NAoYSAooEcd7A==")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", "_motibro_session5=WyPwampMXo96sHPkqLNY7yHLOEI%2FOXXRfwXAl0XjtCHP07Ix5lTOMF%2BvLEnLNgmAqtoGXHL4XI8D3FoYFnbjW5L7iki4ZfnPUtog7R6e6TJxvdGvA9zZikThrByOdcrn5JAAgvsFve3JzL4zan9fXSFS20F4JRcsv4u51bDpeKku%2F4sMivciLZ4wuLK6qlILYfIz0aimChWqKmNsNrAjlrTvDSdQ0j%2FIr9m4GI7q99WGPiBJ36IzQi%2FLot0IZ9UXlbZFh%2BXdWeLcTzQ9EHTML1j%2Fthoj%2B5CXR2ie0AxOiL8NFVQICYQGOhmTMiQ%3D--K08Qxcc%2Bl2uDPKmo--Q7mH6uRcc%2F08wqtGpD0TFQ%3D%3D")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-GPC", "1")
	req.Header.Add("DNT", "1")

	res, err := client.Do(req)
	if err != nil {
		p(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		p(err)
	}

	var responses []Response  
	json.Unmarshal(body, &responses)
	return responses
}

func printInitMessage() {
	p := fmt.Println

	p("Please input:")
	p("q for: quit")
	p("t for: table")
}
