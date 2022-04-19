package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	key, value string
}

type TableColumnPair struct {
	index int
	value string
}

var indexColumn = TableColumnPair{0, "#"}
var startColumn = TableColumnPair{2, "_____Start_____"}
var endColumn = TableColumnPair{3, "_____End_____"}
var trainerColumn = TableColumnPair{4, "_____Trainer_____"}
var typeColumn = TableColumnPair{5, "_____Type_____"}
var statusColumn = TableColumnPair{1, "_____Status_____"}

var responses []Response
var userAgent = HeaderPair{"User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:97.0) Gecko/20100101 Firefox/97.0"}
var accept = HeaderPair{"Accept", "application/json, text/javascript, */*; q=0.01"}
var acceptLanguage = HeaderPair{"Accept-Language", "en-US,en;q=0.5"}
var referer = HeaderPair{"Referer", "https://www.motibro.com/musers/explore"}
var sigInReferer = HeaderPair{"Referer", "https://www.motibro.com/signin"}
var xCSRFToken = HeaderPair{"X-CSRF-Token", "avB6blLSzs34onRGmC6hSy5WBLoEOv/Ggd9TMmCTHTv+PJuCr8AsS2zA60Le9zXdep6AsyxX6NAoYSAooEcd7A=="}
var xRequestedWith = HeaderPair{"X-Requested-With", "XMLHttpRequest"}
var connection = HeaderPair{"Connection", "keep-alive"}
var cookie = HeaderPair{"Cookie", "_motibro_session5=WyPwampMXo96sHPkqLNY7yHLOEI%2FOXXRfwXAl0XjtCHP07Ix5lTOMF%2BvLEnLNgmAqtoGXHL4XI8D3FoYFnbjW5L7iki4ZfnPUtog7R6e6TJxvdGvA9zZikThrByOdcrn5JAAgvsFve3JzL4zan9fXSFS20F4JRcsv4u51bDpeKku%2F4sMivciLZ4wuLK6qlILYfIz0aimChWqKmNsNrAjlrTvDSdQ0j%2FIr9m4GI7q99WGPiBJ36IzQi%2FLot0IZ9UXlbZFh%2BXdWeLcTzQ9EHTML1j%2Fthoj%2B5CXR2ie0AxOiL8NFVQICYQGOhmTMiQ%3D--K08Qxcc%2Bl2uDPKmo--Q7mH6uRcc%2F08wqtGpD0TFQ%3D%3D"}
var secFetchDest = HeaderPair{"Sec-Fetch-Dest", "empty"}
var signInsecFetchDest = HeaderPair{"Sec-Fetch-Dest", "document"}
var signInsecFetchMode = HeaderPair{"Sec-Fetch-Mode", "navigate"}
var secFetchMode = HeaderPair{"Sec-Fetch-Mode", "cors"}
var secFetchSite = HeaderPair{"Sec-Fetch-Site", "same-origin"}
var secGPC = HeaderPair{"Sec-GPC", "1"}
var DNT = HeaderPair{"DNT", "1"}
var origin = HeaderPair{"Origin", "https://www.motibro.com"}
var contentType = HeaderPair{"Content-Type", "application/x-www-form-urlencoded; charset=UTF-8"}
var secFetchUser = HeaderPair{"Sec-Fetch-User", "?1"}
var insecureRequest = HeaderPair{"Upgrade-Insecure-Requests", "1"}
var cacheControl = HeaderPair{"Cache-Control", "max-age=0"}
var ifNoneMatch = HeaderPair{"If-None-Match", "W/\"4a282e0f00ac036ccecdf6fc5ea727f7\""}

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
	setCredentials()
}

func setCredentials() (string, string) {
	var email, password string
	app := tview.NewApplication()
	form := tview.NewForm().
		AddInputField("Email", "", 30, nil, nil).
		AddPasswordField("Password", "", 10, '*', nil).
		AddButton("Quit", func() {
			app.Stop()
		})
	app.EnableMouse(true)
	form.AddButton("Login", func() {
		emailInputField := form.GetFormItemByLabel("Email").(*tview.InputField)
		passwordInputField := form.GetFormItemByLabel("Password").(*tview.InputField)
		email = emailInputField.GetText()
		password = passwordInputField.GetText()

		app.Suspend(func() {
			if signIn(email, password) {
				app := tview.NewApplication()
				runBookingTable(app)
				form.SetFieldBackgroundColor(tcell.ColorBlue)
			} else {
				form.SetFieldBackgroundColor(tcell.ColorRed)
			}
			emailInputField.SetText("")
			passwordInputField.SetText("")
		})
	})
	form.SetBorder(true).SetTitle("Please login - MotiBro").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).SetFocus(form).Run(); err != nil {
		panic(err)
	}
	return email, password
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

