import * as echarts from 'echarts';
import 'echarts-gl';
import * as ecStat from 'echarts-stat';

import { fft, util } from 'fft-js';
import { getFieldName } from '../../js/define';
import { _, unwrapFunctionStore } from 'svelte-i18n';

const $_ = unwrapFunctionStore(_);


const doFFT = (signal, sampleRate) => {
  if (!signal || signal.length < 2) {
    return [];
  }
  // fft-js requires power of 2 length
  let np2 = 2;
  while (np2 * 2 <= signal.length && np2 * 2 <= 512) {
    np2 *= 2;
  }
  // Zero-pad or slice to np2
  let s = signal.slice(0, np2);
  while (s.length < np2) {
    s.push(0);
  }
  try {
    const phasors = fft(s);
    const frequencies = util.fftFreq(phasors, sampleRate);
    const magnitudes = util.fftMag(phasors);
    const r = frequencies.map((f, ix) => {
      const p = f > 0.0 ? 1.0 / f : 0.0;
      return { period: p, frequency: f, magnitude: magnitudes[ix] || 0 };
    });
    return r;
  } catch (e) {
    console.error("doFFT error:", e);
    return [];
  }
};


export const getFFTMap = (logs, field) => {
  const m = new Map();
  m.set('Total', { Name: $_('Js.Total'), Count: 0, Data: [] });
  const validLogs = (logs || []).filter(l => l && l.Time > 946684800000000000);
  if (validLogs.length === 0) {
    return m;
  }
  let st = Infinity;
  let lt = 0;
  const isNone = field === '_None' || !field;

  validLogs.forEach((l) => {
    if (st > l.Time) st = l.Time;
    if (lt < l.Time) lt = l.Time;
    if (!isNone) {
      const k = l.KeyValue[field] || $_("Js.Unknown");
      if (!m.has(k)) {
        m.set(k, { Name: k, Count: 0, Data: [] });
      }
      l.Key = k;
    }
  });

  const dur = Math.max(1, (lt - st) / (1000 * 1000 * 1000));
  // Total bins capped at 512 to ensure FFT performance and avoid WebGL point overflow
  const TARGET_BINS = 512;
  const sampleSec = dur > TARGET_BINS ? Math.ceil(dur / TARGET_BINS) : 1;
  const totalBins = Math.min(TARGET_BINS, Math.max(2, Math.ceil(dur / sampleSec) + 1));

  // Initialize Data array with zeros for all categories
  m.forEach((e) => {
    e.Data = new Array(totalBins).fill(0);
  });

  validLogs.forEach((l) => {
    let binIdx = Math.floor((l.Time - st) / (1000 * 1000 * 1000 * sampleSec));
    if (binIdx < 0) binIdx = 0;
    if (binIdx >= totalBins) binIdx = totalBins - 1;

    m.get('Total').Data[binIdx]++;
    m.get('Total').Count++;

    if (!isNone && l.Key && m.has(l.Key)) {
      m.get(l.Key).Data[binIdx]++;
      m.get(l.Key).Count++;
    }
  });

  m.forEach((e) => {
    e.FFT = doFFT(e.Data, 1 / sampleSec);
  });

  return m;
};

let chart;

export const isWebGLWorking = () => {
  try {
    const canvas = document.createElement('canvas');
    const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
    if (!gl) return false;
    if (typeof gl.isContextLost === 'function' && gl.isContextLost()) return false;
    const s = gl.createShader(gl.VERTEX_SHADER);
    if (!s) return false;
    gl.deleteShader(s);
    return true;
  } catch (e) {
    return false;
  }
};

export const disposeFFTChart = () => {
  if (chart) {
    chart.dispose();
    chart = undefined;
  }
};

export const showFFTChart = (div, field, fftMap, key, fftType, dark) => {
  if (!fftMap || (key !== "" && !fftMap.get(key))) {
    return "";
  }
  const el = typeof div === "string" ? document.getElementById(div) : div;
  if (!el) {
    return "";
  }

  // Check if WebGL is working when 3D is requested
  if (key === "" && !isWebGLWorking()) {
    console.warn("WebGL context is unavailable, falling back to 2D chart.");
    if (chart) {
      chart.dispose();
      chart = undefined;
    }
    chart = echarts.init(el, dark ? "dark" : null);
    showFFT2DChart(fftMap, 'Total', fftType);
    return "fallback_2d";
  }

  if (chart) {
    chart.dispose();
    chart = undefined;
  }
  chart = echarts.init(el, dark ? "dark" : null);

  if (key === "") {
    showFFT3DChart(el, field, fftMap, fftType, dark);
    return "3d";
  }

  showFFT2DChart(fftMap, key, fftType);
  return "2d";
};

