package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/net/html"
)

// "github.com/jedib0t/go-pretty/v6/list"
// "github.com/jedib0t/go-pretty/v6/progress"
// "github.com/jedib0t/go-pretty/v6/text"

type Response struct {
	Start          string
	End            string
	Id             string
	Card_html      string
	Featured_event string
	Booked         string
}

type HeaderPair struct {
	first, second string
}

var responses []Response

var userAgent = HeaderPair{"User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:97.0) Gecko/20100101 Firefox/97.0"}
var accept = HeaderPair{"Accept", "application/json, text/javascript, */*; q=0.01"}
var acceptLanguage = HeaderPair{"Accept-Language", "en-US,en;q=0.5"}
var referer = HeaderPair{"Referer", "https://www.motibro.com/musers/explore"}
var xRequestedWith = HeaderPair{"X-Requested-With", "XMLHttpRequest"}
var connection = HeaderPair{"Connection", "keep-alive"}
var secFetchDest = HeaderPair{"Sec-Fetch-Dest", "empty"}
var secFetchMode = HeaderPair{"Sec-Fetch-Mode", "cors"}
var secFetchSite = HeaderPair{"Sec-Fetch-Site", "same-origin"}
var secGPC = HeaderPair{"Sec-GPC", "1"}
var DNT = HeaderPair{"DNT", "1"}
var origin = HeaderPair{"Origin", "https://www.motibro.com"}
var contentType = HeaderPair{"Content-Type", "application/x-www-form-urlencoded; charset=UTF-8"}

const (
	Booked    string = "BOOKED"
	NotBooked        = "NOT_BOOKED"
)

const (
	Book   string = "BOOK EVENT"
	Cancel        = "CANCEL EVENT"
)

const cBooked string = "Cancel booking"
const closed string = "cancellation closed"
const trainer string = "Bodnár László"
const exerciseType string = "Cross"

func main() {
	app := tview.NewApplication()
	runBookingTable(app)
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
				return Booked
			}
		}
	}
	response.Booked = NotBooked
	return NotBooked
}

func getClasses() []Response {
	p := log.Println

	url := ""
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

func bookAll() {
	p := log.Println

	for i := 0; i < len(responses); i++ {
			url := "https://www.motibro.com/musers/booking_do"
			body := "event_id=" + responses[i].Id + "&function=booking&page="
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


func cancelAll() {
	p := log.Println

	for i := 0; i < len(responses); i++ {
			url := "https://www.motibro.com/musers/booking_do"
			body := "event_id=" + responses[i].Id + "&function=cancel_booking&page="
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

func bookById(id string) {
	p := log.Println
	url := "https://www.motibro.com/musers/booking_do"
	body := "event_id=" + id + "&function=booking&page="	
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

	log.Println(res.Status, "res")
}

func cancel(id string) {
	p := log.Println
	url := "https://www.motibro.com/musers/booking_do"
	body := "event_id=" + id + "&function=cancel_booking&page="

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

func setColumns(table *tview.Table) int {
	columns := strings.Split("# Start End Trainer Type Id Status Func", " ")

	cols := len(columns)
	for c := 0; c < cols; c++ {
		color := tcell.ColorYellow
		table.SetCell(
			0, c,
			tview.NewTableCell(columns[c]).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
	}
	return len(columns)
}

func setRows(table *tview.Table) {
	responses = getClasses()

	for r := 1; r <= len(responses); r++ {
		color := tcell.ColorWhite
		btn := Cancel
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

		if responses[r-1].Booked != Booked {
			color = tcell.ColorRed
			btn = Book
		}

		table.SetCell(
			r, 6,
			tview.NewTableCell(responses[r-1].Booked).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))

		table.SetCell(
			r, 7,
			tview.NewTableCell(btn).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
	}

}

func runBookingTable(app *tview.Application) {
	
	table := tview.NewTable().
		SetBorders(true)

	cols := setColumns(table)

	setRows(table)

	table.Select(1, 1).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, true)
		}
		if key == tcell.KeyTab {
			row, _ := table.GetSelection()
			table.Select(row, cols-1)
			table.SetSelectable(true, true)
		}
	})

	table.SetSelectedFunc(func(row int, column int) {
		if table.GetCell(row, cols-1).Text == Book {
			app.Suspend(func () {
				showModal(Book,row-1)
			})
			defer setRows(table)
		}

		if table.GetCell(row, cols-1).Text == Cancel {
			app.Suspend(func () {
				showModal(Cancel,row-1)
			})

			defer setRows(table)
		}

		table.SetSelectable(true, true)
	})
	if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}

func showModal(actionStr string, index int, ) {
	modalApp := tview.NewApplication()
	modal := tview.NewModal().
		SetText("Event start: " +formatDateTime(responses[index].Start)+" ends: "+formatDateTime(responses[index].End)+"\n"+
			"Do you want to "+ actionStr +" this event?").
		AddButtons([]string{actionStr, "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if actionStr == Cancel {
				cancel(responses[index].Id)
				defer modalApp.Stop()
			}

			if actionStr == Book {
				bookById(responses[index].Id)
				defer modalApp.Stop()
			}
		})
	if err := modalApp.SetRoot(modal, true).SetFocus(modal).Run(); err != nil {
		panic(err)
	}
	modalApp.Stop()
}