func signIn(email, password string) bool {
	password = "X3zrnFzJJ92SBBm"
	url := "https://www.motibro.com/users/login_main_login"
	payload := strings.NewReader("authenticity_token=2z6j%2FhQ9xupm20PnyCBErgHvjxLLjtVij0BTsdCUcKlKuOELcqQLZ3Ieyhh5Abl80%2BmGdgc9HU4T6ddJH2jv4A%3D%3D&email=" + email + "&password=" + password + "&commit=Bel%C3%A9p%C3%A9s")
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)

	if err != nil {
		log.Println(err)
	}
	req.Header.Add(userAgent.key, userAgent.value)
	req.Header.Add(accept.key, accept.value)
	req.Header.Add(acceptLanguage.key, acceptLanguage.value)
	req.Header.Add(sigInReferer.key, sigInReferer.value)
	req.Header.Add(referer.key, referer.value)
	req.Header.Add(xCSRFToken.key, xCSRFToken.value)
	req.Header.Add(xRequestedWith.key, xRequestedWith.value)
	req.Header.Add(connection.key, connection.value)
	req.Header.Add(cookie.key, cookie.value)
	req.Header.Add(signInsecFetchDest.key, signInsecFetchDest.value)
	req.Header.Add(signInsecFetchMode.key, signInsecFetchMode.value)
	req.Header.Add(secFetchSite.key, secFetchSite.value)
	req.Header.Add(secGPC.key, secGPC.value)
	req.Header.Add(DNT.key, DNT.value)

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	return strings.Contains(string(body), "/users/signing_in?email")
}

func getClasses() []Response {
	p := log.Println

	url := "https://www.motibro.com/musers/explore_get_events?date=" + getCurrentDate() + "&member_id=122212&length_days=35&ts=1649588036192&event_ids=1735380&trainer_ids=414&location_ids=&premise_ids="

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		p(err)
	}
	req.Header.Add(userAgent.key, userAgent.value)
	req.Header.Add(accept.key, accept.value)
	req.Header.Add(acceptLanguage.key, acceptLanguage.value)
	req.Header.Add(referer.key, referer.value)
	req.Header.Add(xCSRFToken.key, xCSRFToken.value)
	req.Header.Add(xRequestedWith.key, xRequestedWith.value)
	req.Header.Add(connection.key, connection.value)
	req.Header.Add(cookie.key, cookie.value)
	req.Header.Add(secFetchDest.key, secFetchDest.value)
	req.Header.Add(secFetchMode.key, secFetchMode.value)
	req.Header.Add(secFetchSite.key, secFetchSite.value)
	req.Header.Add(secGPC.key, secGPC.value)
	req.Header.Add(DNT.key, DNT.value)

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

		req.Header.Add(userAgent.key, userAgent.value)
		req.Header.Add(accept.key, accept.value)
		req.Header.Add(acceptLanguage.key, acceptLanguage.value)
		req.Header.Add(referer.key, referer.value)
		req.Header.Add(xCSRFToken.key, xCSRFToken.value)
		req.Header.Add(xRequestedWith.key, xRequestedWith.value)
		req.Header.Add(contentType.key, contentType.value)
		req.Header.Add(origin.key, origin.value)
		req.Header.Add(connection.key, connection.value)
		req.Header.Add(cookie.key, cookie.value)
		req.Header.Add(secFetchDest.key, secFetchDest.value)
		req.Header.Add(secFetchMode.key, secFetchMode.value)
		req.Header.Add(secFetchSite.key, secFetchSite.value)
		req.Header.Add(secGPC.key, secGPC.value)
		req.Header.Add(DNT.key, DNT.value)

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

		req.Header.Add(userAgent.key, userAgent.value)
		req.Header.Add(accept.key, accept.value)
		req.Header.Add(acceptLanguage.key, acceptLanguage.value)
		req.Header.Add(referer.key, referer.value)
		req.Header.Add(xCSRFToken.key, xCSRFToken.value)
		req.Header.Add(xRequestedWith.key, xRequestedWith.value)
		req.Header.Add(contentType.key, contentType.value)
		req.Header.Add(origin.key, origin.value)
		req.Header.Add(connection.key, connection.value)
		req.Header.Add(cookie.key, cookie.value)
		req.Header.Add(secFetchDest.key, secFetchDest.value)
		req.Header.Add(secFetchMode.key, secFetchMode.value)
		req.Header.Add(secFetchSite.key, secFetchSite.value)
		req.Header.Add(secGPC.key, secGPC.value)
		req.Header.Add(DNT.key, DNT.value)

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

	req.Header.Add(userAgent.key, userAgent.value)
	req.Header.Add(accept.key, accept.value)
	req.Header.Add(acceptLanguage.key, acceptLanguage.value)
	req.Header.Add(referer.key, referer.value)
	req.Header.Add(xCSRFToken.key, xCSRFToken.value)
	req.Header.Add(xRequestedWith.key, xRequestedWith.value)
	req.Header.Add(contentType.key, contentType.value)
	req.Header.Add(origin.key, origin.value)
	req.Header.Add(connection.key, connection.value)
	req.Header.Add(cookie.key, cookie.value)
	req.Header.Add(secFetchDest.key, secFetchDest.value)
	req.Header.Add(secFetchMode.key, secFetchMode.value)
	req.Header.Add(secFetchSite.key, secFetchSite.value)
	req.Header.Add(secGPC.key, secGPC.value)
	req.Header.Add(DNT.key, DNT.value)

	res, err := client.Do(req)
	if err != nil {
		p(err)
	}

	defer res.Body.Close()
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

	req.Header.Add(userAgent.key, userAgent.value)
	req.Header.Add(accept.key, accept.value)
	req.Header.Add(acceptLanguage.key, acceptLanguage.value)
	req.Header.Add(referer.key, referer.value)
	req.Header.Add(xCSRFToken.key, xCSRFToken.value)
	req.Header.Add(xRequestedWith.key, xRequestedWith.value)
	req.Header.Add(contentType.key, contentType.value)
	req.Header.Add(origin.key, origin.value)
	req.Header.Add(connection.key, connection.value)
	req.Header.Add(cookie.key, cookie.value)
	req.Header.Add(secFetchDest.key, secFetchDest.value)
	req.Header.Add(secFetchMode.key, secFetchMode.value)
	req.Header.Add(secFetchSite.key, secFetchSite.value)
	req.Header.Add(secGPC.key, secGPC.value)
	req.Header.Add(DNT.key, DNT.value)

	res, err := client.Do(req)
	if err != nil {
		p(err)
	}

	defer res.Body.Close()
}

