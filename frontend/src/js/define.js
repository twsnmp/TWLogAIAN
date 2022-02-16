// GO言語側で定義した値をJavaScript側で使う
let fieldTypes = {};

window.go.main.App.GetFieldTypes().then((r) =>{
  if (r) {
    fieldTypes = r;
  }
});

export const getFieldName = (f) => {
  return fieldTypes[f] ? fieldTypes[f].Name : f + "(未定義)";
};

export const getFieldType = (f) => {
  return fieldTypes[f] ? fieldTypes[f].Type : "string";
};
