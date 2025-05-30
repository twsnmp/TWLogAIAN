import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import { getFieldName } from "../../js/define";
import { _,unwrapFunctionStore } from 'svelte-i18n';

const $_ = unwrapFunctionStore(_);

let chart;

const makeHistogramChart = (div, field, dark) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div), dark ? "dark": null)
  const option = {
    title: {
      show: false,
    },
    toolbox: {},
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      formatter(params) {
        const p = params[0]
        return p.value[0] + ':' + p.value[1]
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      scale: true,
      min: 0,
      name: getFieldName(field),
      nameTextStyle: {
        fontSize: 10,
      },
      axisLabel: {
        fontSize: 8,
      },
    },
    yAxis: {
      name: $_('Js.Count'),
      min: 0,
    },
    series: [
      {
        color: "RGB(14,80,209)",
        type: 'bar',
        showSymbol: false,
        barWidth: '99.3%',
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}


export const showHistogramChart = (div, logs, field, dark) => {
  makeHistogramChart(div,field,dark)
  const data = []
  logs.forEach((l) => {
    data.push(l.KeyValue[field] || 0.0);
  })
  const bins = ecStat.histogram(data)
  chart.setOption({
    series: [
      {
        data: bins.data,
      },
    ],
  })
  chart.resize()
}

export const resizeHistogramChart = () => {
  chart.resize();
}

export const getHistogramChartImage = () => {
  if (chart) {
    return chart.getDataURL({ type:"png"});
  }
  return [];
}
