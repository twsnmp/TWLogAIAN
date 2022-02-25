// GO言語側で定義した値をJavaScript側で使う
let fieldTypes = {};

export const loadFieldTypes = () => {
  window.go.main.App.GetFieldTypes().then((r) =>{
    if (r) {
      fieldTypes = r;
    }
  });
}

export const getFieldName = (f) => {
  if (f=="") {
    return "";
  }
  return fieldTypes[f] ? fieldTypes[f].Name : f + "(未定義)";
};

export const getFieldType = (f) => {
  return fieldTypes[f] ? fieldTypes[f].Type : "string";
};

export const getFieldUnit = (f) => {
  return fieldTypes[f] ? fieldTypes[f].Unit : "";
};

export const isFieldValid = (f) => {
  return fieldTypes[f] ? true : false;
};

export const getFields = (fields,t) => {
  const ret = [];
  fields.forEach((f) => {
    if (getFieldType(f) == t) {
      ret.push(f);
    }
  });
  return ret;
}

export const getTableLimit = () => {
  if(window.innerHeight > 880) {
    return 10 + Math.floor((window.innerHeight - 880 ) /25);
  }
  return 10;
}