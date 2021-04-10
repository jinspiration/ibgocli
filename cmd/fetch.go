package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jinspiration/ibgocli/ibgo"
	"github.com/spf13/cobra"
)

var format = "20060102 15:04:05"

var fetchCmd = &cobra.Command{
	Use:   "fetch [bar/tick] [SYMBOL]",
	Short: "fetch historical data",
	Long:  `fetch historical data in either bar or tick format`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		// fmt.Println("fetch called", port)
		if len(args) != 2 {
			fmt.Println("check help for usage info")
			os.Exit(1)
		}

		if args[0] == "bar" {
			err := fetchBar(cmd, port, args[1])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else if args[0] == "tick" {
			err := fetchTick(cmd, port, args[1])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Please choose either bar or tick")
		}
	},
}

func init() {
	fetchCmd.Flags().StringP("start", "s", "", `startDataTime of fetch request, default one year before endDateTime for bar request or one day for tick request`)
	fetchCmd.Flags().StringP("end", "e", "", `endDataTime of fetch request, default current time`)
	fetchCmd.Flags().StringP("resolution", "r", "1d", `barSize to request. Available intervals: 1m, 1h, 1d`)
	fetchCmd.Flags().StringP("output", "o", "", `output csv file name`)
}

func fetchBar(cmd *cobra.Command, port int, symbol string) error {

	var endDateTime time.Time
	var startDateTime time.Time
	addr := "127.0.0.1:" + strconv.FormatInt(int64(port), 10)
	fmt.Println("Using", addr)

	c, errC := ibgo.NewClient(addr, 0, false)
	if errC != nil {
		return errors.New("Error connecting to TWS/Gateway. Check settings")
	}
	stock, _ := c.USStock(symbol)
	loc, _ := time.LoadLocation(stock.Detail.TimeZoneID)
	endF, _ := cmd.Flags().GetString("end")
	if len(endF) == 0 {
		endDateTime = time.Now()
	} else {
		endDateTime, _ = time.ParseInLocation(format, endF, loc)
	}

	startF, _ := cmd.Flags().GetString("start")
	if len(startF) == 0 {
		startDateTime = endDateTime.AddDate(-1, 0, 0)
	} else {
		startDateTime, _ = time.ParseInLocation(format, startF, loc)
	}
	if startDateTime.After(endDateTime) {
		return errors.New("startDateTime must be prior to endDateTime")
	}
	barSizeF, _ := cmd.Flags().GetString("resolution")
	fmt.Println("requesting bar data of", symbol, "from", startDateTime, "to", endDateTime, "bar size", barSizeF)
	var durationStr string
	var barSize string
	switch barSizeF {
	case "1m":
		barSize = "1 min"
		durationStr = "1 M"
	case "1h":
		barSize = "1 hour"
		durationStr = "1 Y"
	case "1d":
		barSize = "1 day"
		durationStr = "1 Y"
	}
	endDateTimes := make([]string, 0)
	for endDateTime.After(startDateTime) {
		endDateTimes = append(endDateTimes, endDateTime.Format(format))
		if durationStr == "1 Y" {
			endDateTime = endDateTime.AddDate(-1, 0, 0)
		} else if durationStr == "1 M" {
			endDateTime = endDateTime.AddDate(0, -1, 0)
		} else {
			endDateTime = endDateTime.AddDate(0, 0, -1)
		}
	}
	// fmt.Println(endDateTimes)
	// fmt.Println(barSize, durationStr)

	filename, _ := cmd.Flags().GetString("output")
	if len(filename) == 0 {
		filename = symbol + "_" + barSizeF + ".csv"
	}
	file, _ := os.Create(filename)
	defer file.Close()
	csvWriter := csv.NewWriter(file)
	for i := len(endDateTimes) - 1; i > -1; i-- {
		bars, _ := stock.HistoricalBar(endDateTimes[i], durationStr, barSize, "TRADES", true, false)
		fmt.Println(len(bars), "bars requested", "ending", endDateTimes[i])
		for j := 0; j < len(bars); j++ {
			bar := bars[j]
			str := bar.ToCSV()
			csvWriter.Write(str)
			if j%500 == 0 {
				csvWriter.Flush()
			}
		}
	}
	csvWriter.Flush()

	return nil
}

func fetchTick(cmd *cobra.Command, port int, symbol string) error {
	var endDateTime time.Time
	var startDateTime time.Time
	addr := "127.0.0.1:" + strconv.FormatInt(int64(port), 10)
	fmt.Println("Using", addr)
	c, errC := ibgo.NewClient(addr, 0, false)
	if errC != nil {
		return errors.New("Error connecting to TWS/Gateway. Check settings")
	}
	stock, _ := c.USStock(symbol)
	loc, _ := time.LoadLocation(stock.Detail.TimeZoneID)
	endF, _ := cmd.Flags().GetString("end")
	if len(endF) == 0 {
		endDateTime = time.Now()
	} else {
		endDateTime, _ = time.ParseInLocation(format, endF, loc)
	}

	startF, _ := cmd.Flags().GetString("start")
	if len(startF) == 0 {
		startDateTime = endDateTime.AddDate(0, 0, -1)
	} else {
		startDateTime, _ = time.ParseInLocation(format, startF, loc)
	}
	if startDateTime.After(endDateTime) {
		return errors.New("startDateTime must be prior to endDateTime")
	}
	fmt.Println("requesting tick data of", symbol, "from", startDateTime, "to", endDateTime)

	filename, _ := cmd.Flags().GetString("output")
	if len(filename) == 0 {
		filename = symbol + "_tick.csv"
	}
	file, _ := os.Create(filename)
	defer file.Close()
	csvWriter := csv.NewWriter(file)
	for {
		ticks, _ := stock.HistoricalTicks(startDateTime.Format(format), "", 1000, "TRADES", true)
		if len(ticks) == 0 {
			break
		}
		fmt.Println(len(ticks), "ticks requested")
		startDateTime = time.Unix(ticks[len(ticks)-1].Time, 0)
		for j := 0; j < len(ticks); j++ {
			ticks := ticks[j]
			str := ticks.ToCSV()
			csvWriter.Write(str)
			if j%500 == 0 {
				csvWriter.Flush()
			}
		}
		csvWriter.Flush()
		if startDateTime.After(endDateTime) {
			break
		}
	}
	return nil
}
