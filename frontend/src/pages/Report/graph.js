import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import { getFieldName } from "../../js/define";

let chart;

export const showGraph = (div, logs, srcField,dstField,numField,type,dark) => {
  const nodeMap = new Map()
  const edgeMap = new Map()
  logs.forEach((l) => {
    const src = l.KeyValue[srcField];
    const dst = l.KeyValue[dstField];
    if (!src || !dst) {
      return;
    }
    let ek = src + '|' + dst;
    const v = l.KeyValue[numField] || 0.0;
    let e = edgeMap.get(ek)
    if (!e) {
      edgeMap.set(ek, {
        source: src,
        target: dst,
        value: v,
        count: 1,
      })
    } else {
      e.value += v;
      e.count++;
    }
    let n = nodeMap.get(src)
    if (!n) {
      nodeMap.set(src, {
        name: src,
        count: 1,
        category: 0,
        draggable: true,
      })
    } else {
      n.count++;
    }
    n = nodeMap.get(dst)
    if (!n) {
      nodeMap.set(dst, {
        name: dst,
        count: 1,
        category: 1,
        draggable: true,
      })
    } else {
      n.count++;
    }
  })
  const nodes = Array.from(nodeMap.values());
  const edges = Array.from(edgeMap.values());
  const nvs = [];
  const evs = [];
  nodes.forEach((e) => {
    nvs.push(e.value)
  })
  edges.forEach((e) => {
    evs.push(e.value)
  })
  const n95 = ecStat.statistics.quantile(nvs, 0.95)
  const n50 = ecStat.statistics.quantile(nvs, 0.5)
  const e95 = ecStat.statistics.quantile(evs, 0.95)
  let mul = 1.0
  if (type === 'gl') {
    mul = 1.5
  }
  nodes.forEach((e) => {
    e.label = { show: e.value > n95 }
    e.symbolSize = e.value > n95 ? 5 : e.value > n50 ? 4 : 2
    e.symbolSize *= mul
  })
  edges.forEach((e) => {
    e.lineStyle = {
      width: e.value > e95 ? 2 : 1,
    }
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),dark ? "dark" : null);
  const categories = [
    { name: getFieldName(srcField) },
    { name: getFieldName(dstField) },
  ];
  const options = {
    title: {
      show: false,
    },
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    toolbox: {},
    tooltip: {},
    legend: [
      {
        orient: 'vertical',
        top: 50,
        right: 20,
        textStyle: {
          fontSize: 10,
        },
        data: [getFieldName(srcField),getFieldName(dstField)],
      },
    ],
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [],
  }
  if (type === 'circular') {
    options.series = [
      {
        name: getFieldName(srcField)+":"+getFieldName(dstField),
        type: 'graph',
        layout: 'circular',
        circular: {
          rotateLabel: true,
        },
        data: nodes,
        links: edges,
        categories,
        roam: true,
        label: {
          position: 'right',
          formatter: '{b}',
          fontSize: 8,
          fontStyle: 'normal',
        },
        lineStyle: {
          color: 'source',
          curveness: 0.3,
        },
      },
    ]
  } else if (type === 'gl') {
    options.series = [
      {
        name: getFieldName(srcField)+":"+getFieldName(dstField),
        type: 'graphGL',
        nodes,
        edges,
        modularity: {
          resolution: 2,
          sort: true,
        },
        lineStyle: {
          color: 'source',
          opacity: 0.5,
        },
        itemStyle: {
          opacity: 1,
        },
        focusNodeAdjacency: false,
        focusNodeAdjacencyOn: 'click',
        emphasis: {
          label: {
            show: false,
          },
          lineStyle: {
            opacity: 0.5,
            width: 4,
          },
        },
        forceAtlas2: {
          steps: 5,
          stopThreshold: 20,
          jitterTolerence: 10,
          edgeWeight: [0.2, 1],
          gravity: 5,
          edgeWeightInfluence: 0,
        },
        categories,
        label: {
          position: 'right',
          formatter: '{b}',
          fontSize: 8,
          fontStyle: 'normal',
        },
      },
    ]
  } else {
    options.series = [
      {
        name: getFieldName(srcField)+":"+getFieldName(dstField),
        type: 'graph',
        layout: 'force',
        data: nodes,
        links: edges,
        categories,
        roam: true,
        label: {
          position: 'right',
          formatter: '{b}',
          fontSize: 8,
          fontStyle: 'normal',
        },
        lineStyle: {
          color: 'source',
          curveness: 0,
        },
      },
    ]
  }
  chart.setOption(options);
  chart.resize();
  return edgeMap;
}

export const resizeGraph = () => {
  chart.resize();
}

export const getGraphImage = () => {
  if (chart) {
    return chart.getDataURL({ type:"png"});
  }
  return [];
}