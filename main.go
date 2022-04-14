package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
    "time"
    "strings"
	"golang.org/x/net/html"
	"bytes"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"

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
	Booked string
}

type HeaderPair struct {
    first, second string
}

var userAgent = HeaderPair {"User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:97.0) Gecko/20100101 Firefox/97.0"}
var accept = HeaderPair {"Accept", "application/json, text/javascript, */*; q=0.01"}
var acceptLanguage = HeaderPair {"Accept-Language", "en-US,en;q=0.5"}
var referer = HeaderPair {"Referer", "https://www.motibro.com/musers/explore"}
var xCSRFToken = HeaderPair {"X-CSRF-Token", "avB6blLSzs34onRGmC6hSy5WBLoEOv/Ggd9TMmCTHTv+PJuCr8AsS2zA60Le9zXdep6AsyxX6NAoYSAooEcd7A=="}
var xRequestedWith = HeaderPair {"X-Requested-With", "XMLHttpRequest"}
var connection = HeaderPair {"Connection", "keep-alive"}
var cookie = HeaderPair {"Cookie", "_motibro_session5=WyPwampMXo96sHPkqLNY7yHLOEI%2FOXXRfwXAl0XjtCHP07Ix5lTOMF%2BvLEnLNgmAqtoGXHL4XI8D3FoYFnbjW5L7iki4ZfnPUtog7R6e6TJxvdGvA9zZikThrByOdcrn5JAAgvsFve3JzL4zan9fXSFS20F4JRcsv4u51bDpeKku%2F4sMivciLZ4wuLK6qlILYfIz0aimChWqKmNsNrAjlrTvDSdQ0j%2FIr9m4GI7q99WGPiBJ36IzQi%2FLot0IZ9UXlbZFh%2BXdWeLcTzQ9EHTML1j%2Fthoj%2B5CXR2ie0AxOiL8NFVQICYQGOhmTMiQ%3D--K08Qxcc%2Bl2uDPKmo--Q7mH6uRcc%2F08wqtGpD0TFQ%3D%3D"}
var secFetchDest = HeaderPair {"Sec-Fetch-Dest", "empty"}
var secFetchMode = HeaderPair {"Sec-Fetch-Mode", "cors"}
var secFetchSite = HeaderPair {"Sec-Fetch-Site", "same-origin"}
var secGPC = HeaderPair {"Sec-GPC", "1"}
var DNT = HeaderPair {"DNT", "1"}
var origin = HeaderPair {"Origin", "https://www.motibro.com"}
var contentType = HeaderPair {"Content-Type", "application/x-www-form-urlencoded; charset=UTF-8"}

const (
	Booked string = "BOOKED" 
	NotBooked	  = "NOT_BOOKED"
)

const booked string = "BOOKED"
const cBooked string = "Cancel booking"
const closed string = "cancellation closed"
const notBooked string = "NOT_BOOKED"
const trainer string = "Bodnár László"
const exerciseType string = "Cross"


func main() {
	responses := getClasses()


	 runBookingTable(responses)
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
			if len(TxtContent) > 0 && (TxtContent == cBooked || TxtContent == closed) {
				response.Booked = Booked
				return booked
			}
		}
	}
	response.Booked = NotBooked
 	return notBooked 
}

func getClasses() []Response {
	p := fmt.Println

	url := "https://www.motibro.com/musers/explore_get_events?date="+getCurrentDate()+"&member_id=122212&length_days=35&ts=1649588036192&event_ids=1735380&trainer_ids=414&location_ids=&premise_ids="

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		p(err)
	}

	req.Header.Add(userAgent.first, userAgent.second)
	req.Header.Add(accept.first, accept.second)
	req.Header.Add(acceptLanguage.first, acceptLanguage.second)
	req.Header.Add(referer.first, referer.second)
	req.Header.Add(xCSRFToken.first, xCSRFToken.second)
	req.Header.Add(xRequestedWith.first, xRequestedWith.second)
	req.Header.Add(connection.first, connection.second)
	req.Header.Add(cookie.first, cookie.second)
	req.Header.Add(secFetchDest.first, secFetchDest.second)
	req.Header.Add(secFetchMode.first, secFetchMode.second)
	req.Header.Add(secFetchSite.first, secFetchSite.second)
	req.Header.Add(secGPC.first, secGPC.second)
	req.Header.Add(DNT.first, DNT.second)

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

	for i, r := range responses {
		bookHTMLParser(r.Card_html, &responses[i])
	}

	return responses
}

func book(responses []Response) {
	p := fmt.Println

	for i := 0; i < len(responses); i++ {
		if responses[i].Booked == notBooked {
			url := "https://www.motibro.com/musers/booking_do"
			body := "event_id="+responses[i].Id+"&function=booking&page="

			client := &http.Client{}
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))

			if err != nil {
				p(err)
			}

			req.Header.Add(userAgent.first, userAgent.second)
			req.Header.Add(accept.first, accept.second)
			req.Header.Add(acceptLanguage.first, acceptLanguage.second)
			req.Header.Add(referer.first, referer.second)
			req.Header.Add(xCSRFToken.first, xCSRFToken.second)
			req.Header.Add(xRequestedWith.first, xRequestedWith.second)
			req.Header.Add(contentType.first, contentType.second)
			req.Header.Add(origin.first, origin.second)
			req.Header.Add(connection.first, connection.second)
			req.Header.Add(cookie.first, cookie.second)
			req.Header.Add(secFetchDest.first, secFetchDest.second)
			req.Header.Add(secFetchMode.first, secFetchMode.second)
			req.Header.Add(secFetchSite.first, secFetchSite.second)
			req.Header.Add(secGPC.first, secGPC.second)
			req.Header.Add(DNT.first, DNT.second)

			res, err := client.Do(req)
			if err != nil {
				p(err)
			}

			defer res.Body.Close()

			}
		}
}


