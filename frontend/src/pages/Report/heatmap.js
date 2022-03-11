import * as echarts from "echarts";
import { getFieldName } from "../../js/define";
import * as ecStat from "echarts-stat";

let chart;

const getValue = (calcMode,values) => {
  let v = 0.0;
  switch (calcMode) {
    case "sum":
      v = ecStat.statistics.sum(values);
      break;
    case "mean":
      v = ecStat.statistics.mean(values);
      break;
    case "median":
      v = ecStat.statistics.median(values);
      break;
    case "variance":
      v = ecStat.statistics.sampleVariance(values);
      break;
  }
  return v;
}

export const showHeatmap = (div, logs, field, sumMode, calcMode, dark) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), dark ? "dark" : null);
  const hours = [
    "0時",
    "1時",
    "2時",
    "3時",
    "4時",
    "5時",
    "6時",
    "7時",
    "8時",
    "9時",
    "10時",
    "11時",
    "12時",
    "13時",
    "14時",
    "15時",
    "16時",
    "17時",
    "18時",
    "19時",
    "20時",
    "21時",
    "22時",
    "23時",
  ];
  const option = {
    title: {
      show: false,
    },
    grid: {
      left: "10%",
      right: "5%",
      top: 30,
      buttom: 0,
    },
    toolbox: {
      feature: {
        dataZoom: {},
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: "item",
      formatter(params) {
        return (
          params.name +
          " " +
          params.data[1] +
          "時 : " +
          params.data[2].toFixed(1)
        );
      },
      axisPointer: {
        type: "shadow",
      },
    },
    xAxis: {
      type: "category",
      name: "日付",
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 10,
        margin: 2,
      },
      data: [],
    },
    yAxis: {
      type: "category",
      name: "時間帯",
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 10,
        margin: 2,
      },
      data: hours,
    },
    visualMap: {
      min: Infinity,
      max: -Infinity,
      textStyle: {
        fontSize: 8,
      },
      calculable: true,
      realtime: false,
      inRange: {
        color: [
          "#313695",
          "#4575b4",
          "#74add1",
          "#abd9e9",
          "#e0f3f8",
          "#ffffbf",
          "#fee090",
          "#fdae61",
          "#f46d43",
          "#d73027",
          "#a50026",
        ],
      },
    },
    series: [
      {
        name: field ? getFieldName(field) : "回数",
        type: "heatmap",
        data: [],
        emphasis: {
          itemStyle: {
            borderWidth: 1,
          },
        },
        progressive: 1000,
        animation: false,
      },
    ],
  };
  const data = [];
  if (sumMode == "day") {
    let nD = 0;
    let nH = -1;
    let x = -1;
    let day;
    const values = [];
    logs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000));
      day = echarts.time.format(t, "{yyyy}/{MM}/{dd}");
      if (field) {
        values.push(l.KeyValue[field] || 0.0);
      } else {
        values.push(1);
      }
      if (nD !== t.getDate()) {
        option.xAxis.data.push(day);
        nD = t.getDate();
        x++;
      }
      if (nH != t.getHours()) {
        if (nH != -1) {
          const v = getValue(calcMode,values);
          option.series[0].data.push([x, nH, v]);
          data.push([day, option.yAxis.data[nH], v]);
          if (option.visualMap.min > v) {
            option.visualMap.min = v;
          }
          if (option.visualMap.max < v) {
            option.visualMap.max = v;
          }
          values.length = 0;
        }
        nH = t.getHours();
      }
    });
    if (values.length > 0 ){
      const v = getValue(calcMode,values);
      option.series[0].data.push([x, nH, v]);
      data.push([day, option.yAxis.data[nH], v]);
      if (option.visualMap.min > v) {
        option.visualMap.min = v;
      }
      if (option.visualMap.max < v) {
        option.visualMap.max = v;
      }
    }
  } else {
    option.xAxis.name = "曜日";
    option.xAxis.data = ["日", "月", "火", "水", "木", "金", "土"];
    const values = [];
    for (let x = 0; x < 7; x++) {
      const ent = [];
      for (let y = 0; y < 24; y++) {
        ent.push([]);
      }
      values.push(ent);
    }

    logs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000));
      const x = t.getDay();
      const y = t.getHours();
      if (field) {
        values[x][y].push(l.KeyValue[field] || 0.0);
      } else {
        values[x][y].push(1);
      }
    });
    for (let x = 0; x < 7; x++) {
      for (let y = 0; y < 24; y++) {
        const v = getValue(calcMode,values[x][y]);
        if (option.visualMap.min > v) {
          option.visualMap.min = v;
        }
        if (option.visualMap.max < v) {
          option.visualMap.max = v;
        }
        option.series[0].data.push([x, y, v]);
        data.push([option.xAxis.data[x], option.yAxis.data[y], v]);
      }
    }
  }
  chart.setOption(option);
  chart.resize();
  return data;
};

export const resizeHeatmap = () => {
  chart.resize();
};

export const getHeatmapImage = () => {
  if (chart) {
    return chart.getDataURL({ type: "png" });
  }
  return [];
};
