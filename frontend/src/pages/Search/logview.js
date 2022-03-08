// Log を表示するための処理
import * as echarts from 'echarts';
import { html,h } from "gridjs";

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

const formatLevel = (level) => {
  switch (level) {
  case "error":
    return html(`<div class="color-fg-danger">エラー</div>`);
  case "warn":
    return html(`<div class="color-fg-attention">注意</div>`);
  }
  return html(`<div class="color-fg-default">正常</div>`);
}

const selectLogMap = new Map()

export const getSelectedLogs = () => {
  return Array.from(selectLogMap).join("\n");
}

export const clearSelectedLogs = () => {
  selectLogMap.clear();
}

const columnsTimeOnly = [
  {
    id: "select",
    name: "",
    width: "5%",
    sort: false,
    formatter: (cell, row) => {
      return h('input', {
        type: 'checkbox',
        onChange: () => {
          const key = row.cells[3].data
          if (selectLogMap.has(key)) {
            selectLogMap.delete(key); 
          } else {
            selectLogMap.set(key,true);
          }
        }
      });
    },
  },{
    id: "level",
    name: "レベル",
    width: "10%",
    formatter: (cell) => formatLevel(cell),
  },{
    id: "timestamp",
    name: "日時",
    width: "20%",
    formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}'),
    convert: true,
  },{
    name: "ログ",
    width: "65%",
  },
];

const getTimeOnlyLogData = (r, filter) => {
  const d = [];
  r.Logs.forEach((l) => {
    if(filter && filter.st) {
      if (l.Time < filter.st || l.Time > filter.et) {
        return
      }
    }
    d.push(["",getLogLevel(l),l.Time, l.All]);
  });
  return d;
}


const columnsSyslog = [
  {
    id: "level",
    name: "レベル",
    width: "10%",
    formatter: (cell) => formatLevel(cell),
  },{
    id: "timestamp",
    name: "日時",
    width: "15%",
    formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}'),
    convert: true,
  },{
    id: "logsrc",
    name: "送信元",
    width: "15%",
  },{
    id: "tag",
    name: "タグ",
    width: "20%",
  },{
    id: "message",
    name: "メッセージ",
    width: "40%",
  },
];

const columnsAccessLog = [
  {
    id: "code",
    name: "応答",
    width: "8%",
    formatter: (cell) => formatCode(cell),
  },{
    id: "timestamp",
    name: "日時",
    width: "15%",
    formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    convert: true,
  },{
    id: "req",
    name: "リクエスト",
    width: "10%",
  },{
    id: "size",
    name: "サイズ",
    width: "8%",
  },{
    id: "client",
    name: "アクセス元",
    width: "25%",
  },{
    id: "ISO",
    name: "国",
    width: "8%",
  },{
    id: "path",
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

const getAccessLogData = (r, filter) =>{
  const d = [];
  r.Logs.forEach((l) => {
    if(filter && filter.st) {
      if (l.Time < filter.st || l.Time > filter.et) {
        return
      }
    }
    d.push([
      l.KeyValue.response,
      l.Time,
      l.KeyValue.verb,
      l.KeyValue.bytes,
      l.KeyValue.clientip_host ? l.KeyValue.clientip + "(" + l.KeyValue.clientip_host +")" : l.KeyValue.clientip,
      l.KeyValue.clientip_geo_country || "",
      l.KeyValue.request,
    ]);
  });
  return d
}

export const getLogLevel = (l) => {
  let suverity = l.KeyValue.suverity || l.KeyValue.priority;
  if (suverity && suverity != "") {
    // 数値のsuverityを優先する
    suverity %= 8
    return  suverity < 4 ? "error" : suverity == 4 ? "warn" : "normal";
  }
  const level = l.KeyValue.suverity_str || l.KeyValue.level || l.All;
  if (/(alert|error|crit|fatal|emerg|err )/i.test(level)) {
    return "error";
  }
  if (/warn/i.test(level)) {
    return "warn";
  }
  return "normal";
}

const getSyslogData = (r,filter) => {
  const d = [];
  r.Logs.forEach((l) => {
    if(filter && filter.st) {
      if (l.Time < filter.st || l.Time > filter.et) {
        return
      }
    }
    const message = l.KeyValue.message || "";
    const pid = l.KeyValue.pid || "";
    const tag =  l.KeyValue.tag || ((l.KeyValue.program || "") + (pid ? "[" + pid + "]" : ""));
    const src = l.KeyValue.logsource || "";
    const level = getLogLevel(l);
    d.push([level,l.Time,src,tag,message ]);
  });
  return d;
}

export const getLogData = (r,view,filter) => {
  switch (view) {
  case "syslog":
    return getSyslogData(r,filter);
  case "access":
    return getAccessLogData(r,filter);
  }
  return getTimeOnlyLogData(r,filter);
}
