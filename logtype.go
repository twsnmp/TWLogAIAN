package main

import (
	"fmt"
	"os"
	"strings"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v2"
)

var extractorTypes = make(map[string]ExtractorType)
var fieldTypes = make(map[string]FieldType)

func makeDefalutLogTypes() {
	OutLog("makeDefalutLogTypes")
	extractorTypes = make(map[string]ExtractorType)
	fieldTypes = make(map[string]FieldType)
	for k, v := range defalutExtractorTypes {
		v.Key = k
		extractorTypes[k] = v
	}
	for k, v := range defalutFieldTypes {
		v.Key = k
		fieldTypes[k] = v
	}
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
	CanEdit   bool
}

var defalutExtractorTypes = map[string]ExtractorType{
	"syslog": {
		Name:      "syslog(BSD)",
		TimeField: "timestamp",
		Grok:      `%{SYSLOGBASE} %{GREEDYDATA:message}`,
		View:      "syslog",
	},
	"syslogBSD_NOPID": {
		Name:      "syslog(BSD)",
		TimeField: "timestamp",
		Grok:      `%{SYSLOGTIMESTAMP:timestamp}\s+%{SYSLOGHOST:logsource}\s+%{NOTSPACE:tag}:\s+%{GREEDYDATA:message}`,
		View:      "syslog",
	},
	"syslogBSD_PRI": {
		Name:      "syslog(BSD with PRI)",
		TimeField: "timestamp",
		Grok:      `%{SYSLOGTIMESTAMP:timestamp} %{SYSLOGHOST:logsource}\s+%{NOTSPACE:facility_str}\.%{NOTSPACE:severity_str}\s+%{SYSLOGPROG}:\s+%{GREEDYDATA:message}`,
		View:      "syslog",
	},
	"syslogIETF": {
		Name:      "syslog(IETF)",
		TimeField: "timestamp",
		Grok:      `%{TIMESTAMP_ISO8601:timestamp}\s+(?:%{SYSLOGFACILITY} )?%{SYSLOGHOST:logsource}\s+%{NOTSPACE:tag}:\s+%{GREEDYDATA:message}`,
		View:      "syslog",
	},
	"apacheCommon": {
		Name:      "Apache(Common)",
		TimeField: "timestamp",
		Grok:      `%{COMMONAPACHELOG}`,
		IPFields:  "clientip",
		View:      "access",
	},
	"apacheConbined": {
		Name:      "Apache(Conbined)",
		TimeField: "timestamp",
		Grok:      `%{COMBINEDAPACHELOG}`,
		IPFields:  "clientip",
		View:      "access",
	},
	"SSHLOGIN": {
		Name:     "SSH Login",
		Grok:     `%{NOTSPACE:stat} (password|publickey) for( invalid user | )%{USER:user} from %{IP:clientip}`,
		IPFields: "clientip",
	},
}

type FieldType struct {
	Key     string
	Name    string
	Type    string // number,string,latlong,time
	Unit    string
	Mul     float64
	CanEdit bool
}

