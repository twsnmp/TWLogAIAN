import * as echarts from 'echarts';
import * as ecStat from 'echarts-stat';
import { _, unwrapFunctionStore } from 'svelte-i18n';
import { getFieldName } from '../../js/define';

const $_ = unwrapFunctionStore(_);

let chart;

// 遅延データの抽出と統計量計算
export const calculateDelays = (logs, mode = 'delta', field = '') => {
  if (!logs || logs.length === 0) {
    return {
      items: [],
      delays: [],
      stats: {
        count: 0,
        max: 0,
        mean: 0,
        median: 0,
        min: 0,
      },
    };
  }

  const items = [];
  const delays = [];

  logs.forEach((l) => {
    if (!l) return;
    let delaySec = 0;
    let isValid = false;

    if (mode === 'delta') {
      // タイムスタンプ逆転（delta < 0）またはDeltaフィールドの検証
      // processor.go では delta = curr.Time - lastTime と記録されており、
      // 逆転しているログは delta < 0 となる。twsla では delay = -delta で計算。
      const d = l.KeyValue ? l.KeyValue['delta'] : undefined;
      if (typeof d === 'number' && d < 0) {
        delaySec = -d;
        isValid = true;
      }
    } else if (mode === 'field' && field) {
      const v = l.KeyValue ? l.KeyValue[field] : undefined;
      if (typeof v === 'number') {
        delaySec = v;
        isValid = true;
      }
    }

    if (isValid && delaySec > 0) {
      delays.push(delaySec);
      const t = new Date(l.Time / (1000 * 1000));
      items.push({
        time: l.Time,
        date: t,
        delay: delaySec,
        log: l.All || '',
      });
    }
  });

  // 遅延の大きい順（ワースト順）にソート
  items.sort((a, b) => b.delay - a.delay);

  if (delays.length === 0) {
    return {
      items: [],
      delays: [],
      stats: {
        count: 0,
        max: 0,
        mean: 0,
        median: 0,
        min: 0,
      },
    };
  }

  const mean = ecStat.statistics.mean(delays);
  const median = ecStat.statistics.median(delays);
  const max = ecStat.statistics.max(delays);
  const min = ecStat.statistics.min(delays);

  return {
    items,
    delays,
    stats: {
      count: delays.length,
      max,
      mean,
      median,
      min,
    },
  };
};

// 遅延推移散布図を描画
export const showDelayChart = (div, items, dark, fieldName = '') => {
  if (chart) {
    chart.dispose();
    chart = null;
  }
  const el = typeof div === 'string' ? document.getElementById(div) : div;
  if (!el) return;

  chart = echarts.init(el, dark ? 'dark' : null);

  // 時系列表示用に時刻順にプロットデータを用意
  const plotData = [...items]
    .sort((a, b) => a.time - b.time)
    .map((item) => [
      item.date,
      Number(item.delay.toFixed(3)),
      item.log,
    ]);

  const yName = fieldName ? `${getFieldName(fieldName)} (s)` : `${$_('Delay.Delay')} (s)`;

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
        return `${dateStr}<br/><b>${yName}:</b> ${val[1]} s<br/><span style="font-size:10px;color:#888;">${val[2] ? val[2].substring(0, 100) : ''}</span>`;
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
      name: yName,
      min: 0,
      axisLabel: {
        fontSize: 10,
      },
    },
    series: [
      {
        name: yName,
        type: 'scatter',
        data: plotData,
        symbolSize: (val) => {
          // 遅延が大きいほどシンボルを大きくする
          return Math.min(15, Math.max(5, Math.log(val[1] + 1) * 3));
        },
        itemStyle: {
          color: (params) => {
            // 遅延が大きいものを警告色（オレンジ〜赤）に
            return params.value[1] > 3.0 ? '#d73a49' : '#0366d6';
          },
        },
      },
    ],
  };

  chart.setOption(option, true);
  chart.resize();
};

export const disposeDelayChart = () => {
  if (chart) {
    chart.dispose();
    chart = null;
  }
};

export const resizeDelayChart = () => {
  if (chart) {
    chart.resize();
  }
};

export const getDelayChartImage = () => {
  if (chart) {
    return chart.getDataURL();
  }
  return '';
};
