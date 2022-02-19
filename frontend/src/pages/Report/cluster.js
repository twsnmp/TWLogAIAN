import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import { getFieldName } from "../../js/define";

let chart;

const makeClusterChart = (div, xField, yField,dark) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div), dark ? "dark": null)
  const option = {
    title: {
      show: false,
    },
    legend: {
      data: [],
    },
    toolbox: {},
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'value',
      name: getFieldName(xField),
      nameTextStyle: {
        fontSize: 10,
      },
      axisLabel: {
        fontSize: 8,
      },
    },
    yAxis: {
      type: 'value',
      name: getFieldName(yField),
      nameTextStyle: {
        fontSize: 10,
      },
      axisLabel: {
        fontSize: 8,
      },
    },
    series: [],
  }
  chart.setOption(option)
  chart.resize()
}


export const showClusterChart = (div, logs, xField, yField, cluster, dark) => {
  makeClusterChart(div,xField,yField,dark)
  const data = []
  logs.forEach((l) => {
    data.push([ l.KeyValue[xField] || 0.0, l.KeyValue[yField] || 0.0 ])
  })
  const result = ecStat.clustering.hierarchicalKMeans(data, {
    clusterCount: cluster,
    stepByStep: false,
    outputType: 'multiple',
    outputClusterIndexDimension: cluster,
  })
  if (!result) {
    chart.resize();
    return;
  }
  const centroids = result.centroids
  const ptsInCluster = result.pointsInCluster
  const series = []
  for (let i = 0; i < centroids.length; i++) {
    series.push({
      name: 'cluster' + (i + 1),
      type: 'scatter',
      data: ptsInCluster[i],
      markPoint: {
        symbolSize: 30,
        label: {
          show: true,
          position: 'top',
          formatter: (params) => {
            return (
              Math.round(params.data.coord[0] * 100) / 100 +
              ' / ' +
              Math.round(params.data.coord[1] * 100) / 100
            )
          },
          fontSize: 10,
        },
        data: [
          {
            coord: centroids[i],
          },
        ],
      },
    })
  }
  chart.setOption({
    legend: {
      data: [],
    },
    series,
  })
  chart.resize()
}

export const resizeClusterChart = () => {
  chart.resize();
}