var defalutFieldTypes = map[string]FieldType{
	"_all":                 {Name: "ALL", Type: "_all"},
	"httpversion":          {Name: "HTTP Version", Type: "string"},
	"ident":                {Name: "User ID", Type: "string"},
	"request":              {Name: "Path", Type: "string"},
	"response":             {Name: "Resp Code", Type: "number"},
	"agent":                {Name: "User Agent", Type: "string"},
	"bytes":                {Name: "Size", Type: "number"},
	"timestamp":            {Name: "Time stamp", Type: "timestamp"},
	"rawrequest":           {Name: "Raw request", Type: "string"},
	"referrer":             {Name: "Referrer", Type: "string"},
	"time":                 {Name: "Time", Type: "_time"},
	"verb":                 {Name: "Verb", Type: "string"},
	"_id":                  {Name: "ID", Type: "_id"},
	"auth":                 {Name: "Auth", Type: "string"},
	"clientip":             {Name: "Client IP", Type: "string"},
	"clientip_geo":         {Name: "Client Geo", Type: "geo"},
	"clientip_geo_city":    {Name: "Client City", Type: "string"},
	"clientip_geo_country": {Name: "Client Country", Type: "string"},
	"clientip_geo_latlong": {Name: "Client LatLong", Type: "latlong"},
	"clientip_host":        {Name: "Client Host", Type: "string"},
	"logsource":            {Name: "Log Src", Type: "string"},
	"priority":             {Name: "Priority", Type: "number"},
	"severity":             {Name: "Severity", Type: "number"},
	"facility":             {Name: "Facility", Type: "number"},
	"message":              {Name: "Message", Type: "string"},
	"pid":                  {Name: "PID", Type: "number"},
	"program":              {Name: "Process", Type: "string"},
	"delta":                {Name: "Delta", Type: "number"},
	"facility_str":         {Name: "Facility Name", Type: "string"},
	"severity_str":         {Name: "Severity Name", Type: "string"},
	"client":               {Name: "Client", Type: "string"},
	"client_geo":           {Name: "Client Geo", Type: "geo"},
	"client_geo_city":      {Name: "Client City", Type: "string"},
	"client_geo_country":   {Name: "Client Country", Type: "string"},
	"client_host":          {Name: "Client Host", Type: "string"},
	"client_geo_latlong":   {Name: "Client LatLong", Type: "latlong"},
	"server":               {Name: "Server IP", Type: "string"},
	"server_geo":           {Name: "Server Geo", Type: "geo"},
	"server_geo_city":      {Name: "Server City", Type: "string"},
	"server_geo_country":   {Name: "Server Country", Type: "string"},
	"server_host":          {Name: "Server Host", Type: "string"},
	"server_geo_latlong":   {Name: "Server LatLong", Type: "latlong"},
	"user":                 {Name: "User", Type: "string"},
	"uri":                  {Name: "URI", Type: "string"},
	"email":                {Name: "EMail", Type: "string"},
	"uuid":                 {Name: "UUID", Type: "string"},
	"tag":                  {Name: "TAG", Type: "string"},
	"load1m":               {Name: "Load 1min", Type: "number"},
	"load5m":               {Name: "Load 5min", Type: "number"},
	"load15m":              {Name: "Load 15min", Type: "number"},
	"_None":                {Name: "None", Type: "string"},
	"winEventID":           {Name: "Event ID", Type: "number"},
	"winEventRecordID":     {Name: "Record ID", Type: "number"},
	"winChannel":           {Name: "Channel", Type: "number"},
	"winProvider":          {Name: "Provider", Type: "number"},
	"winLevel":             {Name: "Level", Type: "number"},
	"winUserID":            {Name: "User ID", Type: "number"},
	"winComputer":          {Name: "Computer", Type: "string"},
	"score":                {Name: "Score", Type: "number"},
	"anomalyScore":         {Name: "Anomaly Score", Type: "number"},
	"host":                 {Name: "Host", Type: "string"},
	"hostname":             {Name: "Host Name", Type: "string"},
	"level":                {Name: "Level", Type: "string"},
}

func (b *App) setFieldTypes(l *LogEnt) {
	for f, i := range l.KeyValue {
		switch i.(type) {
		case string:
			b.setFieldType(f, "string")
		case float64:
			b.setFieldType(f, "number")
		case *GeoEnt:
			b.setFieldType(f, "geo")
			b.setFieldType(f+"_country", "string")
			b.setFieldType(f+"_city", "string")
			b.setFieldType(f+"_latlong", "latlong")
		}
	}
}

func (b *App) setFieldType(f, t string) {
	if _, ok := fieldTypes[f]; !ok {
		fieldTypes[f] = FieldType{
			Name: f + "(Auto)",
			Type: t,
		}
	}
}

// GetExtractorTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetExtractorTypes() map[string]ExtractorType {
	OutLog("GetExtractorTypes")
	return extractorTypes
}

// SaveExtractorType : 抽出パターンを保存する
func (b *App) SaveExtractorType(et ExtractorType) string {
	if oldet, ok := extractorTypes[et.Key]; ok && !oldet.CanEdit {
		return "Can not add"
	}
	extractorTypes[et.Key] = et
	return ""
}

// DeleteExtractorType : 抽出パターンを削除する
func (b *App) DeleteExtractorType(key, title, message string) string {
	if et, ok := extractorTypes[key]; ok && !et.CanEdit {
		return "Can not Detele"
	}
	result, err := wails.MessageDialog(b.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil || result == "No" {
		return ""
	}
	delete(extractorTypes, key)
	return ""
}

// GetFieldTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetFieldTypes() map[string]FieldType {
	OutLog("GetFieldTypes")
	return fieldTypes
}

