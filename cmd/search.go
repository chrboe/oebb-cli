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

func formatDelayTime(str string) (string, error) {
	if str == "" {
		return "", nil
	}

	delayTime, err := formatConnTime(str)
	if err != nil {
		return "", err
	}
	return rgbterm.InterpretStr("{#ff0000}" + delayTime + "{}"), nil
}

func formatDelayLine(depDelay, arrDelay string, dep, arr *string) string {
	delayLine := ""

	if depDelay == "" {
		delayLine += strings.Repeat(" ", len(*dep)+1)
	} else {
		delayLine += depDelay + " "
		*dep = util.Strikethrough(*dep)
	}

	if arrDelay != "" {
		delayLine += arrDelay
		*arr = util.Strikethrough(*arr)
	}

	return delayLine
}

func displaySection(section oebb.Section) error {
	dep, err := formatConnTime(section.From.Departure)
	if err != nil {
		return err
	}

	arr, err := formatConnTime(section.To.Arrival)
	if err != nil {
		return err
	}

	depDelay, err := formatDelayTime(section.From.DepartureDelay)
	if err != nil {
		return err
	}

	arrDelay, err := formatDelayTime(section.To.ArrivalDelay)
	if err != nil {
		return err
	}

	if depDelay != "" || arrDelay != "" {
		fmt.Println("\t" + formatDelayLine(depDelay, arrDelay, &dep, &arr))
	}

	cname := section.Category.DisplayName
	if cname == "" {
		cname = section.Category.ShortName
	}
	category := rgbterm.InterpretStr(fmt.Sprintf("\033[1m{#ffffff,%s}%-3s{}",
		section.Category.BarColor,
		strings.ToUpper(cname)),
	)

	times := rgbterm.InterpretStr(fmt.Sprintf("{#555555}%s{}{#555555}-{}{#555555}%s{}", dep, arr))
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

	depDelay, err := formatDelayTime(conn.From.DepartureDelay)
	if err != nil {
		return err
	}

	arrDelay, err := formatDelayTime(conn.To.ArrivalDelay)
	if err != nil {
		return err
	}

	if depDelay != "" || arrDelay != "" {
		fmt.Println(formatDelayLine(depDelay, arrDelay, &dep, &arr))
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

func nameOrMeta(station oebb.Station) string {
	if station.Name == "" {
		return station.Meta
	}

	return station.Name
}

var searchCmd = &cobra.Command{
	Use:   "search [from] [to]",
	Short: "Search connections",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		auth, err := oebb.Auth()
		if err != nil {
			panic(err)
		}

		from := args[0]
		to := args[1]

		fromStation, err := oebb.GetStations(from, auth)
		if err != nil {
			panic(err)
		}

		toStation, err := oebb.GetStations(to, auth)
		if err != nil {
			panic(err)
		}

		numResults, err := cmd.Flags().GetInt("results")
		if err != nil {
			panic(err)
		}

		connections, err := oebb.GetConnections(fromStation[0], toStation[0], auth, time.Now(), numResults)
		if err != nil {
			panic(err)
		}

		if len(connections) < 1 {
			errFrom := rgbterm.InterpretStr("{#cc6666}" + nameOrMeta(fromStation[0]) + "{}")
			errTo := rgbterm.InterpretStr("{#cc6666}" + nameOrMeta(toStation[0]) + "{}")
			fmt.Printf("No connections found from %s to %s\n",
				errFrom, errTo)
		}

		for _, conn := range connections {
			displayConnection(conn)
		}
	},
}
