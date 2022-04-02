<script>
  import * as tf from "@tensorflow/tfjs";
  import { createEventDispatcher, onMount } from "svelte";
  
  export let logs = [];
  const dispatch = createEventDispatcher();
  onMount( async () => {
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
    for (let i = 0; i < 2; i++) {
      const h = await autoencoder.fit(x_train, x_train, 
        { epochs: 10, batchSize: 24,
          callbacks: tf.callbacks.earlyStopping({monitor: 'loss'}),
        });
      console.log("Loss after Epoch " + i + " : " + h.history.loss[0]);
    }
    for (let i = 0; i < logs.length; i++) {
      const x_eval = tf.tensor2d(logs[i].KeyValue["vector"], [1, dataLen]);
      const r = autoencoder.evaluate(x_eval, x_eval, {});
      logs[i].KeyValue["anomalyScore"] = r.dataSync()[0];
    }
    console.log("done");
    dispatch("done", {});
  });
</script>