// SaveFieldType : フィールドタイプを削除する
func (b *App) SaveFieldType(ft FieldType) string {
	if oldft, ok := fieldTypes[ft.Key]; ok && !oldft.CanEdit {
		return "Can not add field type"
	}
	fieldTypes[ft.Key] = ft
	return ""
}

// DeleteFieldType : フィールドタイプを削除する
func (b *App) DeleteFieldType(key, title, message string) string {
	if ft, ok := fieldTypes[key]; ok && !ft.CanEdit {
		return "Can not delete"
	}
	result, err := wails.MessageDialog(b.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil || result == "No" {
		return ""
	}
	delete(fieldTypes, key)
	return ""
}

type exportLogType struct {
	ExtractorTypes []ExtractorType
	FieldTypes     []FieldType
}

func (b *App) ExportLogTypes() string {
	file, err := wails.SaveFileDialog(b.ctx, wails.SaveDialogOptions{
		DefaultFilename:      "logtypes.yaml",
		CanCreateDirectories: false,
		Filters: []wails.FileFilter{{
			DisplayName: "Log type yaml",
			Pattern:     "*.yaml",
		}},
	})
	if file == "" {
		return ""
	}
	if err != nil {
		OutLog("ExportLogTypes err=%v", err)
		return fmt.Sprintf("Can not export err=%v", err)
	}
	export := exportLogType{}
	for _, et := range extractorTypes {
		if et.CanEdit {
			export.ExtractorTypes = append(export.ExtractorTypes, et)
		}
	}
	for k, e := range fieldTypes {
		if e.CanEdit {
			e.Key = k
			export.FieldTypes = append(export.FieldTypes, e)
		}
	}
	d, err := yaml.Marshal(&export)
	if err != nil {
		return fmt.Sprintf("Can not export err=%v", err)
	}
	err = os.WriteFile(file, d, 0660)
	if err != nil {
		return fmt.Sprintf("Can not export err=%v", err)
	}
	return ""
}

func (b *App) getIPFields() string {
	m := make(map[string]bool)
	u := []string{}
	for _, ip := range b.processConf.HostFields {
		if !m[ip] {
			m[ip] = true
			u = append(u, ip)
		}
	}
	for _, ip := range b.processConf.GeoFields {
		if !m[ip] {
			m[ip] = true
			u = append(u, ip)
		}
	}
	return strings.Join(u, ",")
}

// ImportLogTypes : ログタイプ定義のインポート
func (b *App) ImportLogTypes() string {
	file, err := wails.OpenFileDialog(b.ctx, wails.OpenDialogOptions{
		DefaultFilename: "logtypes.yaml",
		Filters: []wails.FileFilter{{
			DisplayName: "Log type yaml",
			Pattern:     "*.yaml",
		}},
	})
	if file == "" {
		return ""
	}
	if err != nil {
		OutLog("importLogTypes err=%v", err)
		return err.Error()
	}
	d, err := os.ReadFile(file)
	if err != nil {
		OutLog("importLogTypes file=%v err=%v", file, err)
		return err.Error()
	}
	export := new(exportLogType)
	err = yaml.Unmarshal(d, &export)
	if err != nil {
		OutLog("importLogTypes err=%v", err)
		return err.Error()
	}
	etKeyMap := make(map[string]*ExtractorType)
	for _, et := range extractorTypes {
		etKeyMap[et.Key] = &et
	}
	for _, iet := range export.ExtractorTypes {
		if et, ok := etKeyMap[iet.Key]; ok && !et.CanEdit {
			return "duplicate name=" + iet.Name
		}
	}
	for _, iet := range export.ExtractorTypes {
		if et, ok := etKeyMap[iet.Key]; ok {
			et.Name = iet.Name
			et.Grok = iet.Grok
			et.TimeField = iet.TimeField
			et.IPFields = iet.IPFields
			et.MACFields = iet.MACFields
		} else {
			iet.CanEdit = true
			extractorTypes[iet.Key] = iet
		}
	}
	for _, e := range export.FieldTypes {
		if ft, ok := fieldTypes[e.Key]; !ok || ft.CanEdit {
			fieldTypes[e.Key] = e
		}
	}
	return ""
}
