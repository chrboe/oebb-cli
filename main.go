package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aybabtme/rgbterm"
)

const (
	AUTH_URL     = "https://tickets.oebb.at/api/domain/v3/init"
	STATIONS_URL = "https://tickets.oebb.at/api/hafas/v1/stations"
	JOURNEYS_URL = "https://tickets.oebb.at/api/hafas/v4/timetable"
)

//
// CONNECTIONS
//

type ConnectionsResponse struct {
	Connections []Connection `json:"connections"`
}

type DepartureStation struct {
	Name                       string `json:"name"`
	Esn                        int    `json:"esn"`
	Departure                  string `json:"departure"`
	DeparturePlatform          string `json:"departurePlatform"`
	DeparturePlatformDeviation string `json:"departurePlatformDeviation"`
	ShowAsResolvedMetaStation  bool   `json:"showAsResolvedMetaStation"`
}

type ArrivalStation struct {
	Name                      string `json:"name"`
	Esn                       int    `json:"esn"`
	Arrival                   string `json:"arrival"`
	ArrivalPlatform           string `json:"arrivalPlatform"`
	ArrivalPlatformDeviation  string `json:"arrivalPlatform"`
	ShowAsResolvedMetaStation bool   `json:"showAsResolvedMetaStation"`
}

type LongName struct {
	De string `json:"de"`
	En string `json:"en"`
	It string `json:"it"`
}

type Place struct {
	De string `json:"de"`
	En string `json:"en"`
	It string `json:"it"`
}

type Category struct {
	Name                          string   `json:"name"`
	Number                        string   `json:"number"`
	ShortName                     string   `json:"shortName"`
	DisplayName                   string   `json:"displayName"`
	LongName                      LongName `json:"longName"`
	BackgroundColor               string   `json:"backgroundColor"`
	FontColor                     string   `json:"fontColor"`
	BarColor                      string   `json:"barColor"`
	Place                         Place    `json:"place"`
	JourneyPreviewIconID          string   `json:"journeyPreviewIconId"`
	JourneyPreviewIconColor       string   `json:"journeyPreviewIconColor"`
	AssistantIconID               string   `json:"assistantIconId"`
	Train                         bool     `json:"train"`
	ParallelLongName              string   `json:"parallelLongName"`
	ParallelDisplayName           string   `json:"parallelDisplayName"`
	BackgroundColorDisabledMobile string   `json:"backgroundColorDisabledMobile"`
	BackgroundColorDisabled       string   `json:"backgroundColorDisabled"`
	FontColorDisabled             string   `json:"fontColorDisabled"`
	BarColorDisabled              string   `json:"barColorDisabled"`
}

type Section struct {
	From        DepartureStation `json:"from,omitempty"`
	To          ArrivalStation   `json:"to,omitempty"`
	Duration    int              `json:"duration"`
	Category    Category         `json:"category,omitempty"`
	Type        string           `json:"type"`
	HasRealtime bool             `json:"hasRealtime"`
}

type Connection struct {
	ID       string           `json:"id"`
	From     DepartureStation `json:"from"`
	To       ArrivalStation   `json:"to"`
	Sections []Section        `json:"sections"`
	Switches int              `json:"switches"`
	Duration int              `json:"duration"`
}

//
// JOURNEYS
//

type JourneyRequest struct {
	Reverse           bool        `json:"reverse"`
	DatetimeDeparture string      `json:"datetimeDeparture"`
	Filter            Filter      `json:"filter"`
	Passengers        []Passenger `json:"passengers"`
	Count             int         `json:"count"`
	DebugFilter       DebugFilter `json:"debugFilter"`
	SortType          string      `json:"sortType"`
	From              Station     `json:"from"`
	To                Station     `json:"to"`
	Timeout           struct{}    `json:"timeout"`
}

