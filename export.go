package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

type ExportData struct {
	Type   string
	Title  string
	Header []string
	Data   [][]interface{}
	Image  []byte
}

func (b *App) Export(exportType string, data *ExportData) string {
	var err error
	if exportType == "excel" {
		err = b.exportExcel(data)
	} else {
		err = b.exportCSV(data)
	}
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("ExportTable err=%v", err))
		return fmt.Sprintf("エクスポートできません err=%v", err)
	}
	return ""
}

func (b *App) exportExcel(data *ExportData) error {
	d := time.Now().Format("20060102150405")
	file, err := wails.SaveFileDialog(b.ctx, wails.SaveDialogOptions{
		DefaultFilename:      "TWLogAIAN_Export_" + data.Type + "_" + d + ".xlsx",
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "Excelファイル",
			Pattern:     "*.xlsx",
		}},
	})
	if err != nil {
		return err
	}
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", data.Title)
	row := 3
	col := 'A'
	for _, h := range data.Header {
		f.SetCellValue("Sheet1", fmt.Sprintf("%c%d", col, row), h)
		col++
	}
	imgCol := col + 2
	row++
	for _, l := range data.Data {
		col = 'A'
		for _, i := range l {
			f.SetCellValue("Sheet1", fmt.Sprintf("%c%d", col, row), i)
			col++
		}
		row++
	}
	if len(data.Image) > 0 {
		f.AddPictureFromBytes("Sheet1", fmt.Sprintf("%c2", imgCol), "", data.Type, "*.png", data.Image)
	}
	if err := f.SaveAs(file); err != nil {
		return err
	}
	return nil
}

func (b *App) exportCSV(data *ExportData) error {
	d := time.Now().Format("20060102150405")
	file, err := wails.SaveFileDialog(b.ctx, wails.SaveDialogOptions{
		DefaultFilename:      "TWLogAIAN_Export_" + data.Type + "_" + d + ".csv",
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "CSV ファイル",
			Pattern:     "*.csv",
		}},
	})
	if err != nil {
		return err
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{data.Title})
	w.Write(data.Header)
	for _, l := range data.Data {
		data := []string{}
		for _, i := range l {
			data = append(data, fmt.Sprintf("%v", i))
		}
		w.Write(data)
	}
	w.Flush()
	return w.Error()
}