func bookById(id string) {
	p := fmt.Println
			url := "https://www.motibro.com/musers/booking_do"
			body := "event_id="+id+"&function=booking&page="

			client := &http.Client{}
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))

			if err != nil {
				p(err)
			}

			req.Header.Add(userAgent.first, userAgent.second)
			req.Header.Add(accept.first, accept.second)
			req.Header.Add(acceptLanguage.first, acceptLanguage.second)
			req.Header.Add(referer.first, referer.second)
			req.Header.Add(xCSRFToken.first, xCSRFToken.second)
			req.Header.Add(xRequestedWith.first, xRequestedWith.second)
			req.Header.Add(contentType.first, contentType.second)
			req.Header.Add(origin.first, origin.second)
			req.Header.Add(connection.first, connection.second)
			req.Header.Add(cookie.first, cookie.second)
			req.Header.Add(secFetchDest.first, secFetchDest.second)
			req.Header.Add(secFetchMode.first, secFetchMode.second)
			req.Header.Add(secFetchSite.first, secFetchSite.second)
			req.Header.Add(secGPC.first, secGPC.second)
			req.Header.Add(DNT.first, DNT.second)

			res, err := client.Do(req)
			if err != nil {
				p(err)
			}

			defer res.Body.Close()
}



func cancel(id string) {
	p := fmt.Println
			url := "https://www.motibro.com/musers/booking_do"
			body := "event_id="+id+"&function=booking&page="

			client := &http.Client{}
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))

			if err != nil {
				p(err)
			}

			req.Header.Add(userAgent.first, userAgent.second)
			req.Header.Add(accept.first, accept.second)
			req.Header.Add(acceptLanguage.first, acceptLanguage.second)
			req.Header.Add(referer.first, referer.second)
			req.Header.Add(xCSRFToken.first, xCSRFToken.second)
			req.Header.Add(xRequestedWith.first, xRequestedWith.second)
			req.Header.Add(contentType.first, contentType.second)
			req.Header.Add(origin.first, origin.second)
			req.Header.Add(connection.first, connection.second)
			req.Header.Add(cookie.first, cookie.second)
			req.Header.Add(secFetchDest.first, secFetchDest.second)
			req.Header.Add(secFetchMode.first, secFetchMode.second)
			req.Header.Add(secFetchSite.first, secFetchSite.second)
			req.Header.Add(secGPC.first, secGPC.second)
			req.Header.Add(DNT.first, DNT.second)

			res, err := client.Do(req)
			if err != nil {
				p(err)
			}

			defer res.Body.Close()
}

func runBookingTable(responses []Response) {
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(true)

	columns :=	strings.Split("# Start End Trainer Type Id Status", " ")

	cols, rows := len(columns), len(responses)
	for c := 0; c < cols; c++ {
				color := tcell.ColorYellow
				table.SetCell(
					0, c,
				tview.NewTableCell(columns[c]).
					SetTextColor(color).
					SetAlign(tview.AlignCenter))
		}
	
	for r := 1; r <= rows; r++ {	
		color := tcell.ColorWhite
			table.SetCell(
				r, 0,
			tview.NewTableCell(strconv.Itoa(r)).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
			
			table.SetCell(
				r, 1,
			tview.NewTableCell(formatDateTime(responses[r-1].Start)).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
				
			table.SetCell(
				r, 2,
			tview.NewTableCell(formatDateTime(responses[r-1].End)).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
				
			table.SetCell(
				r, 3,
			tview.NewTableCell(trainer).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))		

			table.SetCell(
				r, 4,
			tview.NewTableCell(exerciseType).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))	

			table.SetCell(
				r, 5,
			tview.NewTableCell(responses[r-1].Id).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
			
			if (responses[r-1].Booked != booked) {
				color = tcell.ColorRed
			}	
			table.SetCell(
				r, 6,
			tview.NewTableCell(responses[r-1].Booked).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
	}	

	table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, false)
		}
	}).SetSelectedFunc(func(row int, column int) {	
		newCellBooked := 	tview.NewTableCell(booked).
		SetAlign(tview.AlignCenter)

		newCellNotBooked := 	tview.NewTableCell(booked).
		SetAlign(tview.AlignCenter)

		if(table.GetCell(row, cols-1).Text == notBooked) {
			bookById(responses[row].Id)
			table.SetCell(row, cols-1, newCellBooked)

			for i := 0; i < cols; i++ {
				table.GetCell(row, i).SetTextColor(tcell.ColorGreen)	
			}
		}

		if(table.GetCell(row, cols-1).Text == booked) {
			cancel(responses[row].Id)
			table.SetCell(row, cols-1, newCellNotBooked)

			for i := 0; i < cols; i++ {
				table.GetCell(row, i).SetTextColor(tcell.ColorRed)	
			}
		}
		
		table.SetSelectable(true, false)
	})
	if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}