func setColumns(table *tview.Table) {
	columns := strings.Split(indexColumn.value+
		" "+
		startColumn.value+
		" "+
		endColumn.value+
		" "+
		trainerColumn.value+
		" "+
		typeColumn.value+
		" "+
		statusColumn.value, " ")

	cols := len(columns)
	for c := 0; c < cols; c++ {
		color := tcell.ColorYellow
		table.SetCell(
			0, c,
			tview.NewTableCell(columns[c]).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
	}
}

func setRows(table *tview.Table) {
	responses = getClasses()
	for r := 1; r <= len(responses); r++ {
		color := tcell.ColorWhite
		backgroundColor := tcell.ColorBlack
		table.SetCell(
			r, indexColumn.index,
			tview.NewTableCell(strconv.Itoa(r)).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))

		table.SetCell(
			r, startColumn.index,
			tview.NewTableCell(formatDateTime(responses[r-1].Start)).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))

		table.SetCell(
			r, endColumn.index,
			tview.NewTableCell(formatDateTime(responses[r-1].End)).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))

		table.SetCell(
			r, trainerColumn.index,
			tview.NewTableCell(trainer).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))

		table.SetCell(
			r, typeColumn.index,
			tview.NewTableCell(exerciseType).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))

		color = tcell.ColorWhite
		if responses[r-1].Booked != Booked {
			backgroundColor = tcell.ColorDarkRed
		} else {
			backgroundColor = tcell.ColorDarkGreen
		}

		table.SetCell(
			r, statusColumn.index,
			tview.NewTableCell(responses[r-1].Booked).
				SetTextColor(color).
				SetBackgroundColor(backgroundColor).
				SetAlign(tview.AlignCenter))

	}
}

func runBookingTable(app *tview.Application) {
	table := tview.NewTable().
		SetBorders(true)
	setColumns(table)

	setRows(table)

	table.Select(1, 1).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, false)
		}
		if key == tcell.KeyTab {

		}
	})

	table.SetSelectedFunc(func(row int, column int) {
		if table.GetCell(row, statusColumn.index).Text == NotBooked {
			app.Suspend(func() {
				showModal(Book, row-1)
			})
			defer setRows(table)
		}

		if table.GetCell(row, statusColumn.index).Text == Booked {
			app.Suspend(func() {
				showModal(Cancel, row-1)
			})

			defer setRows(table)
		}

		table.SetSelectable(true, false)
	})

	frame := tview.NewFrame(table).
		SetBorders(0, 0, 0, 0, 0, 0)

	frame.SetBorder(true).
		SetTitle(fmt.Sprintf("MotiBro Table"))

	flex := tview.NewFlex().
		AddItem(tview.NewBox().SetBorder(false), 0, 1, false).
		AddItem(frame, 0, 9, true).
		AddItem(tview.NewBox().SetBorder(false), 0, 1, false)

	if err := app.SetRoot(flex, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}

func showModal(actionStr string, index int) {
	modalApp := tview.NewApplication()
	modal := tview.NewModal().
		SetText("Event start: " + formatDateTime(responses[index].Start) + " ends: " + formatDateTime(responses[index].End) + "\n" +
			"Do you want to " + actionStr + " this event?").
		AddButtons([]string{actionStr, "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {

			if buttonIndex == 0 && actionStr == Cancel {
				cancel(responses[index].Id)
				defer modalApp.Stop()
			} else if buttonIndex == 0 && actionStr == Book {
				bookById(responses[index].Id)
				defer modalApp.Stop()
			} else {
				modalApp.Stop()
			}
		})
	if err := modalApp.SetRoot(modal, true).SetFocus(modal).Run(); err != nil {
		panic(err)
	}
}
