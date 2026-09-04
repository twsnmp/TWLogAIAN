import * as echarts from 'echarts';
import * as ecStat from 'echarts-stat';
import { _, unwrapFunctionStore } from 'svelte-i18n';

const $_ = unwrapFunctionStore(_);

let chart;

// ログ配列からインターバル（秒）および統計量を計算
export const calculateIntervals = (logs) => {
  if (!logs || logs.length < 2) {
    return {
      items: [],
      deltas: [],
      stats: {
        count: 0,
        mean: 0,
        median: 0,
        mode: 0,
        min: 0,
        max: 0,
        stddev: 0,
      },
    };
  }

  // タイムスタンプ順にソート
  const sorted = [...logs]
    .filter((l) => l && l.Time > 946684800000000000)
    .sort((a, b) => a.Time - b.Time);

  if (sorted.length < 2) {
    return {
      items: [],
      deltas: [],
      stats: {
        count: 0,
        mean: 0,
        median: 0,
        mode: 0,
        min: 0,
        max: 0,
        stddev: 0,
      },
    };
  }

  const items = [];
  const deltas = [];

  for (let i = 1; i < sorted.length; i++) {
    const prev = sorted[i - 1];
    const curr = sorted[i];
    // ナノ秒から秒に変換
    const deltaSec = (curr.Time - prev.Time) / (1000 * 1000 * 1000);
    deltas.push(deltaSec);

    const t = new Date(curr.Time / (1000 * 1000));
    items.push({
      time: curr.Time,
      date: t,
      delta: deltaSec,
      log: curr.All || '',
    });
  }

  // 統計量の計算
  const mean = ecStat.statistics.mean(deltas);
  const median = ecStat.statistics.median(deltas);
  const min = ecStat.statistics.min(deltas);
  const max = ecStat.statistics.max(deltas);
  const variance = ecStat.statistics.sampleVariance(deltas);
  const stddev = Math.sqrt(variance);

  // 最頻値（簡易ヒストグラムベースのビン、または頻出値）
  let mode = median;
  try {
    const bins = ecStat.histogram(deltas);
    let maxCount = -1;
    let modeBin = null;
    bins.data.forEach((b) => {
      if (b[1] > maxCount) {
        maxCount = b[1];
        modeBin = b[0];
      }
    });
    if (modeBin !== null) {
      mode = modeBin;
    }
  } catch (e) {
    mode = median;
  }

  return {
    items,
    deltas,
    stats: {
      count: deltas.length,
      mean,
      median,
      mode,
      min,
      max,
      stddev,
    },
  };
};

// インターバル推移散布図を描画
export const showIntervalChart = (div, items, dark, mode = 'scatter') => {
  if (chart) {
    chart.dispose();
    chart = null;
  }
  const el = typeof div === 'string' ? document.getElementById(div) : div;
  if (!el) return;

  chart = echarts.init(el, dark ? 'dark' : null);

  const data = items.map((item) => [
    item.date,
    Number(item.delta.toFixed(3)),
    item.log,
  ]);

  const option = {
    title: {
      show: false,
    },
    toolbox: {
      feature: {
        dataZoom: {},
      },
    },
    dataZoom: [
      {
        type: 'slider',
        xAxisIndex: 0,
        bottom: 6,
        height: 20,
      },
      {
        type: 'inside',
        xAxisIndex: 0,
      },
    ],
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        const val = params.value;
        const dateStr = echarts.time.format(val[0], '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}');
        return `${dateStr}<br/><b>${$_('Interval.Interval')}:</b> ${val[1]} s<br/><span style="font-size:10px;color:#888;">${val[2] ? val[2].substring(0, 100) : ''}</span>`;
      },
    },
    grid: {
      left: '6%',
      right: '6%',
      top: 25,
      bottom: 55,
      containLabel: true,
    },
    xAxis: {
      type: 'time',
      name: $_('Js.Time'),
      nameLocation: 'end',
      nameGap: 10,
      nameTextStyle: {
        fontSize: 10,
      },
      axisLabel: {
        fontSize: 9,
        margin: 10,
        formatter: (value) => {
          return echarts.time.format(new Date(value), '{MM}/{dd} {HH}:{mm}');
        },
      },
    },
    yAxis: {
      type: 'value',
      name: `${$_('Interval.Interval')} (s)`,
      min: 0,
      axisLabel: {
        fontSize: 10,
      },
    },
    series: [
      {
        name: $_('Interval.Interval'),
        type: mode,
        data,
        symbolSize: 5,
        itemStyle: {
          color: 'RGB(14, 130, 209)',
        },
      },
    ],
  };

  chart.setOption(option, true);
  chart.resize();
};

// ヒストグラム表示
export const showIntervalHistogramChart = (div, deltas, dark) => {
  if (chart) {
    chart.dispose();
    chart = null;
  }
  const el = typeof div === 'string' ? document.getElementById(div) : div;
  if (!el || !deltas || deltas.length === 0) return;

  chart = echarts.init(el, dark ? 'dark' : null);
  const bins = ecStat.histogram(deltas);

  const option = {
    title: {
      show: false,
    },
    toolbox: {
      feature: {
        dataZoom: {},
      },
    },
    dataZoom: [
      {
        type: 'slider',
        xAxisIndex: 0,
        bottom: 6,
        height: 20,
      },
    ],
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        const p = params[0];
        return `${$_('Interval.Interval')}: ~${Number(p.value[0]).toFixed(2)}s<br/>${$_('Js.Count')}: ${p.value[1]}`;
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '6%',
      right: '6%',
      top: 25,
      bottom: 55,
      containLabel: true,
    },
    xAxis: {
      scale: true,
      min: 0,
      name: `${$_('Interval.Interval')} (s)`,
      nameTextStyle: {
        fontSize: 10,
      },
      axisLabel: {
        fontSize: 9,
      },
    },
    yAxis: {
      name: $_('Js.Count'),
      min: 0,
    },
    series: [
      {
        color: 'RGB(14,80,209)',
        type: 'bar',
        showSymbol: false,
        barWidth: '99.3%',
        data: bins.data,
      },
    ],
  };

  chart.setOption(option, true);
  chart.resize();
};

export const disposeIntervalChart = () => {
  if (chart) {
    chart.dispose();
    chart = null;
  }
};

export const resizeIntervalChart = () => {
  if (chart) {
    chart.resize();
  }
};

export const getIntervalChartImage = () => {
  if (chart) {
    return chart.getDataURL();
  }
  return '';
};
