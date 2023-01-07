import {
  addMessages,
  locale,
  _,
  init, 
  getLocaleFromNavigator
} from 'svelte-i18n';

import en from './en.json';
import ja from './ja.json';


let currentLocale = "en";
const initI18n = () => {
  addMessages('en', en);
  addMessages('ja', ja);
  currentLocale = getLocaleFromNavigator() || 'en';
  init({
    fallbackLocale: 'en',
    initialLocale: currentLocale,
  });  
}

const  setLocale = (_locale) => {
  currentLocale = _locale;
  locale.set(_locale);
}

const getLocale =  () => {
  return currentLocale;
}

export {
  _,
  initI18n,
  setLocale,
  getLocale
};