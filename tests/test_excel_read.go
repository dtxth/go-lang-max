package main

import (
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test_excel_read <file.xlsx>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	fmt.Printf("Opening file: %s\n", filePath)

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Printf("ERROR opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	fmt.Printf("Sheets found: %d\n", len(sheets))
	for i, sheet := range sheets {
		fmt.Printf("  Sheet %d: %s\n", i+1, sheet)
	}

	if len(sheets) == 0 {
		fmt.Println("No sheets found!")
		os.Exit(1)
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Printf("ERROR getting rows: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nTotal rows: %d\n", len(rows))
	
	if len(rows) > 0 {
		fmt.Printf("First row (header) columns: %d\n", len(rows[0]))
		if len(rows[0]) > 0 {
			fmt.Printf("First 5 headers: %v\n", rows[0][:min(5, len(rows[0]))])
		}
	}

	if len(rows) > 1 {
		fmt.Printf("Second row (data) columns: %d\n", len(rows[1]))
		if len(rows[1]) >= 18 {
			fmt.Println("✅ Row has 18+ columns")
		} else {
			fmt.Printf("⚠️  Row has only %d columns (expected 18)\n", len(rows[1]))
		}
	}

	fmt.Println("\n✅ File read successfully!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
