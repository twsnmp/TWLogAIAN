import * as echarts from 'echarts'
import { _,unwrapFunctionStore } from 'svelte-i18n';

const $_ = unwrapFunctionStore(_);


let chart;

export const getRanking = (logs, type) => {
  const m = new Map();
  logs.forEach((l) => {
    const k = l.KeyValue[type];
    const e = m.get(k)
    if (!e) {
      m.set(k, {
        Name: k,
        Total: 1,
      })
    } else {
      e.Total += 1
    }
  })
  const r = Array.from(m.values())
  r.sort((a, b) => b.Total - a.Total)
  return r
}

export const showRankingChart = (div, list, max,dark) => {
  const total = []
  const category = []
  for (let i = list.length > max ? max - 1 : list.length - 1; i >= 0; i--) {
    total.push(list[i].Total)
    category.push(list[i].Name)
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),dark ? "dark" : "");
  chart.setOption({
    title: {
      show: false,
    },
    toolbox: {},
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '20%',
      right: '10%',
      top: '5%',
      bottom: '10%',
      containLabel: false,
    },
    xAxis: {
      type: 'value',
      name: $_('Js.Count'),
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 8,
        margin: 2,
      },
    },
    yAxis: {
      type: 'category',
      data: category,
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        name: $_('Js.Count'),
        type: 'bar',
        data: total,
      },
    ],
  })
  chart.resize();
}

export const resizeRankingChart = () => {
  chart.resize();
}

export const getRankingChartImage = () => {
  if (chart) {
    return chart.getDataURL({ type:"png"});
  }
  return [];
}