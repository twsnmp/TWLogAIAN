import * as echarts from "echarts";
import * as ecStat from "echarts-stat";
import { getFieldName, getFieldUnit } from "../../js/define";
import { _,unwrapFunctionStore } from 'svelte-i18n';

const $_ = unwrapFunctionStore(_);

let chart;

const calcRegression = (logs, field, chartType) => {
  const data = [];
  logs.forEach((l) => {
    const val = (l.KeyValue && l.KeyValue[field] != null) ? Number(l.KeyValue[field]) : 0.0;
    data.push([l.Time / (1000 * 1000), isNaN(val) ? 0.0 : val]);
  });
  return ecStat.regression(chartType, data);
}

const setChartData = (series,t, values) => {
  const data = [t.getTime()* 1000 * 1000 ];
  const name = echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
  const mean = ecStat.statistics.mean(values);
  series[0].data.push({ 
    name,
    value: [t, mean],
  });
  data.push(mean);
  const max = ecStat.statistics.max(values);
  series[1].data.push({ 
    name,
    value: [t, max],
  });
  data.push(max);
  const min = ecStat.statistics.min(values);
  series[2].data.push({ 
    name,
    value: [t, min],
  });
  data.push(min);
  const median = ecStat.statistics.median(values);
  series[3].data.push({ 
    name,
    value: [t, median],
  });
  data.push(median);
  const variance = ecStat.statistics.sampleVariance(values);
  series[4].data.push({ 
    name,
    value: [t, variance],
  });
  data.push(variance);
  return data;
}

export const showTimeChart = (div, logs, field, chartType, dark) => {
  if (chart) {
    chart.dispose();
  }
  const validLogs = (logs || []).filter(l => l && l.Time > 946684800000000000);
  if (validLogs.length === 0) {
    return [];
  }
  const el = typeof div === "string" ? document.getElementById(div) : div;
  if (!el) {
    return [];
  }
  chart = echarts.init(el, dark ? "dark" : null);
  const option = {
    title: {
      show: false,
    },
    dataZoom: [
      {
        type: "inside",
        xAxisIndex: 0,
      },
      {
        type: "slider",
        xAxisIndex: 0,
        bottom: 8,
        height: 20,
      },
    ],
    legend: {
      top: 6,
      left: "center",
      data: [getFieldName(field)],
      textStyle: {
        fontSize: 11,
        color: dark ? "#c9d1d9" : "#24292f",
      },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: {
        type: "shadow",
      },
    },
    grid: {
      left: "8%",
      right: "8%",
      top: 40,
      bottom: 60,
    },
    xAxis: {
      type: "time",
      name: $_("Js.Time"),
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
        name: getFieldName(field) + " " + getFieldUnit(field),
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
        name: getFieldName(field),
        type: "line",
        large: true,
        symbol: "none",
        data: [],
      },
    ],
  };
  let data = [];
  if (chartType == "1h" || chartType == "1m" ) {
    option.series[0].name = $_("Js.Mean");
    option.series.push({
      name: $_("Js.Max"),
      type: "line",
      large: true,
      data: [],
    });
    option.series.push({
      name: $_("Js.Min"),
      type: "line",
      large: true,
      data: [],
    });
    option.series.push({
      name: $_("Js.Median"),
      type: "line",
      large: true,
      data: [],
    });
    option.series.push({
      name: $_("Js.Variant"),
      type: "line",
      large: true,
      yAxisIndex: 1,
      data: [],
    });
    option.yAxis.push({
      type: "value",
      name: $_("Js.Variant"),
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 8,
        margin: 2,
      },
    });
    option.legend.data[0]= $_("Js.Mean");
    option.legend.data.push($_("Js.Max"));
    option.legend.data.push($_("Js.Min"));
    option.legend.data.push($_("Js.Median"));
    option.legend.data.push($_("Js.Variant"));
    let tS = -1;
    const values = [];
    const dt = chartType == "1h" ? 3600 * 1000 : 60 * 1000;
    validLogs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000));
      let tC = Math.floor(t.getTime() / dt);
      if (tS != tC) {
        if (tS > 0 ) {
          if (values.length > 0 ){
            tS++;
            data.push(setChartData(option.series,new Date(tS * dt),values));
            values.length = 0;
            if (tC - tS > 100) {
              for (let i = 0; i < 5 && tS < tC; i++) {
                tS++;
                setChartData(option.series,new Date(tS * dt),[0,0,0,0]);
              }
              if (tS < tC - 5) {
                tS = tC - 5;
              }
              while( tS < tC) {
                tS++;
                setChartData(option.series,new Date(tS * dt),[0,0,0,0]);
              }
            } else {
              while( tS < tC) {
                tS++;
                setChartData(option.series,new Date(tS * dt),[0,0,0,0]);
              }
            }
          }
        }
        tS = tC;
      }
      const val = (l.KeyValue && l.KeyValue[field] != null) ? Number(l.KeyValue[field]) : 0.0;
      values.push(isNaN(val) ? 0.0 : val);
    });
    if (values.length > 0 ){
      tS++;
      data.push(setChartData(option.series,new Date(tS * dt),values));
    }
  } else if (chartType != "") {
    option.series[0] = {
      name: getFieldName(field),
      type: 'scatter',
      label: {
          emphasis: {
              show: true
          }
      },
      data: [],
    }
    validLogs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000));
      const val = (l.KeyValue && l.KeyValue[field] != null) ? Number(l.KeyValue[field]) : 0.0;
      option.series[0].data.push([t, isNaN(val) ? 0.0 : val]);
    });
    const reg = calcRegression(validLogs, field, chartType);
    option.legend.data.push($_('Js.Regression') +'('+ reg.expression +")");
    option.series.push({
        name: $_('Js.Regression') +'('+ reg.expression +")",
        type: 'line',
        showSymbol: false,
        data: reg.points,
        markPoint: {
            itemStyle: {
                normal: {
                    color: 'transparent'
                }
            },
            label: {
                normal: {
                    show: true,
                    formatter: reg.expression,
                    textStyle: {
                        color: '#333',
                        fontSize: 12
                    }
                }
            },
            data: [{
                coord: reg.points[reg.points.length - 1]
            }]
        }
      });
      data = reg.points;
  } else {
    validLogs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000));
      const name = echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
      const val = (l.KeyValue && l.KeyValue[field] != null) ? Number(l.KeyValue[field]) : 0.0;
      const numVal = isNaN(val) ? 0.0 : val;
      option.series[0].data.push({
        name,
        value: [t, numVal],
      });
      data.push([l.Time, numVal]);
    });
  }
  chart.setOption(option, true);
  chart.resize();
  return data;
};

export const resizeTimeChart = () => {
  if (chart) {
    chart.resize();
  }
};

export const disposeTimeChart = () => {
  if (chart) {
    chart.dispose();
    chart = undefined;
  }
};

export const getTimeChartImage = () => {
  if (chart) {
    return chart.getDataURL({ type: "png" });
  }
  return [];
};
