package main

import (
	"encoding/base64"
	"encoding/csv"
	"fmt"
	_ "image/png"
	"os"
	"strings"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

type ExportData struct {
	Type   string
	Title  string
	Header []string
	Data   [][]interface{}
	Image  string
}

func (b *App) Export(exportType string, data *ExportData) string {
	var err error
	switch exportType {
	case "logtypes":
		err = b.ExportLogTypes()
	case "excel":
		err = b.exportExcel(data)
	case "csv":
		err = b.exportCSV(data)
	default:
		return "サポートしていません"
	}
	if err != nil {
		OutLog("ExportTable err=%v", err)
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
	if data.Image != "" {
		b64data := data.Image[strings.IndexByte(data.Image, ',')+1:]
		img, err := base64.StdEncoding.DecodeString(b64data)
		if err != nil {
			return err
		}
		f.AddPictureFromBytes("Sheet1", fmt.Sprintf("%c2", imgCol), "", data.Type, ".png", img)
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
