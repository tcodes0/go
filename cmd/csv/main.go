// Copyright 2024 Raphael Thomazella. All rights reserved.
//  Use of this source code is governed by the BSD-3-Clause
//  license that can be found in the LICENSE file and online
//  at https://opensource.org/license/BSD-3-clause.

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func main() {
	exportCSV, err := os.Open("/home/vacation/Desktop/tcodes0-go/.local/export.csv")
	if err != nil {
		panic(err)
	}

	reader := csv.NewReader(exportCSV)

	rows, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	err = exportCSV.Close()
	if err != nil {
		panic(err)
	}

	readGroups(rows)

	exportCSV, err = os.OpenFile("/home/vacation/Desktop/tcodes0-go/.local/export.csv", os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}

	defer exportCSV.Close()
	writer := csv.NewWriter(exportCSV)

	err = writer.WriteAll(rows)
	if err != nil {
		panic(err)
	}
}

func readGroups(rows [][]string) {
	groupsTXT, err := os.Open("/home/vacation/Desktop/tcodes0-go/.local/pw-groups.txt")
	if err != nil {
		panic(err)
	}

	defer groupsTXT.Close()

	scanner := bufio.NewScanner(groupsTXT)
	var grp string

	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			panic(err)
		}

		line := scanner.Text()
		if strings.HasPrefix(line, "    ") {
			if grp == "" {
				panic("empty grp")
			}

			updateRow(rows, grp, line[4:])
		} else {
			grp = line
		}
	}
}

func updateRow(rows [][]string, group, title string) {
outer:
	for iRow, row := range rows {
		for iCol, colValue := range row {
			if iRow != 0 {
				if colValue == title {
					rows[iRow][iCol-1] = group

					break outer
				}
			}
		}
	}
}

func readit() {
	exportCSV, err := os.Open("/home/vacation/Desktop/interface/.local/export.csv")
	if err != nil {
		panic(err)
	}

	reader := csv.NewReader(exportCSV)

	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	head := make([]string, 156)
	title := 4
	user := 5
	// pass := 6
	// grpName := 2

	for j, record := range records {
		// fmt.Println(record)

	recordLoop:
		for i, value := range record {
			if j == 0 {
				head[i] = value

				continue
			}

			if i == title {
				if value == "" {
					value = "<empty>"
				}

				fmt.Println(value)
			}

			if i > user {
				continue recordLoop
			}
		}
	}
}
