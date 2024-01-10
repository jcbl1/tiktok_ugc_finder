// Package fileopers provides subsidiary functions that are I/O related.
package fileopers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	ugcinfo "github.com/jcbl1/tiktok_ugc_finder/ugc_info"
	"github.com/xuri/excelize/v2"
)

func genFilename(ext string) string {
	uuidHead := strings.Split(uuid.NewString(), "-")[0]
	return workingDir + "/result-" + time.Now().Format("20060102-") + uuidHead + "." + ext
}

// SaveResultsAsJSON saves ugcs in a JSON file.
func SaveResultsAsJSON(ugcs []ugcinfo.UGCInfo) error {
	filename := genFilename("json")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(ugcs)
	if err != nil {
		return err
	}

	if _, err = f.Write(data); err != nil {
		return err
	}

	logResultsSaved(filename)

	return nil
}

// SaveResultsAsXLSX saves ugcs in an XLSX file.
func SaveResultsAsXLSX(ugcs []ugcinfo.UGCInfo) error {
	excel := excelize.NewFile() // creates a new working book.
	defer excel.Close()

	sheet := "Sheet1" // ensures the sheet name is "Sheet1"
	excel.SetSheetName("Sheet1", sheet)

	if err := setSheetHeaders(excel, sheet); err != nil { // sets styles of the first row.
		return err
	}

	for i, ugc := range ugcs { // writes each ugc to Sheet1 of excel.
		if err := excel.SetCellStr(sheet, fmt.Sprintf("A%d", i+2), ugc.Name); err != nil {
			return err
		}
		if err := excel.SetCellStr(sheet, fmt.Sprintf("B%d", i+2), ugc.Signature); err != nil {
			return err
		}
		if err := excel.SetCellStr(sheet, fmt.Sprintf("C%d", i+2), ugc.UniqueID); err != nil {
			return err
		}
		if err := excel.SetCellHyperLink(sheet, fmt.Sprintf("C%d", i+2), "https://www.tiktok.com/@"+ugc.UniqueID, "External"); err != nil {
			return err
		}
		if err := excel.SetCellInt(sheet, fmt.Sprintf("D%d", i+2), ugc.FollowerCount); err != nil {
			return err
		}
		if err := excel.SetCellStr(sheet, fmt.Sprintf("E%d", i+2), ugc.Gender); err != nil {
			return err
		}
		if err := excel.SetCellInt(sheet, fmt.Sprintf("F%d", i+2), ugc.AP); err != nil {
			return err
		}
		if err := excel.SetCellFloat(sheet, fmt.Sprintf("G%d", i+2), float64(ugc.AI), 4, 32); err != nil {
			return err
		}
		if err := excel.SetCellStr(sheet, fmt.Sprintf("H%d", i+2), strings.Join(ugc.Email, " ")); err != nil {
			return err
		}
		if err := excel.SetCellStr(sheet, fmt.Sprintf("I%d", i+2), ugc.LatestVideoTime.Format("2006/01/02")); err != nil {
			return err
		}
	}

	filename := genFilename("xlsx")
	if err := excel.SaveAs(filename); err != nil { // saves to file
		return err
	}
	logResultsSaved(filename)

	return nil
}

func setSheetHeaders(excel *excelize.File, sheet string) error {
	style, err := excel.NewStyle(
		&excelize.Style{
			Font: &excelize.Font{Bold: true},
		},
	)
	if err != nil {
		return err
	}
	if err := excel.SetRowStyle(sheet, 1, 1, style); err != nil {
		return err
	}
	if err := excel.SetCellStr(sheet, "A1", "Name"); err != nil {
		return err
	}
	if err := excel.SetCellStr(sheet, "B1", "Signature"); err != nil {
		return err
	}
	if err := excel.SetCellStr(sheet, "C1", "Unique ID"); err != nil {
		return err
	}
	if err := excel.SetCellStr(sheet, "D1", "Follower Count"); err != nil {
		return err
	}
	if err := excel.SetCellStr(sheet, "E1", "Gender"); err != nil {
		return err
	}
	if err := excel.SetCellStr(sheet, "F1", "Average Play"); err != nil {
		return err
	}
	if err := excel.SetCellStr(sheet, "G1", "Average Interaction Rate"); err != nil {
		return err
	}
	if err := excel.SetCellStr(sheet, "H1", "Email(s)"); err != nil {
		return err
	}
	if err := excel.SetCellStr(sheet, "I1", "Latest Video Time"); err != nil {
		return err
	}

	return nil
}

func logResultsSaved(filename string) {
	log.Println("Results saved at", filename)
}
