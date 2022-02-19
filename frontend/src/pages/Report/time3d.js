
import * as echarts from 'echarts'
import 'echarts-gl'
import * as ecStat from 'echarts-stat'
import { getFieldName } from "../../js/define";

let chart;

export const showTime3DChart = (div, logs, xField, zField, colorField,dark) => {
  const m = new Map()
  const colors = []
  logs.forEach((l) => {
    const t = new Date(l.Time / (1000 *1000))
    const x = l.KeyValue[xField] || "";
    const z = l.KeyValue[zField] ? l.KeyValue[zField]  * 1 : 0.0;
    const c = l.KeyValue[colorField] ? l.KeyValue[colorField] * 1 : 0.0;
    colors.push(c)
    const e = m.get(x)
    if (!e) {
      m.set(x, {
        Name: x,
        Total: 1,
        Time: [t],
        ZValue: [z],
        Color: [c],
      })
    } else {
      e.Total += 1
      e.Time.push(t)
      e.ZValue.push(z)
      e.Color.push(c)
    }
  })
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data = []
  l.forEach((e) => {
    for (let i = 0; i < e.Time.length && i < 15000; i++) {
      data.push([e.Name, e.Time[i], e.ZValue[i], e.Color[i]])
    }
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),dark ? "dark" : "")
  const options = {
    title: {
      show: false,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: true,
      min: ecStat.statistics.min(colors),
      max: ecStat.statistics.max(colors),
      dimension: 3,
      inRange: {
        color: dark ?
        [
          '#383899',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#e0f3f8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#ab0026',
        ]  : [
          '#313695',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#d0e3e8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#a50026',
        ],
      },
    },
    xAxis3D: {
      type: 'category',
      name: getFieldName(xField),
      data: cat,
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
        color: dark ? "#ccc" : "#222"
      },
      axisLabel: {
        fontSize: 8,
        margin: 2,
        color: dark ? "#ccc" : "#222",
      },
    },
    yAxis3D: {
      type: 'time',
      name: '日時',
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
        color: dark ? "#ccc" : "#222",
      },
      axisLabel: {
        fontSize: 8,
        color: dark ? "#ccc" : "#222",
        formatter(value, index) {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: getFieldName(zField),
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
        color: dark ? "#ccc" : "#222",
      },
      axisLabel: {
        fontSize: 8,
        margin: 2,
        color: dark ? "#ccc" : "#222",
      },
    },
    grid3D: {
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: '3D集計',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: [xField, 'Time', zField, colorField],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
  return data;
}

export const resizeTime3DChart = () => {
  chart.resize();
}