type Filter struct {
	Regionaltrains     bool `json:"regionaltrains"`
	Direct             bool `json:"direct"`
	ChangeTime         bool `json:"changeTime"`
	Wheelchair         bool `json:"wheelchair"`
	Bikes              bool `json:"bikes"`
	Trains             bool `json:"trains"`
	Motorail           bool `json:"motorail"`
	DroppedConnections bool `json:"droppedConnections"`
}

type ChallengedFlags struct {
	HasHandicappedPass bool `json:"hasHandicappedPass"`
	HasAssistanceDog   bool `json:"hasAssistanceDog"`
	HasWheelchair      bool `json:"hasWheelchair"`
	HasAttendant       bool `json:"hasAttendant"`
}

type Passenger struct {
	Type                string          `json:"type"`
	ID                  int             `json:"id"`
	Me                  bool            `json:"me"`
	Remembered          bool            `json:"remembered"`
	ChallengedFlags     ChallengedFlags `json:"challengedFlags"`
	Relations           []interface{}   `json:"relations"`
	Cards               []interface{}   `json:"cards"`
	BirthdateChangeable bool            `json:"birthdateChangeable"`
	BirthdateDeletable  bool            `json:"birthdateDeletable"`
	NameChangeable      bool            `json:"nameChangeable"`
	PassengerDeletable  bool            `json:"passengerDeletable"`
	IsSelected          bool            `json:"isSelected"`
}

type DebugFilter struct {
	NoAggregationFilter bool `json:"noAggregationFilter"`
	NoEqclassFilter     bool `json:"noEqclassFilter"`
	NoNrtpathFilter     bool `json:"noNrtpathFilter"`
	NoPaymentFilter     bool `json:"noPaymentFilter"`
	UseTripartFilter    bool `json:"useTripartFilter"`
	NoVbxFilter         bool `json:"noVbxFilter"`
	NoCategoriesFilter  bool `json:"noCategoriesFilter"`
}

type Station struct {
	Latitude  int    `json:"latitude"`
	Longitude int    `json:"longitude"`
	Name      string `json:"name"`
	Meta      string `json:"meta"`
	Number    int    `json:"number"`
}

//
// AUTH
//

type AuthResponse struct {
	AccessToken string `json:"accessToken"`
	Token       struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	} `json:"token"`
	Channel            string    `json:"channel"`
	SupportID          string    `json:"supportId"`
	CashID             string    `json:"cashId"`
	OrgUnit            int       `json:"orgUnit"`
	LegacyUserMigrated bool      `json:"legacyUserMigrated"`
	UserID             string    `json:"userId"`
	PersonID           string    `json:"personId"`
	CustomerID         string    `json:"customerId"`
	Realm              string    `json:"realm"`
	SessionID          string    `json:"sessionId"`
	SessionTimeout     int       `json:"sessionTimeout"`
	SessionVersion     string    `json:"sessionVersion"`
	SessionCreatedAt   time.Time `json:"sessionCreatedAt"`
	XffxIP             string    `json:"xffxIP"`
	Cookie             string
}

func auth() (AuthResponse, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", AUTH_URL, nil)
	resp, err := client.Do(req)

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var authResp AuthResponse
	json.Unmarshal(buf.Bytes(), &authResp)

	return authResp, err
}

