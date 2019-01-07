package internal

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

func CreateTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator(" ")
	table.SetColWidth(100)
	return table
}
