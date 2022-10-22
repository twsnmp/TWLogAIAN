package main

import (
	"fmt"
	"io/ioutil"
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
		Name:      "syslog(BSD/PIDなし)",
		TimeField: "timestamp",
		Grok:      `%{SYSLOGTIMESTAMP:timestamp}\s+%{SYSLOGHOST:logsource}\s+%{NOTSPACE:tag}:\s+%{GREEDYDATA:message}`,
		View:      "syslog",
	},
	"syslogBSD_PRI": {
		Name:      "syslog(BSD/文字列PRI付き)",
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
		Name:     "SSHのログイン",
		Grok:     `%{NOTSPACE:stat} (password|publickey) for( invalid user | )%{USER:user} from %{IP:clientip}`,
		IPFields: "clientip",
	},
	// {
	// 	Key:       "DEVICE",
	// 	Name:      "デバイス情報(ip)",
	// 	Grok:      `mac=%{MAC:mac}.+ip=%{IP:ip}`,
	// 	IPFields:  "ip",
	// 	MACFields: "mac",
	// },
	// {
	// 	Key:       "DEVICER",
	// 	Name:      "デバイス情報(mac)",
	// 	Grok:      `ip=%{IP:ip}.+mac=%{MAC:mac}`,
	// 	IPFields:  "ip",
	// 	MACFields: "mac",
	// },
	// {
	// 	Key:  "WELFFLOW",
	// 	Name: "WELFフロー",
	// 	Grok: `src=%{IP:src}:%{BASE10NUM:sport}:.+dst=%{IP:dst}:%{BASE10NUM:dport}:.+proto=%{WORD:prot}.+sent=%{BASE10NUM:sent}.+rcvd=%{BASE10NUM:rcvd}.+spkt=%{BASE10NUM:spkt}.+rpkt=%{BASE10NUM:rpkt}`,
	// },
	// {
	// 	Key:  "UPTIME",
	// 	Name: "負荷(uptime)",
	// 	Grok: `load average: %{BASE10NUM:load1m}, %{BASE10NUM:load5m}, %{BASE10NUM:load15m}`,
	// },
	// {
	// 	Key:  "TWPCAP_STATS",
	// 	Name: "TWPCAPの統計情報",
	// 	Grok: `type=Stats,total=%{BASE10NUM:total},count=%{BASE10NUM:count},ps=%{BASE10NUM:ps}`,
	// },
	// {
	// 	Key:       "TWPCAP_IPTOMAC",
	// 	Name:      "TWPCAPのIPとMACアドレス",
	// 	Grok:      `type=IPToMAC,ip=%{IP:ip},mac=%{MAC:mac},count=%{BASE10NUM:count},change=%{BASE10NUM:chnage},dhcp=%{BASE10NUM:dhcp}`,
	// 	IPFields:  "ip",
	// 	MACFields: "mac",
	// },
	// {
	// 	Key:       "TWPCAP_DNS",
	// 	Name:      "TWPCAPのDNS問い合わせ",
	// 	Grok:      `type=DNS,sv=%{IP:server},DNSType=%{WORD:dnsType},Name=%{IPORHOST:name},count=%{BASE10NUM:count},change=%{BASE10NUM:chnage},lcl=%{IP:lastIP},lMAC=%{MAC:lastMAC}`,
	// 	IPFields:  "server,lastIP",
	// 	MACFields: "lastMAC",
	// },
	// {
	// 	Key:      "TWPCAP_DHCP",
	// 	Name:     "TWPCAPのDHCPサーバー情報",
	// 	Grok:     `type=DHCP,sv=%{IP:server},count=%{BASE10NUM:count},offer=%{BASE10NUM:offer},ack=%{BASE10NUM:ack},nak=%{BASE10NUM:nak}`,
	// 	IPFields: "server",
	// },
	// {
	// 	Key:      "TWPCAP_NTP",
	// 	Name:     "TWPCAPのNTPサーバー情報",
	// 	Grok:     `type=NTP,sv=%{IP:server},count=%{BASE10NUM:count},change=%{BASE10NUM:change},lcl=%{IP:client},version=%{BASE10NUM:version},stratum=%{BASE10NUM:stratum},refid=%{WORD:refid}`,
	// 	IPFields: "client,server",
	// },
	// {
	// 	Key:      "TWPCAP_RADIUS",
	// 	Name:     "TWPCAPのRADIUS通信情報",
	// 	Grok:     `type=RADIUS,cl=%{IP:client},sv=%{IP:server},count=%{BASE10NUM:count},req=%{BASE10NUM:request},accept=%{BASE10NUM:accept},reject=%{BASE10NUM:reject},challenge=%{BASE10NUM:challenge}`,
	// 	IPFields: "client,server",
	// },
	// {
	// 	Key:      "TWPCAP_TLSFlow",
	// 	Name:     "TWPCAPのTLS通信情報",
	// 	Grok:     `type=TLSFlow,cl=%{IP:client},sv=%{IP:server},serv=%{WORD:service},count=%{BASE10NUM:count},handshake=%{BASE10NUM:handshake},alert=%{BASE10NUM:alert},minver=%{DATA:minver},maxver=%{DATA:maxver},cipher=%{DATA:cipher},ft=`,
	// 	IPFields: "client,server",
	// },
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
	"_all":                 {Name: "ログの行全体", Type: "_all"},
	"httpversion":          {Name: "HTTPバージョン", Type: "string"},
	"ident":                {Name: "ユーザーID", Type: "string"},
	"request":              {Name: "パス", Type: "string"},
	"response":             {Name: "応答コード", Type: "number"},
	"agent":                {Name: "ユーザーエージェント", Type: "string"},
	"bytes":                {Name: "サイズ", Type: "number"},
	"timestamp":            {Name: "タイムスタンプ", Type: "timestamp"},
	"rawrequest":           {Name: "生リクエスト", Type: "string"},
	"referrer":             {Name: "リファラー", Type: "string"},
	"time":                 {Name: "日時", Type: "_time"},
	"verb":                 {Name: "リクエスト", Type: "string"},
	"_id":                  {Name: "内部ID", Type: "_id"},
	"auth":                 {Name: "ユーザー名", Type: "string"},
	"clientip":             {Name: "クライアントIP", Type: "string"},
	"clientip_geo":         {Name: "クライアント位置", Type: "geo"},
	"clientip_geo_city":    {Name: "クライアント都市", Type: "string"},
	"clientip_geo_country": {Name: "クライアントの国", Type: "string"},
	"clientip_geo_latlong": {Name: "クライアントの緯度経度", Type: "latlong"},
	"clientip_host":        {Name: "クライアントのホスト名", Type: "string"},
	"logsource":            {Name: "ログ送信元", Type: "string"},
	"priority":             {Name: "プライオリティー", Type: "number"},
	"severity":             {Name: "優先度", Type: "number"},
	"facility":             {Name: "ファシリティー", Type: "number"},
	"message":              {Name: "メッセージ", Type: "string"},
	"pid":                  {Name: "PID", Type: "number"},
	"program":              {Name: "プロセス名", Type: "string"},
	"delta":                {Name: "前ログとの時間差", Type: "number"},
	"facility_str":         {Name: "ファシリティー", Type: "string"},
	"severity_str":         {Name: "優先度", Type: "string"},
	"client":               {Name: "クライアントIP", Type: "string"},
	"client_geo":           {Name: "クライアント位置", Type: "geo"},
	"client_geo_city":      {Name: "クライアント都市", Type: "string"},
	"client_geo_country":   {Name: "クライアントの国", Type: "string"},
	"client_host":          {Name: "クライアントのホスト名", Type: "string"},
	"client_geo_latlong":   {Name: "クライアントの緯度経度", Type: "latlong"},
	"server":               {Name: "サーバーIP", Type: "string"},
	"server_geo":           {Name: "サーバー位置", Type: "geo"},
	"server_geo_city":      {Name: "サーバー都市", Type: "string"},
	"server_geo_country":   {Name: "サーバーの国", Type: "string"},
	"server_host":          {Name: "サーバーのホスト名", Type: "string"},
	"server_geo_latlong":   {Name: "サーバーの緯度経度", Type: "latlong"},
	"user":                 {Name: "ユーザーID", Type: "string"},
	"uri":                  {Name: "URI", Type: "string"},
	"email":                {Name: "メールアドレス", Type: "string"},
	"uuid":                 {Name: "UUID", Type: "string"},
	"tag":                  {Name: "タグ", Type: "string"},
	"load1m":               {Name: "1分間負荷", Type: "number"},
	"load5m":               {Name: "5分間負荷", Type: "number"},
	"load15m":              {Name: "15分間負荷", Type: "number"},
	"_None":                {Name: "項目なし", Type: "string"},
	"winEventID":           {Name: "イベントID", Type: "number"},
	"winEventRecordID":     {Name: "レコードID", Type: "number"},
	"winChannel":           {Name: "チャネル", Type: "number"},
	"winProvider":          {Name: "プロバイダー", Type: "number"},
	"winLevel":             {Name: "レベル", Type: "number"},
	"winUserID":            {Name: "ユーザーID", Type: "number"},
	"winComputer":          {Name: "コンピュータ名", Type: "string"},
	"score":                {Name: "検索スコア", Type: "number"},
	"anomalyScore":         {Name: "異常スコア", Type: "number"},
	"host":                 {Name: "ホスト", Type: "string"},
	"hostname":             {Name: "ホスト名", Type: "string"},
	"level":                {Name: "レベル", Type: "string"},
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
			Name: f + "(自動追加)",
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
		return "抽出パターンを変更できません"
	}
	extractorTypes[et.Key] = et
	return ""
}

