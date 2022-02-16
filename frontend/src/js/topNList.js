import * as echarts from 'echarts'

let chart;

// 指定の値の出現回数
export const getTopList = (logs, type) => {
  const m = new Map();
  logs.forEach((l) => {
    const k = l.KeyValue[type];
    if (!k) {
      console.log(k, l);
      return
    }
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

export const showTopNChart = (div, list, max) => {
  const total = []
  const category = []
  for (let i = list.length > max ? max - 1 : list.length - 1; i >= 0; i--) {
    total.push(list[i].Total)
    category.push(list[i].Name)
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  chart.setOption({
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
      {
        offset: 0,
        color: '#4b5769',
      },
      {
        offset: 1,
        color: '#404a59',
      },
    ]),
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
    },
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
      name: '件数',
    },
    yAxis: {
      type: 'category',
      data: category,
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        name: '件数',
        type: 'bar',
        data: total,
      },
    ],
  })
  chart.resize();
}

export const resizeTopNChart = () => {
  chart.resize();
}