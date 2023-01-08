import * as echarts from "echarts";

let chart;


export const showLogHeatmap = (div, timeLine, dark) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), dark ? "dark" : null);
  if (timeLine.length < 1) {
    return;
  }
  const hours = [
    "0",
    "1",
    "2",
    "3",
    "4",
    "5",
    "6",
    "7",
    "8",
    "9",
    "10",
    "11",
    "12",
    "13",
    "14",
    "15",
    "16",
    "17",
    "18",
    "19",
    "20",
    "21",
    "22",
    "23",
  ];
  const option = {
    title: {
      show: false,
    },
    grid: {
      left: "10%",
      right: "5%",
      top: 30,
      buttom: 0,
    },
    toolbox: {
      feature: {
        dataZoom: {},
      },
    },
    tooltip: {
      trigger: "item",
      formatter(params) {
        return (
          params.name +
          " " +
          params.data[1] +
          " : " +
          params.data[2]
        );
      },
      axisPointer: {
        type: "shadow",
      },
    },
    xAxis: {
      type: "category",
      name: "Date",
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 10,
        margin: 2,
      },
      data: [],
    },
    yAxis: {
      type: "category",
      name: "Time",
      nameTextStyle: {
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 10,
        margin: 2,
      },
      data: hours,
    },
    visualMap: {
      min: Infinity,
      max: -Infinity,
      textStyle: {
        fontSize: 8,
      },
      calculable: true,
      realtime: false,
      inRange: {
        color: [
          "#313695",
          "#4575b4",
          "#74add1",
          "#abd9e9",
          "#e0f3f8",
          "#ffffbf",
          "#fee090",
          "#fdae61",
          "#f46d43",
          "#d73027",
          "#a50026",
        ],
      },
    },
    series: [
      {
        name: "Count",
        type: "heatmap",
        data: [],
        emphasis: {
          itemStyle: {
            borderWidth: 1,
          },
        },
        progressive: 1000,
        animation: false,
      },
    ],
  };
  const data = [];
  for (let t in timeLine) {
    data.push({
      time: new Date(t * 3600 * 1000),
      count: timeLine[t],
    });
  }
  data.sort((a,b) => a.time > b.time);
  let nD = 0;
  let x = -1;
  let day;
  data.forEach((e) => {
    day = echarts.time.format(e.time, "{yyyy}/{MM}/{dd}");
    if (nD !== e.time.getDate()) {
      option.xAxis.data.push(day);
      nD = e.time.getDate();
      x++;
    }
    const v = e.count;
    option.series[0].data.push([x, e.time.getHours(),v]);
    if (option.visualMap.min > v) {
      option.visualMap.min = v;
    }
    if (option.visualMap.max < v) {
      option.visualMap.max = v;
    }
  });
  chart.setOption(option);
  chart.resize();
  return data;
};

export const resizeLogHeatmap = () => {
  chart.resize();
};

export const getLogHeatmapImage = () => {
  if (chart) {
    return chart.getDataURL({ type: "png" });
  }
  return [];
};
