package main

// GetExtractorTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetExtractorTypes() []ExtractorType {
	return extractorTypes
}

// GetFieldTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetFieldTypes() map[string]*FieldType {
	return fieldTypes
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
	"priority":             {Name: "プライオリティー", Type: "string"},
	"logsource":            {Name: "ログ送信元", Type: "string"},
	"message":              {Name: "メッセージ", Type: "string"},
	"facility":             {Name: "ファシリティー", Type: "string"},
	"pid":                  {Name: "PID", Type: "number"},
	"program":              {Name: "プロセス名", Type: "string"},
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
			setFieldType(f+"_latlong", "string")
		}
	}
}

func setFieldType(f, t string) {
	if _, ok := fieldTypes[f]; !ok {
		fieldTypes[f] = &FieldType{
			Name: f + "(自動追加)",
			Type: t,
		}
	}
}
