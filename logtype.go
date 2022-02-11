package main

// GetExtractorTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetExtractorTypes() []ExtractorType {
	return extractorTypes
}

// GetFieldTypes : 定義済みのログタイプのリスト情報を提供する
func (b *App) GetFieldTypes() map[string]FieldType {
	return fieldTypes
}

// ExtractorType : ログからデータを取得するパターン定義
type ExtractorType struct {
	Key       string
	Name      string
	Grok      string
	TimeField string
	IP        bool
	IPFields  string
	View      string
}

var extractorTypes = []ExtractorType{
	{
		Key:       "syslog",
		Name:      "syslog",
		TimeField: "timestamp",
		Grok:      `%{SYSLOGBASE} %{GREEDYDATA:message}`,
		IP:        false,
		View:      "syslog",
	},
	{
		Key:       "apacheCommon",
		Name:      "Apache(Common)",
		TimeField: "timestamp",
		Grok:      `%{COMMONAPACHELOG}`,
		IP:        true,
		IPFields:  "clientip",
		View:      "access",
	},
	{
		Key:       "apacheConbined",
		Name:      "Apache(Conbined)",
		TimeField: "timestamp",
		Grok:      `%{COMBINEDAPACHELOG}`,
		IP:        true,
		IPFields:  "clientip",
		View:      "access",
	},
}

type FieldType struct {
	Name string
	Type string // number,string,latlong,time
	Unit string
}

var fieldTypes = map[string]FieldType{
	"_all":                 {Name: "ログの行全体", Type: "string"},
	"httpversion":          {Name: "HTTPバージョン", Type: "string"},
	"ident":                {Name: "識別子", Type: "string"},
	"request":              {Name: "パス", Type: "string"},
	"response":             {Name: "応答コード", Type: "number"},
	"agent":                {Name: "ユーザーエージェント", Type: "string"},
	"bytes":                {Name: "サイズ", Type: "number"},
	"clientip_geo_latlong": {Name: "クラアンと位置", Type: "latlong"},
	"timestamp":            {Name: "タイムスタンプ", Type: "string"},
	"rawrequest":           {Name: "元のリクエスト", Type: "string"},
	"referrer":             {Name: "リファラー", Type: "string"},
	"time":                 {Name: "日時", Type: "string"},
	"verb":                 {Name: "リクエスト", Type: "string"},
	"_id":                  {Name: "内部ID", Type: "string"},
	"auth":                 {Name: "ユーザー名", Type: "string"},
	"clientip":             {Name: "クライアントIP", Type: "string"},
	"clientip_geo_city":    {Name: "クライアント都市", Type: "string"},
	"clientip_geo_country": {Name: "クライアントの国", Type: "string"},
	"clientip_host":        {Name: "クライアントのホスト名", Type: "string"},
}
