<script>
  import * as echarts from "echarts";
  import * as tf from "@tensorflow/tfjs";
  import { createEventDispatcher, onMount } from "svelte";
  
  export let logs = [];
  export let dark = true;
  export let chartData = [];
  const dispatch = createEventDispatcher();
  let chart;
  const makeChart = () => {
    if (chart) {
      chart.dispose();
    }
    chart = echarts.init(document.getElementById("trainChart"), dark ? "dark" : null);
    const option = {
      title: {
        show: false,
      },
      tooltip: {
        trigger: "axis",
        axisPointer: {
          type: "shadow",
        },
      },
      grid: {
        left: "10%",
        right: "10%",
        top: 40,
        buttom: 0,
      },
      xAxis: {
        type: "time",
        name: "Time",
        axisLabel: {
          fontSize: "8px",
          formatter: (value, index) => {
            const date = new Date(value);
            return echarts.time.format(date, "{HH}:{mm}:{ss}");
          },
        },
        nameTextStyle: {
          fontSize: 10,
          margin: 2,
        },
        splitLine: {
          show: false,
        },
      },
      yAxis: [
        {
          type: "value",
          name: "Loss",
          nameTextStyle: {
            fontSize: 10,
            margin: 2,
          },
          axisLabel: {
            fontSize: 8,
            margin: 2,
          },
        },
      ],
      series: [
        {
          name: "Loss",
          type: "line",
          large: true,
          symbol: "none",
          color: "#c81100",
          data: chartData,
        },
      ],
    };
    chart.setOption(option);
    chart.resize();
  }
  const autoencoder = async () =>{
    tf.setBackend('cpu');
    console.log(tf.getBackend());
    if(logs.length < 10) {
      console.log("no logs",logs.length);
      dispatch("done", {});
      return;
    }
    if(!logs[0].KeyValue["vector"] || logs[0].KeyValue["vector"].length <1 ) {
      console.log("no vector",logs.length);
      dispatch("done", {});
      return;
    }
    console.log("start autoencoder");
    const vectors = [];
    logs.forEach((l) => {
      vectors.push(l.KeyValue["vector"]);
    });
    console.log("vector="+ vectors.length);

    const dataLen = vectors[0].length;
    const input = tf.input({ shape: [dataLen] });
    const encoded1 = tf.layers.dense({ units: Math.ceil(dataLen / 2), activation: 'relu' });
    const encoded2 = tf.layers.dense({ units: Math.ceil(dataLen / 4), activation: 'relu' });
    const encoded3 = tf.layers.dense({ units: Math.ceil(dataLen / 2), activation: 'relu' });
    const decoded = tf.layers.dense({ units: dataLen, activation: 'sigmoid' });
    const output = decoded.apply(encoded3.apply(encoded2.apply(encoded1.apply(input))));
    const autoencoder = tf.model({ inputs: input, outputs: output });
    autoencoder.compile({ optimizer: 'adam', loss: 'meanSquaredError' });
    const x_train = tf.tensor2d(vectors, [vectors.length, dataLen]);
    const h = await autoencoder.fit(x_train, x_train, { 
        epochs: 10, 
        batchSize: 32,
        callbacks:{
          onEpochEnd: (e,logs) => {
            console.log(e,logs);
            const t = new Date();
            const name = echarts.time.format(t, "{HH}:{mm}:{ss}");
            chartData.push({
              name,
              value: [t, logs.loss || 0.0],
            });
            if(chart) {
              chart.setOption({
                series:{
                  data: chartData,
                },
              })
            }
          },
        },
    });
    for (let i = 0; i < logs.length; i++) {
      const x_eval = tf.tensor2d(logs[i].KeyValue["vector"], [1, dataLen]);
      const r = autoencoder.evaluate(x_eval, x_eval, {});
      const d = r.dataSync();
      if(d && d.length >0) {
        logs[i].KeyValue["anomalyScore"] = d[0];
      }
    }
    console.log("autoencode done");
    dispatch("done", {});
  }

  onMount(() => {
    makeChart();
    if (chartData.length < 1) {
      autoencoder();
    }
  });

</script>

<div id="trainChart" />

<style>
  #trainChart {
    width: 95%;
    height: 200px;
    margin: 5px auto;
  }
</style>
