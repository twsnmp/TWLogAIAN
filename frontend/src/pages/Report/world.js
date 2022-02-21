import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import WorldData from 'world-map-geojson'

let chart;

export const showWorldMap = (div, logs,geoField,numField,dark) => {
  if (chart) {
    chart.dispose()
  }
  const values = [];
  const counts = [];
  const m = new Map()
  logs.forEach((l) => {
    if (m.length > 10000) {
      return
    }
    const latlong = l.KeyValue[geoField  +"_latlong"];
    if (!latlong || !latlong.includes(",")) {
      return;
    }
    const e = m.get(latlong);
    const v = l.KeyValue[numField] || 0.0;
    if (!e) {
      m.set(latlong,{
        count: 1,
        value: v,
        country:l.KeyValue[geoField  +"_country"] || "",
        city: l.KeyValue[geoField  +"_city"] || "",
      });
    } else {
      e.count++;
      e.value = e.value < v ? v : e.value; 
    }
  })

  m.forEach((e)=>{
    values.push(e.value);
    counts.push(e.count);
  });

  chart = echarts.init(document.getElementById(div),dark ? "dark" : null);
  echarts.registerMap('world', WorldData)
  const option = {
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    visualMap: [
      {
        show: true,
        min: ecStat.statistics.min(values),
        max: ecStat.statistics.max(values),
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
      }, {
        show: false,
        min: ecStat.statistics.min(counts),
        max: ecStat.statistics.max(counts), 
        dimension: 3,
        inRange: {
          symbolSize:[3,16],
        },
      },
    ],
    geo: {
      map: 'world',
      silent: true,
      emphasis: {
        label: {
          show: false,
        },
      },
      itemStyle: {
        borderWidth: 0.2,
      },
      roam: true,
    },
    toolbox: {},
    tooltip: {
      trigger: 'item',
      formatter: "{b}",
    },
    series: [
      {
        type: 'scatter',
        coordinateSystem: 'geo',
        label: {
          show: false,
          formatter: "{b}",
        },
        emphasis: {
          label: {
            show: true,
          },
        },
        symbolSize: 6,
        data: [],
      },
    ],
  }
  m.forEach((e,k) => {
    const a = k.split(',')
    if (a.length < 2) {
      return;
    }
    let name = e.country ? e.country : k;
    if (e.city) {
      name += "(" + e.city + ")";
    }
    option.series[0].data.push({
      name: name + " v=" + e.value + " c=" + e.count,
      value: [a[1] * 1.0, a[0] * 1.0, e.value,e.count],
    })
  });
  chart.setOption(option);
  chart.resize();
  chart.on("dblclick",(p)=>{
    const url = 'https://www.google.com/maps/search/?api=1&query=' + p.value[1] + "," + p.value[0];
    window.runtime.BrowserOpenURL(url);
  });
  return m;
}

export const resizeWorldMap = () => {
  chart.resize();
}
