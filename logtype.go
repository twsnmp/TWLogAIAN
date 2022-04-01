package main

import (
	"io/ioutil"
	"strings"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v2"
)

// GetExtractorTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetExtractorTypes() []ExtractorType {
	ret := []ExtractorType{}
	ret = append(ret, extractorTypes...)
	ret = append(ret, b.importedExtractorTypes...)
	return ret
}

// GetFieldTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetFieldTypes() map[string]FieldType {
	ret := make(map[string]FieldType)
	for k, v := range fieldTypes {
		ret[k] = v
	}
	for k, v := range b.importedFieldTypes {
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
		Key:       "syslogBSD_NOPID",
		Name:      "syslog(BSD/PIDなし)",
		TimeField: "timestamp",
		Grok:      `%{SYSLOGTIMESTAMP:timestamp}\s+%{SYSLOGHOST:logsource}\s+%{NOTSPACE:tag}:\s+%{GREEDYDATA:message}`,
		View:      "syslog",
	},
	{
		Key:       "syslogBSD_PRI",
		Name:      "syslog(BSD/文字列PRI付き)",
		TimeField: "timestamp",
		Grok:      `%{SYSLOGTIMESTAMP:timestamp} %{SYSLOGHOST:logsource}\s+%{NOTSPACE:facility_str}\.%{NOTSPACE:severity_str}\s+%{SYSLOGPROG}:\s+%{GREEDYDATA:message}`,
		View:      "syslog",
	},
	{
		Key:       "syslogIETF",
		Name:      "syslog(IETF)",
		TimeField: "timestamp",
		Grok:      `%{TIMESTAMP_ISO8601:timestamp}\s+(?:%{SYSLOGFACILITY} )?%{SYSLOGHOST:logsource}\s+%{NOTSPACE:tag}:\s+%{GREEDYDATA:message}`,
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
	Key  string
	Name string
	Type string // number,string,latlong,time
	Unit string
}

var fieldTypes = map[string]FieldType{
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
		if _, ok := b.importedFieldTypes[f]; !ok {
			b.importedFieldTypes[f] = FieldType{
				Name: f + "(自動追加)",
				Type: t,
			}
		}
	}
}

type exportLogType struct {
	ExtractorTypes []ExtractorType
	FieldTypes     []FieldType
}

func (b *App) ExportLogTypes() error {
	file, err := wails.SaveFileDialog(b.ctx, wails.SaveDialogOptions{
		DefaultFilename:      "logtypes.yaml",
		CanCreateDirectories: false,
		Filters: []wails.FileFilter{{
			DisplayName: "Yaml ファイル",
			Pattern:     "*.yaml",
		}},
	})
	if err != nil {
		OutLog("ExportLogTypes err=%v", err)
		return err
	}
	export := exportLogType{}
	if b.config.Extractor == "custom" {
		d := time.Now().Format("20060102150405")
		export.ExtractorTypes = append(export.ExtractorTypes, ExtractorType{
			Key:       "custom_" + d,
			Name:      "カスタム" + d,
			Grok:      b.config.Grok,
			IPFields:  b.getIPFields(),
			MACFields: b.config.MACFields,
			TimeField: b.config.TimeField,
		})
	} else {
		export.ExtractorTypes = append(export.ExtractorTypes, extractorTypes...)
		export.ExtractorTypes = append(export.ExtractorTypes, b.importedExtractorTypes...)
		for k, e := range fieldTypes {
			e.Key = k
			export.FieldTypes = append(export.FieldTypes, e)
		}
	}
	for k, e := range b.importedFieldTypes {
		e.Key = k
		export.FieldTypes = append(export.FieldTypes, e)
	}
	d, err := yaml.Marshal(&export)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, d, 0660)
	if err != nil {
		return err
	}
	return nil
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
	if err != nil {
		OutLog("importLogTypes err=%v", err)
		return "ファイルを選択できません err=" + err.Error()
	}
	b.importedExtractorTypes = []ExtractorType{}
	b.importedFieldTypes = make(map[string]FieldType)
	d, err := ioutil.ReadFile(file)
	if err != nil {
		OutLog("importLogTypes err=%v", err)
		return "ファイルを読み込めません err=" + err.Error()
	}
	export := new(exportLogType)
	err = yaml.Unmarshal(d, &export)
	if err != nil {
		OutLog("importLogTypes err=%v", err)
		return "ファイルのフォーマットが違います err=" + err.Error()
	}
	b.importedExtractorTypes = append(b.importedExtractorTypes, export.ExtractorTypes...)
	for _, e := range export.FieldTypes {
		b.importedFieldTypes[e.Key] = e
	}
	return ""
}

// DeleteLogTypes : インポートしたログタイプを削除する
func (b *App) DeleteLogTypes() string {
	b.importedExtractorTypes = []ExtractorType{}
	b.importedFieldTypes = make(map[string]FieldType)
	return ""
}

// HasImportedLogTypes : インポートしたログタイプの有無を返す
func (b *App) HasImportedLogTypes() bool {
	return len(b.importedExtractorTypes) > 0
}
