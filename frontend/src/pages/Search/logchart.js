import * as echarts from "echarts";
import { getLogLevel } from "./logview";

let chart;

const baseOption = {
  title: {
    show: false,
  },
  toolbox: {
    feature: {
      dataZoom: {},
    },
  },
  dataZoom: [{}],
  tooltip: {
    trigger: "axis",
    axisPointer: {
      type: "shadow",
    },
  },
  grid: {
    left: "8%",
    right: "8%",
    top: 10,
    buttom: 0,
  },
  xAxis: {
    type: "time",
    name: "Time",
    axisLabel: {
      fontSize: "8px",
      formatter: (value, index) => {
        const date = new Date(value);
        return echarts.time.format(date, "{MM}/{dd} {HH}:{mm}");
      },
    },
    nameTextStyle: {
      fontSize: 8,
      margin: 2,
    },
    splitLine: {
      show: false,
    },
  },
  yAxis: {
    type: "value",
    name: "Count",
    nameTextStyle: {
      fontSize: 8,
      margin: 2,
    },
    axisLabel: {
      fontSize: 8,
      margin: 2,
    },
  },
};


const addMultiChartData = (data, count, ctm, newCtm) => {
  let t = new Date(ctm * 60 * 1000);
  for (const k in count) {
    data[k].push({
      name: echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"),
      value: [t, count[k]],
    });
  }
  ctm++;
  for (; ctm < newCtm; ctm++) {
    t = new Date(ctm * 60 * 1000);
    for (const k in count) {
      data[k].push({
        name: echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"),
        value: [t, 0],
      });
    }
  }
  return ctm;
};

const getChartLogLevel = (l) => {
  const code = l.KeyValue.response;
  if (code) {
    return code < 300 ? "normal" : code < 400 ? "warn" : "error";
  }
  return getLogLevel(l);
};

export const showLogChart = (div, logs, dark, cb) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), dark ? "dark" : "");
  chart.setOption(baseOption);
  const data = {
    normal: [],
    warn: [],
    error: [],
  };
  const count = {
    normal: 0,
    warn: 0,
    error: 0,
  };
  let ctm;
  let st = Infinity;
  let lt = 0;
  logs.forEach((l) => {
    const lvl = getChartLogLevel(l);
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60));
    if (!ctm) {
      ctm = newCtm;
    }
    if (ctm !== newCtm) {
      ctm = addMultiChartData(data, count, ctm, newCtm);
      for (const k in count) {
        count[k] = 0;
      }
    }
    if (st > l.Time) {
      st = l.Time;
    }
    if (lt < l.Time) {
      lt = l.Time;
    }
    count[lvl]++;
  });
  addMultiChartData(data, count, ctm, ctm + 1);
  chart.setOption({
    grid: {
      left: "2%",
      right: "5%",
      top: 50,
      buttom: 0,
    },
    series: [
      {
        name: "正常",
        type: "bar",
        color: "RGB(14,80,209)",
        stack: "count",
        large: true,
        data: data.normal,
      },
      {
        name: "注意",
        type: "bar",
        color: "RGB(255,248,185)",
        stack: "count",
        large: true,
        data: data.warn,
      },
      {
        name: "エラー",
        type: "bar",
        color: "RGB(194,11,35)",
        stack: "count",
        large: true,
        data: data.error,
      },
    ],
    legend: {
      textStyle: {
        fontSize: 10,
      },
      data: ["正常", "注意", "エラー"],
    },
  });
  chart.resize();
  if (cb) {
    chart.on("datazoom", (e) => {
      if (e.batch && e.batch.length === 2) {
        if (e.batch[0].startValue) {
          // Select ZOOM
          cb(
            e.batch[0].startValue * 1000 * 1000,
            e.batch[0].endValue * 1000 * 1000
          );
        } else if( e.batch[0].end == 100 ) {
          // Reset ZOOM
          cb(false,false);
        }
      } else if (e.start !== undefined && e.end !== undefined) {
        // Scroll ZOOM
        cb(st + (lt - st) * (e.start / 100), st + (lt - st) * (e.end / 100));
      }
    });
  }
};

export const resizeLogChart = () => {
  if (chart) {
    chart.resize();
  }
};

export const getLogChartImage = () => {
  if (chart) {
    return chart.getDataURL({ type: "png" });
  }
  return [];
};
