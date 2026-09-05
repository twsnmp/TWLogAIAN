import * as echarts from "echarts";

let emailChart;

export const showEmailSPFChart = (el, stats, dark) => {
  if (emailChart) {
    emailChart.dispose();
  }
  if (!el) return;
  emailChart = echarts.init(el, dark ? "dark" : undefined);

  const data = [
    { name: "Pass", value: stats.SPFPass || 0, itemStyle: { color: "#2ea44f" } },
    { name: "Fail", value: stats.SPFFail || 0, itemStyle: { color: "#cf222e" } },
    { name: "SoftFail", value: stats.SPFSoftFail || 0, itemStyle: { color: "#d29922" } },
    { name: "None", value: stats.SPFNone || 0, itemStyle: { color: "#8b949e" } },
    { name: "Other", value: stats.SPFOther || 0, itemStyle: { color: "#0969da" } },
  ].filter((d) => d.value > 0);

  const option = {
    backgroundColor: "transparent",
    title: {
      text: "SPF Verification Status",
      left: "center",
      textStyle: {
        fontSize: 14,
        color: dark ? "#c9d1d9" : "#24292f",
      },
    },
    tooltip: {
      trigger: "item",
      formatter: "{b}: {c} ({d}%)",
    },
    legend: {
      orient: "vertical",
      left: "left",
      textStyle: {
        color: dark ? "#c9d1d9" : "#24292f",
      },
    },
    series: [
      {
        name: "SPF",
        type: "pie",
        radius: ["40%", "70%"],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 6,
          borderColor: dark ? "#161b22" : "#fff",
          borderWidth: 2,
        },
        label: {
          show: true,
          formatter: "{b}: {c} ({d}%)",
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: "bold",
          },
        },
        data: data.length > 0 ? data : [{ name: "No Data", value: 0 }],
      },
    ],
  };

  emailChart.setOption(option, true);
};

export const showEmailCountBarChart = (el, list, dark, title) => {
  if (emailChart) {
    emailChart.dispose();
  }
  if (!el) return;
  emailChart = echarts.init(el, dark ? "dark" : undefined);

  const topItems = (list || []).slice(0, 15).reverse();
  const names = topItems.map((item) => item.Name);
  const counts = topItems.map((item) => item.Count);

  const option = {
    backgroundColor: "transparent",
    title: {
      text: title,
      left: "center",
      textStyle: {
        fontSize: 14,
        color: dark ? "#c9d1d9" : "#24292f",
      },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
    },
    grid: {
      left: "15%",
      right: "10%",
      bottom: "10%",
      top: "15%",
      containLabel: true,
    },
    xAxis: {
      type: "value",
      boundaryGap: [0, 0.01],
      splitLine: {
        lineStyle: {
          color: dark ? "#30363d" : "#e1e4e8",
        },
      },
    },
    yAxis: {
      type: "category",
      data: names,
      axisLabel: {
        formatter: (val) => (val.length > 25 ? val.substring(0, 22) + "..." : val),
      },
    },
    series: [
      {
        name: "Count",
        type: "bar",
        data: counts,
        itemStyle: {
          color: dark ? "#58a6ff" : "#0969da",
          borderRadius: [0, 4, 4, 0],
        },
      },
    ],
  };

  emailChart.setOption(option, true);
};

export const showEmailDelayChart = (el, items, dark) => {
  if (emailChart) {
    emailChart.dispose();
  }
  if (!el) return;
  emailChart = echarts.init(el, dark ? "dark" : undefined);

  const seriesData = (items || [])
    .filter((item) => item.Time > 0)
    .map((item) => [
      item.Time / (1000 * 1000), // ms
      item.Delay,
      item.From,
      item.To,
      item.Subject,
      item.Relay,
    ]);

  const option = {
    backgroundColor: "transparent",
    title: {
      text: "Email Delivery Delay Over Time",
      left: "center",
      textStyle: {
        fontSize: 14,
        color: dark ? "#c9d1d9" : "#24292f",
      },
    },
    tooltip: {
      trigger: "item",
      formatter: (params) => {
        const val = params.value;
        const d = new Date(val[0]);
        const timeStr = echarts.time.format(d, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
        return `
          <div style="font-size:12px; max-width: 400px; word-break: break-all;">
            <strong>${timeStr}</strong><br/>
            Delay: <b>${val[1]}s</b> (Relays: ${val[5]})<br/>
            From: ${val[2] || ""}<br/>
            To: ${val[3] || ""}<br/>
            Subject: ${val[4] || ""}
          </div>
        `;
      },
    },
    grid: {
      left: "8%",
      right: "6%",
      bottom: "18%",
      top: "15%",
    },
    xAxis: {
      type: "time",
      splitLine: {
        lineStyle: {
          color: dark ? "#30363d" : "#e1e4e8",
        },
      },
    },
    yAxis: {
      type: "value",
      name: "Delay (s)",
      splitLine: {
        lineStyle: {
          color: dark ? "#30363d" : "#e1e4e8",
        },
      },
    },
    dataZoom: [
      {
        type: "inside",
        xAxisIndex: 0,
      },
      {
        type: "slider",
        xAxisIndex: 0,
        bottom: 8,
        height: 20,
      },
    ],
    series: [
      {
        name: "Delay",
        type: "scatter",
        symbolSize: 8,
        itemStyle: {
          color: (param) => {
            const delay = param.data[1];
            if (delay > 300) return "#cf222e";
            if (delay > 60) return "#d29922";
            return "#2ea44f";
          },
          opacity: 0.8,
        },
        data: seriesData,
      },
    ],
  };

  emailChart.setOption(option, true);
};

export const disposeEmailChart = () => {
  if (emailChart) {
    emailChart.dispose();
    emailChart = undefined;
  }
};

export const resizeEmailChart = () => {
  if (emailChart) {
    emailChart.resize();
  }
};

export const getEmailChartImage = () => {
  if (emailChart) {
    return emailChart.getDataURL({
      type: "png",
      pixelRatio: 2,
      backgroundColor: "#fff",
    });
  }
  return "";
};
