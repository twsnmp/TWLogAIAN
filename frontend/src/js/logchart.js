import * as echarts from 'echarts'

let chart;

export const makeLogCountChart = (div) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = {
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
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '2%',
      right: '2%',
      top: 10,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: 'Time',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value, index) => {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      nameTextStyle: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: 'Count',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 8,
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
        type: 'bar',
        name: 'Count',
        color: '#1f78b4',
        large: true,
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const addChartData = (data, count, ctm, newCtm) => {
  let t = new Date(ctm * 60 * 1000)
  data.push({
    name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    value: [t, count],
  })
  ctm++
  for (; ctm < newCtm; ctm++) {
    t = new Date(ctm * 60 * 1000)
    data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, 0],
    })
  }
  return ctm
}

export const updateLogCountChart = (logs) => {
  const data = []
  let count = 0
  let ctm
  logs.forEach((e) => {
    const newCtm = Math.floor(e.Time / (1000 * 1000 * 1000 * 60))
    if (!ctm) {
      ctm = newCtm
    }
    if (ctm !== newCtm) {
      ctm = addChartData(data, count, ctm, newCtm)
      count = 0
    }
    count++
  })
  addChartData(data, count, ctm, ctm + 1)
  chart.setOption({
    series: [
      {
        data,
      },
    ],
  })
  chart.resize()
}
