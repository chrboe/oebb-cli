package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chrboe/oebb"

	"github.com/aybabtme/rgbterm"
)

func main() {
	resp, err := oebb.Auth()
	if err != nil {
		panic(err)
	}

	if len(os.Args) != 3 {
		fmt.Println("usage")
		return
	}

	from := os.Args[1]
	to := os.Args[2]

	fromStation, err := oebb.GetStations(from, resp)
	toStation, err := oebb.GetStations(to, resp)

	connections, err := oebb.GetConnections(fromStation[0], toStation[0], resp)
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
