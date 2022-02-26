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
	{
		Key:      "SSHLOGIN",
		Name:     "SSHのログイン",
		Grok:     `%{NOTSPACE:stat} (password|publickey) for( invalid user | )%{USER:user} from %{IP:clientip}`,
		IPFields: "clientip",
	},
	{
		Key:       "DEVICE",
		Name:      "デバイス情報(ip)",
		Grok:      `mac=%{MAC:mac}.+ip=%{IP:ip}`,
		IPFields:  "ip",
		MACFields: "mac",
	},
	{
		Key:       "DEVICER",
		Name:      "デバイス情報(mac)",
		Grok:      `ip=%{IP:ip}.+mac=%{MAC:mac}`,
		IPFields:  "ip",
		MACFields: "mac",
	},
	{
		Key:  "WELFFLOW",
		Name: "WELFフロー",
		Grok: `src=%{IP:src}:%{BASE10NUM:sport}:.+dst=%{IP:dst}:%{BASE10NUM:dport}:.+proto=%{WORD:prot}.+sent=%{BASE10NUM:sent}.+rcvd=%{BASE10NUM:rcvd}.+spkt=%{BASE10NUM:spkt}.+rpkt=%{BASE10NUM:rpkt}`,
	},
	{
		Key:  "UPTIME",
		Name: "負荷(uptime)",
		Grok: `load average: %{BASE10NUM:load1m}, %{BASE10NUM:load5m}, %{BASE10NUM:load15m}`,
	},
	{
		Key:  "TWPCAP_STATS",
		Name: "TWPCAPの統計情報",
		Grok: `type=Stats,total=%{BASE10NUM:total},count=%{BASE10NUM:count},ps=%{BASE10NUM:ps}`,
	},
	{
		Key:       "TWPCAP_IPTOMAC",
		Name:      "TWPCAPのIPとMACアドレス",
		Grok:      `type=IPToMAC,ip=%{IP:ip},mac=%{MAC:mac},count=%{BASE10NUM:count},change=%{BASE10NUM:chnage},dhcp=%{BASE10NUM:dhcp}`,
		IPFields:  "ip",
		MACFields: "mac",
	},
	{
		Key:       "TWPCAP_DNS",
		Name:      "TWPCAPのDNS問い合わせ",
		Grok:      `type=DNS,sv=%{IP:server},DNSType=%{WORD:dnsType},Name=%{IPORHOST:name},count=%{BASE10NUM:count},change=%{BASE10NUM:chnage},lcl=%{IP:lastIP},lMAC=%{MAC:lastMAC}`,
		IPFields:  "server,lastIP",
		MACFields: "lastMAC",
	},
	{
		Key:      "TWPCAP_DHCP",
		Name:     "TWPCAPのDHCPサーバー情報",
		Grok:     `type=DHCP,sv=%{IP:server},count=%{BASE10NUM:count},offer=%{BASE10NUM:offer},ack=%{BASE10NUM:ack},nak=%{BASE10NUM:nak}`,
		IPFields: "server",
	},
	{
		Key:      "TWPCAP_NTP",
		Name:     "TWPCAPのNTPサーバー情報",
		Grok:     `type=NTP,sv=%{IP:server},count=%{BASE10NUM:count},change=%{BASE10NUM:change},lcl=%{IP:client},version=%{BASE10NUM:version},stratum=%{BASE10NUM:stratum},refid=%{WORD:refid}`,
		IPFields: "client,server",
	},
	{
		Key:      "TWPCAP_RADIUS",
		Name:     "TWPCAPのRADIUS通信情報",
		Grok:     `type=RADIUS,cl=%{IP:client},sv=%{IP:server},count=%{BASE10NUM:count},req=%{BASE10NUM:request},accept=%{BASE10NUM:accept},reject=%{BASE10NUM:reject},challenge=%{BASE10NUM:challenge}`,
		IPFields: "client,server",
	},
	{
		Key:      "TWPCAP_TLSFlow",
		Name:     "TWPCAPのTLS通信情報",
		Grok:     `type=TLSFlow,cl=%{IP:client},sv=%{IP:server},serv=%{WORD:service},count=%{BASE10NUM:count},handshake=%{BASE10NUM:handshake},alert=%{BASE10NUM:alert},minver=%{DATA:minver},maxver=%{DATA:maxver},cipher=%{DATA:cipher},ft=`,
		IPFields: "client,server",
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
	"clientip_geo_latlong": {Name: "クライアントの緯度経度", Type: "latlong"},
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
