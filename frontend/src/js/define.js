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
  return fieldTypes[f] ? fieldTypes[f].Name : f + "(unknown)";
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
    if (t.includes(getFieldType(f))) {
      ret.push(f);
    }
  });
  return ret;
}

export const getTableLimit = () => {
  if(window.innerHeight > 880) {
    return 10 + Math.floor((window.innerHeight - 880 ) /30);
  }
  return 10;
}