func getJourneys(from, to Station, a AuthResponse) (*ConnectionsResponse, error) {
	client := &http.Client{}

	jr := JourneyRequest{
		Reverse:           false,
		DatetimeDeparture: time.Now().Format("2006-01-02T15:04:05.999"),
		Filter: Filter{
			Regionaltrains:     false,
			Direct:             false,
			ChangeTime:         false,
			Wheelchair:         false,
			Bikes:              false,
			Trains:             false,
			Motorail:           false,
			DroppedConnections: false,
		},
		Passengers: []Passenger{
			Passenger{
				Type:       "ADULT",
				ID:         1554277150,
				Me:         false,
				Remembered: false,
				ChallengedFlags: ChallengedFlags{
					HasHandicappedPass: false,
					HasAssistanceDog:   false,
					HasWheelchair:      false,
					HasAttendant:       false,
				},
				Relations:           []interface{}{},
				Cards:               []interface{}{},
				BirthdateChangeable: true,
				BirthdateDeletable:  true,
				NameChangeable:      true,
				PassengerDeletable:  true,
				IsSelected:          false,
			},
		},
		Count: 5,
		DebugFilter: DebugFilter{
			NoAggregationFilter: false,
			NoEqclassFilter:     false,
			NoNrtpathFilter:     false,
			NoPaymentFilter:     false,
			UseTripartFilter:    false,
			NoVbxFilter:         false,
			NoCategoriesFilter:  false,
		},
		SortType: "DEPARTURE",
		From:     from,
		To:       to,
	}

	body, err := json.Marshal(jr)

	req, err := http.NewRequest("POST", JOURNEYS_URL, bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Channel", a.Channel)
	req.Header.Add("AccessToken", a.AccessToken)
	req.Header.Add("SessionId", a.SessionID)
	req.Header.Add("x-ts-supportid", "WEB_"+a.SupportID)

	req.AddCookie(&http.Cookie{Name: "ts-cookie", Value: a.Cookie})
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	connections := &ConnectionsResponse{}
	json.Unmarshal(buf.Bytes(), connections)

	return connections, err
}

func getStations(name string, a AuthResponse) ([]Station, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", STATIONS_URL+"?name="+url.QueryEscape(name), nil)
	req.Header.Add("Channel", a.Channel)
	req.Header.Add("AccessToken", a.AccessToken)
	req.Header.Add("SessionId", a.SessionID)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var stations []Station
	json.Unmarshal(buf.Bytes(), &stations)

	return stations, nil
}

func main() {
	resp, err := auth()
	if err != nil {
		panic(err)
	}

	if len(os.Args) != 3 {
		fmt.Println("usage")
		return
	}

	from := os.Args[1]
	to := os.Args[2]

	fromStation, err := getStations(from, resp)
	toStation, err := getStations(to, resp)

	connections, err := getJourneys(fromStation[0], toStation[0], resp)
	if err != nil {
		panic(err)
	}

	for _, conn := range connections.Connections {
		depTime, _ := time.Parse("2006-01-02T15:04:05.999", conn.From.Departure)
		arrTime, _ := time.Parse("2006-01-02T15:04:05.999", conn.To.Arrival)
		minutes := conn.Duration / 1000 / 60
		durHours := minutes / 60
		durMinutes := minutes % 60
		durStr := fmt.Sprintf("{#ffff00}%02d:%02d{}", durHours, durMinutes)
		durStr = rgbterm.InterpretStr(durStr)

		fromStr := "\033[1m" + rgbterm.InterpretStr("{#cc6666}"+conn.From.Name+"{}")
		toStr := "\033[1m" + rgbterm.InterpretStr("{#cc6666}"+conn.To.Name+"{}")
		fmt.Printf("%s-%s (%s) %s -> %s\n", depTime.Format("15:04"), arrTime.Format("15:04"), durStr, fromStr, toStr)
		for _, section := range conn.Sections {
			dep, _ := time.Parse("2006-01-02T15:04:05.999", section.From.Departure)
			arr, _ := time.Parse("2006-01-02T15:04:05.999", section.To.Arrival)
			cname := section.Category.DisplayName
			if cname == "" {
				cname = section.Category.ShortName
			}
			category := rgbterm.InterpretStr(fmt.Sprintf("\033[1m{#ffffff,%s}%-3s{}",
				section.Category.BarColor,
				strings.ToUpper(cname)),
			)

			times := rgbterm.InterpretStr(fmt.Sprintf("{#555555}%s-%s{}", dep.Format("15:04"), arr.Format("15:04")))
			fmt.Printf("\t%s %s %s -> %s\n", times, category, section.From.Name, section.To.Name)
		}

		fmt.Println()
	}
}