const showFFT2DChart = (fftMap, key, fftType) => {
  const entry = fftMap.get(key) || fftMap.get('Total');
  if (!entry || !entry.FFT) return;
  const fftData = entry.FFT;
  const freq = fftType === 'hz';
  const fft = [];
  fftData.forEach((e) => {
    if (freq) {
      fft.push([e.frequency, e.magnitude]);
    } else {
      if (e.period > 0) {
        fft.push([e.period, e.magnitude]);
      }
    }
  });

  const options = {
    title: { show: false },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: '10%',
      bottom: '10%',
    },
    xAxis: {
      type: 'value',
      name: freq ? $_('Js.Frequency') + "(Hz)" : $_('Js.Cycle') + "(Sec)",
      nameTextStyle: { fontSize: 10, margin: 2 },
      axisLabel: { fontSize: '8px' },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      name: 'Count',
      nameTextStyle: { fontSize: 10, margin: 2 },
      axisLabel: { fontSize: 8, margin: 2 },
    },
    series: [
      {
        name: $_('Js.Count'),
        type: 'bar',
        color: '#5470c6',
        emphasis: { focus: 'series' },
        showSymbol: false,
        data: fft,
      },
    ],
  };
  chart.setOption(options, true);
  chart.resize();
};

const showFFT3DChart = (el, field, fftMap, fftType, dark) => {
  const data = [];
  const freq = fftType === 'hz';
  const colors = [];
  const cat = [];

  let topEntries = Array.from(fftMap.entries())
    .filter(([k]) => k !== 'Total')
    .sort((a, b) => (b[1].Count || 0) - (a[1].Count || 0))
    .slice(0, 30);

  // If no specific category exists (e.g. "項目なし" or field === '_None'), show Total
  if (topEntries.length === 0 && fftMap.has('Total')) {
    topEntries = [['Total', fftMap.get('Total')]];
  }

  topEntries.forEach(([_, e]) => {
    cat.push(e.Name);
    if (!e.FFT) return;
    e.FFT.forEach((f) => {
      if (freq) {
        if (f.frequency <= 0.0) return;
        data.push([e.Name, f.frequency, f.magnitude, f.period]);
        colors.push(f.magnitude);
      } else {
        if (!f.period || f.period <= 0.0 || !isFinite(f.period)) return;
        data.push([e.Name, f.period, f.magnitude, f.frequency]);
        colors.push(f.magnitude);
      }
    });
  });

  // If data is empty (e.g. DC component only), fallback to 2D Total chart
  if (data.length === 0) {
    showFFT2DChart(fftMap, 'Total', fftType);
    return;
  }

  let minVal = colors.length > 0 ? ecStat.statistics.min(colors) : 0;
  let maxVal = colors.length > 0 ? ecStat.statistics.max(colors) : 1;
  if (minVal === maxVal) {
    maxVal = minVal + 1;
  }

  const fieldName = (field === '_None' || !field) ? $_('FFT.NoItem') : getFieldName(field);

  const options = {
    title: { show: false },
    tooltip: {},
    animationDurationUpdate: 500,
    visualMap: {
      show: true,
      min: minVal,
      max: maxVal,
      dimension: 2,
      inRange: {
        color: dark
          ? ['#383899', '#4575b4', '#74add1', '#abd9e9', '#e0f3f8', '#ffffbf', '#fee090', '#fdae61', '#f46d43', '#d73027', '#ab0026']
          : ['#313695', '#4575b4', '#74add1', '#abd9e9', '#d0e3e8', '#ffffbf', '#fee090', '#fdae61', '#f46d43', '#d73027', '#a50026'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: fieldName,
      data: cat,
      nameTextStyle: { fontSize: 10, margin: 2, color: dark ? "#ccc" : "#222" },
      axisLabel: { fontSize: 8, margin: 2, color: dark ? "#ccc" : "#222" },
    },
    yAxis3D: {
      type: freq ? 'value' : 'log',
      name: freq ? $_('Js.Frequency') + "(Hz)" : $_('Js.Cycle') + "(Sec)",
      nameTextStyle: { fontSize: 10, margin: 2, color: dark ? "#ccc" : "#222" },
      axisLabel: { fontSize: 8, color: dark ? "#ccc" : "#222" },
    },
    zAxis3D: {
      type: 'value',
      name: 'Count',
      nameTextStyle: { fontSize: 10, margin: 2, color: dark ? "#ccc" : "#222" },
      axisLabel: { fontSize: 8, margin: 2, color: dark ? "#ccc" : "#222" },
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
          fieldName,
          freq ? 'Frequency' : 'Cycle',
          'Count',
          freq ? $_('Js.Cycle') : $_('Js.Frequency'),
        ],
        data,
      },
    ],
  };

  try {
    chart.setOption(options, true);
    chart.resize();
  } catch (err) {
    console.warn("FFT 3D WebGL render failed, falling back to 2D chart:", err);
    if (chart) {
      chart.dispose();
      chart = undefined;
    }
    setTimeout(() => {
      if (el) {
        chart = echarts.init(el, dark ? "dark" : null);
        showFFT2DChart(fftMap, 'Total', fftType);
      }
    }, 10);
  }
};

export const resizeFFTChart = () => {
  if (chart) {
    chart.resize();
  }
};

export const getFFTChartImage = () => {
  if (chart) {
    return chart.getDataURL({ type: "png" });
  }
  return [];
};