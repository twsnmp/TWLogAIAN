package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetExtractorTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetExtractorTypes() []ExtractorType {
	ret := []ExtractorType{}
	ret = append(ret, extractorTypes...)
	ret = append(ret, importedExtractorTypes...)
	return ret
}

// GetFieldTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetFieldTypes() map[string]*FieldType {
	ret := make(map[string]*FieldType)
	for k, v := range fieldTypes {
		ret[k] = v
	}
	for k, v := range importedFieldTypes {
		ret[k] = v
	}
	return ret
}

// ExtractorType : ログからデータを取得するパターン定義
type ExtractorType struct {
	Key       string
	Name      string
	Grok      string
	TimeField string
	IPFields  string
	MACFields string
	View      string
}

var extractorTypes = []ExtractorType{
	{
		Key:       "syslog",
		Name:      "syslog(BSD)",
		TimeField: "timestamp",
		Grok:      `%{SYSLOGBASE} %{GREEDYDATA:message}`,
		View:      "syslog",
	},
	{
		Key:       "syslog",
		Name:      "syslog(BSD/文字列PRI付き)",
		TimeField: "timestamp",
		Grok:      `%{SYSLOGTIMESTAMP:timestamp} %{SYSLOGHOST:logsource}\s+%{NOTSPACE:facility_str}\.%{NOTSPACE:severity_str}\s+%{SYSLOGPROG}:\s+%{GREEDYDATA:message}`,
		View:      "syslog",
	},
	{
		Key:       "syslogIETF",
		Name:      "syslog(IETF)",
		TimeField: "timestamp",
		Grok:      `%{TIMESTAMP_ISO8601:timestamp} (?:%{SYSLOGFACILITY} )?%{SYSLOGHOST:logsource} %{SYSLOGPROG} %{GREEDYDATA:message}`,
		View:      "syslog",
	},
	{
		Key:       "apacheCommon",
		Name:      "Apache(Common)",
		TimeField: "timestamp",
		Grok:      `%{COMMONAPACHELOG}`,
		IPFields:  "clientip",
		View:      "access",
	},
	{
		Key:       "apacheConbined",
		Name:      "Apache(Conbined)",
		TimeField: "timestamp",
		Grok:      `%{COMBINEDAPACHELOG}`,
		IPFields:  "clientip",
		View:      "access",
	},
}

type FieldType struct {
	Name string
	Type string // number,string,latlong,time
	Unit string
}

var fieldTypes = map[string]*FieldType{
	"_all":                 {Name: "ログの行全体", Type: "_all"},
	"httpversion":          {Name: "HTTPバージョン", Type: "string"},
	"ident":                {Name: "識別子", Type: "string"},
	"request":              {Name: "パス", Type: "string"},
	"response":             {Name: "応答コード", Type: "number"},
	"agent":                {Name: "ユーザーエージェント", Type: "string"},
	"bytes":                {Name: "サイズ", Type: "number"},
	"clientip_geo_latlong": {Name: "クライアントの緯度経度", Type: "latlong"},
	"timestamp":            {Name: "タイムスタンプ", Type: "timestamp"},
	"rawrequest":           {Name: "リクエスト", Type: "string"},
	"referrer":             {Name: "リファラー", Type: "string"},
	"time":                 {Name: "日時", Type: "_time"},
	"verb":                 {Name: "リクエスト", Type: "string"},
	"_id":                  {Name: "内部ID", Type: "_id"},
	"auth":                 {Name: "ユーザー名", Type: "string"},
	"clientip":             {Name: "クライアントIP", Type: "string"},
	"clientip_geo":         {Name: "クライアント位置", Type: "geo"},
	"clientip_geo_city":    {Name: "クライアント都市", Type: "string"},
	"clientip_geo_country": {Name: "クライアントの国", Type: "string"},
	"clientip_host":        {Name: "クライアントのホスト名", Type: "string"},
	"priority":             {Name: "プライオリティー", Type: "number"},
	"logsource":            {Name: "ログ送信元", Type: "string"},
	"message":              {Name: "メッセージ", Type: "string"},
	"facility":             {Name: "ファシリティー", Type: "number"},
	"pid":                  {Name: "PID", Type: "number"},
	"program":              {Name: "プロセス名", Type: "string"},
	"delta":                {Name: "前ログとの時間差", Type: "number"},
	"facility_str":         {Name: "ファシリティー", Type: "string"},
	"severity_str":         {Name: "優先度", Type: "string"},
}

func setFieldTypes(l *LogEnt) {
	for f, i := range l.KeyValue {
		switch i.(type) {
		case string:
			setFieldType(f, "string")
		case float64:
			setFieldType(f, "number")
		case *GeoEnt:
			setFieldType(f, "geo")
			setFieldType(f+"_country", "string")
			setFieldType(f+"_city", "string")
			setFieldType(f+"_latlong", "latlong")
		}
	}
}

func setFieldType(f, t string) {
	if _, ok := fieldTypes[f]; !ok {
		if _, ok := importedFieldTypes[f]; !ok {
			fieldTypes[f] = &FieldType{
				Name: f + "(自動追加)",
				Type: t,
			}
		}
	}
}

var importedExtractorTypes = []ExtractorType{}
var importedFieldTypes = make(map[string]*FieldType)

// importExtractorTypes : 抽出パターン定義のインポート
func (b *App) importExtractorTypes() {
	importedExtractorTypes = []ExtractorType{}
	f, err := os.Open(filepath.Join(b.workdir, "extractor.tsv"))
	if err != nil {
		return
	}
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.Comment = '#'
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return
	}
	for i, v := range records {
		if len(v) > 1 {
			e := ExtractorType{
				Key:  fmt.Sprintf("EXT%d", i),
				Name: v[0],
				Grok: v[1],
			}
			if len(v) > 2 {
				e.TimeField = v[2]
				if len(v) > 3 {
					e.IPFields = v[3]
					if len(v) > 4 {
						e.MACFields = v[4]
						if len(v) > 5 {
							e.View = v[5]
						}
					}
				}
			}
			importedExtractorTypes = append(importedExtractorTypes, e)
		}
	}
}

// importFieldTypes : 抽出項目のインポート
func (b *App) importFieldTypes() {
	importedFieldTypes = make(map[string]*FieldType)
	f, err := os.Open(filepath.Join(b.workdir, "fields.tsv"))
	if err != nil {
		return
	}
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.Comment = '#'
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return
	}
	for _, v := range records {
		if len(v) > 2 {
			f := FieldType{
				Name: v[1],
				Type: v[2],
			}
			if len(v) > 3 {
				f.Unit = v[3]
			}
			importedFieldTypes[v[0]] = &f
		}
	}
}
