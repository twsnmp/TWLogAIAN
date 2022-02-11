// GO言語側で定義した値をJavaScript側で使う
export let fieldTypes = {};

window.go.main.App.GetFieldTypes().then((r) =>{
  if (r) {
    fieldTypes = r;
  }
});
