package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/aybabtme/rgbterm"
	"github.com/chrboe/oebb"
	"github.com/chrboe/oebb-cli/util"
	"github.com/spf13/cobra"
)

func parseConnTime(str string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05.999", str)
}

func formatConnTime(str string) (string, error) {
	t, err := parseConnTime(str)
	if err != nil {
		return "", err
	}
	return t.Format("15:04"), nil
}

func formatDuration(dur int) string {
	minutes := dur / 1000 / 60
	durHours := minutes / 60
	durMinutes := minutes % 60
	durStr := fmt.Sprintf("{#ffff00}%02d:%02d{}", durHours, durMinutes)
	return rgbterm.InterpretStr(durStr)
}

func displaySection(section oebb.Section) error {
	dep, _ := formatConnTime(section.From.Departure)
	arr, _ := formatConnTime(section.To.Arrival)
	cname := section.Category.DisplayName
	if cname == "" {
		cname = section.Category.ShortName
	}
	category := rgbterm.InterpretStr(fmt.Sprintf("\033[1m{#ffffff,%s}%-3s{}",
		section.Category.BarColor,
		strings.ToUpper(cname)),
	)

	times := rgbterm.InterpretStr(fmt.Sprintf("{#555555}%s-%s{}", dep, arr))
	fmt.Printf("\t%s %s %s -> %s\n", times, category, section.From.Name, section.To.Name)
	return nil
}

func displayConnection(conn oebb.Connection) error {
	dep, err := formatConnTime(conn.From.Departure)
	if err != nil {
		return err
	}

	arr, err := formatConnTime(conn.To.Arrival)
	if err != nil {
		return err
	}

	durStr := formatDuration(conn.Duration)
	fromStr := util.Bold(rgbterm.InterpretStr("{#cc6666}" + conn.From.Name + "{}"))
	toStr := util.Bold(rgbterm.InterpretStr("{#cc6666}" + conn.To.Name + "{}"))

	fmt.Printf("%s-%s (%s) %s -> %s\n", dep, arr, durStr, fromStr, toStr)
	for _, section := range conn.Sections {
		displaySection(section)
	}

	fmt.Println()
	return nil
}

var searchCmd = &cobra.Command{
	Use:   "search [from] [to]",
	Short: "Search connections",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := oebb.Auth()
		if err != nil {
			panic(err)
		}

		from := args[0]
		to := args[1]

		fromStation, err := oebb.GetStations(from, resp)
		toStation, err := oebb.GetStations(to, resp)

		connections, err := oebb.GetConnections(fromStation[0], toStation[0], resp)
		if err != nil {
			panic(err)
		}

		for _, conn := range connections {
			displayConnection(conn)
		}
	},
}
