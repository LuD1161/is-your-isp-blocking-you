/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type IfConfigResponse struct {
	IP         string  `json:"ip"`
	Country    string  `json:"country"`
	RegionName string  `json:"region_name"`
	ZipCode    string  `json:"zip_code"`
	City       string  `json:"city"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	AsnOrg     string  `json:"asn_org"`
}

// getISPCmd represents the getISP command
var getISPCmd = &cobra.Command{
	Use:   "getISP",
	Short: "Get current ISP",
	Run: func(cmd *cobra.Command, args []string) {
		proxyTransport := SetProxyTransport()
		result, err := GetISP(proxyTransport)
		if err != nil {
			log.Fatal().Msgf("Error unmarshalling data from ifconfig : %s", err.Error())
		}
		// Print tabular output
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleLight)
		t.SetTitle("Current ISP Details")
		t.AppendHeader(table.Row{"Country", "IP", "ISP", "Region", "City"})
		t.AppendRows([]table.Row{
			{result.Country, result.IP, result.AsnOrg, result.RegionName, result.City},
		})
		t.AppendSeparator()
		t.Render()
	},
}

func init() {
	rootCmd.AddCommand(getISPCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getISPCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getISPCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
