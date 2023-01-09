import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'

import {fft,util} from 'fft-js';
import { getFieldName } from '../../js/define';
import { _,unwrapFunctionStore } from 'svelte-i18n';

const $_ = unwrapFunctionStore(_);


const doFFT = (signal, sampleRate) => {
  const np2 = 1 << (31 - Math.clz32(signal.length))
  while (signal.length !== np2) {
    signal.shift()
  }
  const phasors = fft(signal)
  const frequencies = util.fftFreq(phasors, sampleRate)
  const magnitudes = util.fftMag(phasors)
  const r = frequencies.map((f, ix) => {
    const p = f > 0.0 ? 1.0 / f : 0.0
    return { period: p, frequency: f, magnitude: magnitudes[ix] }
  })
  return r
}


export const getFFTMap = (logs, field) => {
  const m = new Map()
  m.set('Total', { Name: $_('Js.Total'), Count: 0, Data: [] })
  let st = Infinity;
  let lt = 0;
  logs.forEach((l) => {
    const k = l.KeyValue[field] || $_("Js.Unknown");
    const e = m.get(k);
    if (!e) {
      m.set(k, { Name: k, Count: 0, Data: [] })
    }
    l.Key = k;
    if (st > l.Time) {
      st = l.Time;
    }
    if (lt < l.Time) {
      lt = l.Time;
    }
  })
  let sampleSec = 1;
  const dur = (lt - st) /(1000 * 1000 * 1000);
  if (dur > 3600 * 24 * 365) {
    sampleSec = 3600;
  } else if (dur > 3600 * 24 * 30) {
    sampleSec = 600;
  } else if (dur > 3600 * 24 * 7) {
    sampleSec = 120;
  } else if (dur > 3600 * 24) {
    sampleSec = 60;
  }
  let cts
  logs.forEach((l) => {
    if (!cts) {
      cts = Math.floor(l.Time / (1000 * 1000 * 1000 * sampleSec))
      m.get('Total').Count++
      m.get(l.Key).Count++
      return
    }
    const newCts = Math.floor(l.Time / (1000 * 1000 * 1000 * sampleSec))
    if (cts !== newCts) {
      m.forEach((e) => {
        e.Data.push(e.Count)
        e.Count = 0
      })
      cts++
      for (; cts < newCts; cts++) {
        m.forEach((e) => {
          e.Data.push(0)
        })
      }
    }
    m.get('Total').Count++
    m.get(l.Key).Count++
  })
  m.forEach((e) => {
    e.FFT = doFFT(e.Data, 1/sampleSec)
  })
  return m
}

let chart;

export const showFFTChart = (div,field, fftMap, key, fftType,dark) => {
  if (!fftMap || (key != "" && !fftMap.get(key)) ) {
    return
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),dark ? "dark" : null)
  if (key == "") {
    showFFT3DChart(field,fftMap,fftType,dark)
    return
  }
  const fftData = fftMap.get(key).FFT
  const freq = fftType === 'hz'
  const fft = []
  if (freq) {
    fftData.forEach((e) => {
      fft.push([e.frequency, e.magnitude])
    })
  } else {
    fftData.forEach((e) => {
      fft.push([e.period, e.magnitude])
    })
  }
  const options = {
    title: {
      show: false,
    },
    toolbox: {},
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: '10%',
      buttom: '10%',
    },
    xAxis: {
      type: 'value',
      name: freq ? $_('Js.Frequency') + "(Hz)" : $_('Js.Cycle') + "(Sec)",
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: '8px',
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: 'Count',
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
        color: '#5470c6',
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: fft,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const showFFT3DChart = (field,fftMap, fftType,dark) => {
  const data = []
  const freq = fftType === 'hz'
  const colors = []
  const cat = []
  fftMap.forEach((e) => {
    if (e.Name === 'Total') {
      return
    }
    cat.push(e.Name)
    if (freq) {
      e.FFT.forEach((f) => {
        if (f.frequency === 0.0) {
          return
        }
        data.push([e.Name, f.frequency, f.magnitude, f.period])
        colors.push(f.magnitude)
      })
    } else {
      e.FFT.forEach((f) => {
        if (f.period === 0.0) {
          return
        }
        data.push([e.Name, f.period, f.magnitude, f.frequency])
        colors.push(f.magnitude)
      })
    }
  })
  const options = {
    title: {
      show: false,
    },
    toolbox: {},
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: true,
      min: ecStat.statistics.min(colors),
      max: ecStat.statistics.max(colors),
      dimension: 2,
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
      name: getFieldName(field),
      data: cat,
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
    yAxis3D: {
      type: freq ? 'value' : 'log',
      name: freq ? $_('Js.Frequency') + "(Hz)" : $_('Js.Cycle') + "(Sec)",
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
        color: dark ? "#ccc" : "#222",
      },
      axisLabel: {
        fontSize: 8,
        color: dark ? "#ccc" : "#222",
      },
    },
    zAxis3D: {
      type: 'value',
      name: 'Count',
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
        name: 'Syslog FFT',
        type: 'scatter3D',
        symbolSize: 3,
        dimensions: [
          getFieldName(field),
          freq ? 'Frequency' : 'Cycle',
          'Count',
          freq ? $_('Js.Cycle') : $_('Js.Frequency'),
        ],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

export const resizeFFTChart = () => {
  chart.resize();
}

export const getFFTChartImage = () => {
  if (chart) {
    return chart.getDataURL({ type:"png"});
  }
  return [];
}