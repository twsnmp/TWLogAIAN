
import * as echarts from 'echarts'
import 'echarts-gl'
import * as ecStat from 'echarts-stat'

let chart;

const getLatLong = (loc) => {
  if (!loc) {
    return [139.548088, 35.856222]
  }
  const a = loc.split(',')
  if (a.length !== 2) {
    return [139.548088, 35.856222]
  }
  return [a[1], a[0]]
}

export const showGlobe = (div, logs, srcField, dstField,numField, dark) => {
  const m = new Map();
  logs.forEach((l) => {
    if (m.length > 20000) {
      return
    }
    const src = l.KeyValue[srcField];
    if (!src) {
      return;
    }
    const dst = l.KeyValue[dstField] || dstField || "Tokyo";
    const v = l.KeyValue[numField] || 0.0;
    const s = getLatLong(src);
    const d = getLatLong(dst);
    const srcCountry = l.KeyValue[srcField.replace("_latlong","_country")];
    const srcCity = l.KeyValue[srcField.replace("_latlong","_city")] || "";
    const dstCountry = l.KeyValue[dstField.replace("_latlong","_country")] || "";
    const dstCity = l.KeyValue[dstField.replace("_latlong","_cift")] || "";
    const srcName = src + (srcCountry ? "("+srcCountry + "/" + srcCity + ")" : "");
    const dstName = dst + (dstCountry ? "("+dstCountry + "/" + dstCity + ")" : "");
    const k = src + ":" + dst;
    let e = m.get(k)
    if (!e) {
      m.set(k, {
        src: srcName,
        dst: dstName,
        s: s,
        d: d,
        value: v,
        count: 1,
      })
    } else {
      e.value += v;
      e.count++;
    }
  });
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div), dark ? "dark" : null);
  const option = {
    backgroundColor: '#000',
    globe: {
      baseTexture: '/assets/images/world.topo.bathy.200401.jpg',
      heightTexture: '/assets/images/bathymetry_bw_composite_4k.jpg',
      shading: 'lambert',
      light: {
        ambient: {
          intensity: 0.4,
        },
        main: {
          intensity: 0.4,
        },
      },
      viewControl: {
        autoRotate: false,
      },
    },
    series: [
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: 'flow',
        data: [],
      },
    ],
  }
  const vs = [];
  const cs = [];
  m.forEach((e) => {
    vs.push(e.value)
    cs.push(e.count)
  })
  const th95 = ecStat.statistics.quantile(vs, 0.95);
  const th50 = ecStat.statistics.quantile(vs, 0.5);
  const maxc = ecStat.statistics.max(cs) || 1.0;
  m.forEach((e)=>{
    option.series[0].data.push(
      {
        coords: [ e.s, e.d ],
        value: e.value,
        name: e.src + "->" + e.dst,
        lineStyle: {
          color: e.value < th50 ? "#1f78b4" : e.value < th95 ? "#dfdf22" : "#e31a1c",
          width: (e.count * 8 / maxc) || 1,
          opacity: 0.5,
        }
      }
    );
  });
  chart.setOption(option)
  chart.resize()
  return m;
}

export const resizeGlobe = () => {
  chart.resize();
}

export const getGlobeImage = () => {
  if (chart) {
    return chart.getDataURL({ type:"png"});
  }
  return [];
}