// DeleteExtractorType : 抽出パターンを削除する
func (b *App) DeleteExtractorType(key string) string {
	if et, ok := extractorTypes[key]; ok && !et.CanEdit {
		return "抽出パターンを削除できません"
	}
	result, err := wails.MessageDialog(b.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         "抽出パターンの削除",
		Message:       "削除しますか?",
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
		return "フィールドタイプを保存できません"
	}
	fieldTypes[ft.Key] = ft
	return ""
}

// DeleteFieldType : フィールドタイプを削除する
func (b *App) DeleteFieldType(key string) string {
	if ft, ok := fieldTypes[key]; ok && !ft.CanEdit {
		return "フィールドタイプを削除できません"
	}
	result, err := wails.MessageDialog(b.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         "フィールドタイプの削除",
		Message:       "削除しますか?",
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
			DisplayName: "Yaml ファイル",
			Pattern:     "*.yaml",
		}},
	})
	if file == "" {
		return ""
	}
	if err != nil {
		OutLog("ExportLogTypes err=%v", err)
		return fmt.Sprintf("エクスポートできません。 err=%v", err)
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
		return fmt.Sprintf("エクスポートできません。 err=%v", err)
	}
	err = ioutil.WriteFile(file, d, 0660)
	if err != nil {
		return fmt.Sprintf("エクスポートできません。 err=%v", err)
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
			DisplayName: "Yaml ファイル",
			Pattern:     "*.yaml",
		}},
	})
	if file == "" {
		return ""
	}
	if err != nil {
		OutLog("importLogTypes err=%v", err)
		return "ファイルを選択できません err=" + err.Error()
	}
	d, err := ioutil.ReadFile(file)
	if err != nil {
		OutLog("importLogTypes file=%v err=%v", file, err)
		return "ファイルを読み込めません err=" + err.Error()
	}
	export := new(exportLogType)
	err = yaml.Unmarshal(d, &export)
	if err != nil {
		OutLog("importLogTypes err=%v", err)
		return "ファイルのフォーマットが違います err=" + err.Error()
	}
	etKeyMap := make(map[string]*ExtractorType)
	for _, et := range extractorTypes {
		etKeyMap[et.Key] = &et
	}
	for _, iet := range export.ExtractorTypes {
		if et, ok := etKeyMap[iet.Key]; ok && !et.CanEdit {
			return "組み込みと同じ定義があります。 name=" + iet.Name
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
