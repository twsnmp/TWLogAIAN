import * as echarts from "echarts";
import * as ecStat from "echarts-stat";
import { getFieldName, getFieldUnit } from "../../js/define";

let chart;

const calcRegression = (logs, field, chartType) => {
  const data = [];
  logs.forEach((l) => {
    data.push([l.Time / (1000 * 1000), l.KeyValue[field] || 0.0]);
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
  chart = echarts.init(document.getElementById(div), dark ? "dark" : null);
  const option = {
    title: {
      show: false,
    },
    toolbox: {},
    dataZoom: [{}],
    legend: {
      data: [getFieldName(field)],
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
      name: "Time",
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
    option.series[0].name = "Mean";
    option.series.push({
      name: "Max",
      type: "line",
      large: true,
      data: [],
    });
    option.series.push({
      name: "Min",
      type: "line",
      large: true,
      data: [],
    });
    option.series.push({
      name: "Median",
      type: "line",
      large: true,
      data: [],
    });
    option.series.push({
      name: "Variant",
      type: "line",
      large: true,
      yAxisIndex: 1,
      data: [],
    });
    option.yAxis.push({
      type: "value",
      name: "Variant",
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 8,
        margin: 2,
      },
    });
    option.legend.data[0]= "Mean";
    option.legend.data.push("Max");
    option.legend.data.push("Min");
    option.legend.data.push("Median");
    option.legend.data.push("Variant");
    let tS = -1;
    const values = [];
    const dt = chartType == "1h" ? 3600 * 1000 : 60 * 1000;
    logs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000));
      let tC = Math.floor(t.getTime() / dt);
      if (tS != tC) {
        if (tS > 0 ) {
          if (values.length > 0 ){
            tS++;
            data.push(setChartData(option.series,new Date(tS * dt),values));
            values.length = 0;
            while( tS < tC) {
              tS++;
              setChartData(option.series,new Date(tS * dt),[0,0,0,0]);
            }
          }
        }
        tS = tC;
      }
      values.push(l.KeyValue[field] || 0.0);
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
    logs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000));
      option.series[0].data.push([t, l.KeyValue[field] || 0.0]);
    });
    const reg = calcRegression(logs, field, chartType);
    option.legend.data.push('Regression('+ reg.expression +")");
    option.series.push({
        name: 'Regression('+ reg.expression +")",
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
    logs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000));
      const name = echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
      option.series[0].data.push({
        name,
        value: [t, l.KeyValue[field] || 0.0],
      });
      data.push([l.Time,l.KeyValue[field] || 0.0]);
    });
  }
  chart.setOption(option);
  chart.resize();
  return data;
};

export const resizeTimeChart = () => {
  chart.resize();
};

export const getTimeChartImage = () => {
  if (chart) {
    return chart.getDataURL({ type: "png" });
  }
  return [];
};
