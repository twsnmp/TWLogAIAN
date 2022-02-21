import * as echarts from "echarts";
import * as ecStat from "echarts-stat";
import { getFieldName } from "../../js/define";

let chart;

const calcRegression = (logs,field) => {
  const data = []
  logs.forEach((l) => {
    data.push([l.Time /(1000 * 1000 * 1000),l.KeyValue[field] || 0.0 ])
  })
  return ecStat.regression('linear', data);
}

export const showTimeChart = (
  div,
  logs,
  numField1,
  numField2,
  chartType,
  dark
) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), dark ? "dark" : null);
  const option = {
    title: {
      show: false,
    },
    toolbox: {},
    dataZoom: [{}],
    legend: {
      data: [getFieldName(numField1)],
      textStyle: {
        fontSize: 10,
      },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: {
        type: "shadow",
      },
    },
    grid: {
      left: "10%",
      right: "10%",
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: "time",
      name: "日時",
      axisLabel: {
        fontSize: "8px",
        formatter: (value, index) => {
          const date = new Date(value);
          return echarts.time.format(date, "{MM}/{dd} {HH}:{mm}");
        },
      },
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: [
      {
        type: "value",
        name: getFieldName(numField1),
        nameTextStyle: {
          fontSize: 10,
          margin: 2,
        },
        axisLabel: {
          fontSize: 8,
          margin: 2,
        },
      },
    ],
    series: [
      {
        name: getFieldName(numField1),
        type: "line",
        large: true,
        symbol: "none",
        data: [],
      },
    ],
  };
  if (numField2 != "") {
    const name = getFieldName(numField2);
    option.series.push({
      name,
      type: "line",
      large: true,
      yAxisIndex: 1,
      data: [],
    });
    option.yAxis.push({
      type: "value",
      name,
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 8,
        margin: 2,
      },
    });
    option.legend.data.push(name);
  }
  if( chartType == "forcast") {
    const reg1 = calcRegression(logs,numField1);
    let reg2;
    if (numField2 != "") {
      reg2 = calcRegression(logs,numField2);
    }
    const sh = Math.floor(Date.now() / ( 3600 * 1000))
    for (let h = sh; h < sh + (24*30); h++) {
      const x = h * 3600 * 1000
      const t = new Date(x)
      option.series[0].data.push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        value: [t, reg1.parameter.intercept + reg1.parameter.gradient * x],
      });
      if (reg2) {
        option.series[1].data.push({
          name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
          value: [t, reg2.parameter.intercept + reg2.parameter.gradient * x],
        });
      }
    }
  } else {
    logs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000));
      const name = echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
      option.series[0].data.push({
        name,
        value: [t, l.KeyValue[numField1] || 0.0],
      });
      if (numField2 != "") {
        option.series[1].data.push({
          name,
          value: [t, l.KeyValue[numField2] || 0.0],
        });
      }
    });
  }
  chart.setOption(option);
  chart.resize();
};

export const resizeTimeChart = () => {
  chart.resize();
};
