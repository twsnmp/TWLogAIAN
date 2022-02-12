// Log を表示するための処理
import * as echarts from 'echarts';
import { html } from "gridjs";

const formatCode = (code) => {
  if (code < 300) {
    return html(`<div class="color-fg-default">${code}</div>`);
  } else if (code < 400) {
    return html(`<div class="color-fg-attention">${code}</div>`);
  } else if (code < 500) {
    return html(`<div class="color-fg-danger">${code}</div>`);
  }
  return html(`<div class="color-fg-danger-emphasis">${code}</div>`);
}

const columnsTimeOnly = [
  {
    name: "スコア",
    width: "10%",
    formatter: (cell) => Number.parseFloat(cell).toFixed(2),
  },{
    name: "日時",
    width: "20%",
    formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}'),
  },{
    name: "ログ",
    width: "70%",
  },
];

const columnsSyslog = [
  {
    name: "レベル",
    width: "10%",
  },{
    name: "日時",
    width: "15%",
    formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}'),
  },{
    name: "送信元",
    width: "15%",
  },{
    name: "タグ",
    width: "15%",
  },{
    name: "PID",
    width: "5%",
  },{
    name: "メッセージ",
    width: "40%",
  },
];

const columnsAccessLog = [
  {
    name: "応答",
    width: "8%",
    formatter: (cell) => formatCode(cell),
  },{
    name: "日時",
    width: "15%",
    formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
  },{
    name: "リクエスト",
    width: "10%",
  },{
    name: "サイズ",
    width: "8%",
  },{
    name: "アクセス元",
    width: "25%",
  },{
    name: "国",
    width: "8%",
  },{
    name: "パス",
    width: "26%",
  },
];

export const getLogColums = (view) => {
  switch(view) {
    case "syslog":
        return columnsSyslog;
    case "access":
      return columnsAccessLog;
  }
  return columnsTimeOnly;
}

const getAccessLogData = (r) =>{
  const d = [];
  r.Logs.forEach((l) => {
    let cl = l.KeyValue.clientip;
    if (l.KeyValue.clientip_host) {
      cl += "(" + l.KeyValue.clientip_host +")"
    }
    let country = l.KeyValue.clientip_geo ? l.KeyValue.clientip_geo.Country :"";
    d.push([
      l.KeyValue.response,
      l.Time,
      l.KeyValue.verb,
      l.KeyValue.bytes,
      cl,
      country,
      l.KeyValue.request,
    ]);
  });
  return d
}

const getTimeOnlyLogData = (r) => {
  const d = [];
  r.Logs.forEach((l) => {
    d.push([l.Score, l.Time, l.All]);
  });
  return d;
}

const getSyslogData = (r) => {
  const d = [];
  r.Logs.forEach((l) => {
    const message = l.KeyValue.message || "";
    const pid = l.KeyValue.pid || "";
    const tag = l.KeyValue.program || "";
    const src = l.KeyValue.logsource || "";
    const pri = l.KeyValue.priority || "";
    d.push([pri,l.Time,src,tag,pid,message ]);
  });
  return d;
}

export const getLogData = (r) => {
  switch (r.View) {
  case "syslog":
    return getSyslogData(r);
  case "access":
    return getAccessLogData(r);
  }
  return getTimeOnlyLogData(r);
}

