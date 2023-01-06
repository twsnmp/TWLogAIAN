import {
  dictionary,
  locale,
  _,
  init, 
  getLocaleFromNavigator
} from 'svelte-i18n';

let currentLocale = "en";
const initI18n = () => {
  dictionary.set({
    en: {
      "Wellcome": {
        "Title": "Wellcom to TWLogAIAN",
        "Line1": "TWLogAIAN is an AI-assisted log analysis tool.",
        "Line2": "The manual is written in the following link",
        "Line3": "The source code is available on GitHUB.",
        "Line4": "Please let us know of any bugs or requests through ’Feedback’ or the link below.",
        "Thanks": "Thank you for using TWLogAIAN.",
        "Feedback": "Feedback",
        "Start": "Start",
      },
      "Feedback": {
        "Title": "Feedback",
        "SendMsg": "Sending ...",
        "SentMsg": "Feedback sent!",
        "SendError": "Failed to send feedback.",
        "Message": "Message",
        "CancelBtn": "Cancel",
        "SendBtn": "Send",
      },
    },
    ja: {
      "Wellcome": {
        "Title": "ようこそ TWLogAIANへ",
        "Line1": "TWLogAIANはAIアシスト付きログ分析ツールです。",
        "Line2": "マニュアルはnoteに書いています。",
        "Line3": "ソースコードはGitHUBにあります。",
        "Line4": "バグや要望は「フィードバック」か以下のリンクからお知らせください。",
        "Thanks": "TWLogAIANを利用いただきありがとうございます。",
        "Feedback": "フィードバック",
        "Start": "分析をはじめる",
      },
      "Feedback": {
        "Title": "フィードバック",
        "SendMsg": "フィードバックを送信中",
        "SentMsg": "フィードバックを送信しました",
        "SendError": "フィードバックの送信に失敗しました",
        "Message": "メッセージ",
        "CancelBtn": "キャンセル",
        "SendBtn": "送信",
      },
    },
  });
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