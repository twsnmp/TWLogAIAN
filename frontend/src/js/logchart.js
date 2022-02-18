import * as echarts from 'echarts'

let chart;

const baseOption = {
  title: {
    show: false,
  },
  dataZoom: [{}],
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'shadow',
    },
  },
  grid: {
    left: '5%',
    right: '5%',
    top: 10,
    buttom: 0,
  },
  xAxis: {
    type: 'time',
    name: 'Time',
    axisLabel: {
      fontSize: '8px',
      formatter: (value, index) => {
        const date = new Date(value)
        return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
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
    type: 'value',
    name: 'Count',
    nameTextStyle: {
      fontSize: 8,
      margin: 2,
    },
    axisLabel: {
      fontSize: 8,
      margin: 2,
    },
  },
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

const showLogCountChart = (div,logs,dark) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  chart.setOption(baseOption);
  const data = []
  let count = 0
  let ctm
  logs.forEach((l) => {
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60))
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
        type: 'bar',
        name: 'Count',
        color: '#1f78b4',
        large: true,
        data: data,
      },
    ],
  });
  chart.resize()
}

const addMultiChartData = (data, count, ctm, newCtm) => {
  let t = new Date(ctm * 60 * 1000)
  for (const k in count) {
    data[k].push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, count[k]],
    })
  }
  ctm++
  for (; ctm < newCtm; ctm++) {
    t = new Date(ctm * 60 * 1000)
    for (const k in count) {
      data[k].push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        value: [t, 0],
      })
    }
  }
  return ctm
}

const getSyslogLevel = (pri) => {
  if(pri==""){
    return "other";
  }
  // TODO:<pri>の場合、文字列の場合も対応する
  if( pri < 3){
    return "high";
  } else if( pri < 4){
    return "low";
  } else if (pri == 4){
    return "warn";
  }
  return "other";
}

const showAccessLogChart = (div,logs,dark) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),dark ? "dark" : "");
  chart.setOption(baseOption);
  const data = {
    normal: [],
    warn: [],
    error: [],
  }
  const count = {
    normal: 0,
    warn: 0,
    error: 0,
  }
  let ctm
  logs.forEach((l) => {
    const code = l.KeyValue.response || 200;
    const lvl = code < 300 ? "normal" : code < 400 ? "warn" : "error";
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60))
    if (!ctm) {
      ctm = newCtm
    }
    if (ctm !== newCtm) {
      ctm = addMultiChartData(data, count, ctm, newCtm)
      for (const k in count) {
        count[k] = 0
      }
    }
    count[lvl]++
  })
  addMultiChartData(data, count, ctm, ctm + 1)
  chart.setOption(    {
    grid: {
      left: '2%',
      right: '5%',
      top: 50,
      buttom: 0,
    },
    series: [
      {
        name: '正常',
        type: 'bar',
        color: '#1f78b4',
        stack: 'count',
        large: true,
        data: data.normal,
      },
      {
        name: '注意',
        type: 'bar',
        color: '#dfdf22',
        stack: 'count',
        large: true,
        data: data.warn,
      },
      {
        name: 'エラー',
        type: 'bar',
        color: '#e31a1c',
        stack: 'count',
        large: true,
        data: data.error,
      },
    ],
    legend: {
      textStyle: {
        fontSize: 10,
      },
      data: ['正常', '注意', 'エラー'],
    },
  })
  chart.resize()
}

const showSyslogChart = (div,logs,dark) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),dark ? "dark" : "")
  chart.setOption(baseOption);
  const data = {
    high: [],
    low: [],
    warn: [],
    other: [],
  }
  const count = {
    high: 0,
    low: 0,
    warn: 0,
    other: 0,
  }
  let ctm
  logs.forEach((l) => {
    const lvl = getSyslogLevel(l.KeyValue.priority || "");
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60))
    if (!ctm) {
      ctm = newCtm
    }
    if (ctm !== newCtm) {
      ctm = addMultiChartData(data, count, ctm, newCtm)
      for (const k in count) {
        count[k] = 0
      }
    }
    count[lvl]++
  })
  addMultiChartData(data, count, ctm, ctm + 1)
  chart.setOption({
    grid: {
      left: '2%',
      right: '5%',
      top: 50,
      buttom: 0,
    },
    series: [
      {
        name: '重度',
        type: 'bar',
        color: '#e31a1c',
        stack: 'count',
        large: true,
        data: data.high,
      },
      {
        name: '軽度',
        type: 'bar',
        color: '#fb9a99',
        stack: 'count',
        large: true,
        data: data.low,
      },
      {
        name: '注意',
        type: 'bar',
        color: '#dfdf22',
        stack: 'count',
        large: true,
        data: data.warn,
      },
      {
        name: 'その他',
        type: 'bar',
        color: '#1f78b4',
        stack: 'count',
        large: true,
        data: data.other,
      },
    ],
    legend: {
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['重度', '軽度', '注意', 'その他'],
    },
  })
  chart.resize()
}

export const showLogChart = (div,r,dark) => {
  switch (r.View) {
  case "access":
    showAccessLogChart(div,r.Logs,dark);
    return
  case "syslog":
    showSyslogChart(div,r.Logs,dark);
    return
  }
  showLogCountChart(div,r.Logs,dark);
}

export const resizeLogChart = () => {
  if(chart) {
    chart.resize()
  }
}