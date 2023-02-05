import * as echarts from "echarts";
import { html, h } from "gridjs";
import { getFieldName } from "../../js/define";
import { _,unwrapFunctionStore } from 'svelte-i18n';

const $_ = unwrapFunctionStore(_);
 
const formatCode = (code) => {
  if (code < 300) {
    return html(`<div class="color-fg-default">${code}</div>`);
  } else if (code < 400) {
    return html(`<div class="color-fg-attention">${code}</div>`);
  }
  return html(`<div class="color-fg-danger">${code}</div>`);
};

const formatLevel = (level) => {
  switch (level) {
    case "error":
      return html(`<div class="color-fg-danger">${$_('Js.Error')}</div>`);
    case "warn":
      return html(`<div class="color-fg-attention">${$_('Js.Warnning')}</div>`);
  }
  return html(`<div class="color-fg-default">${$_('Js.Normal')}</div>`);
};

const columnsTimeOnly = () => {
  return  [
  {
    id: "level",
    name: $_('Js.Level'),
    width: "6%",
    formatter: (cell) => cell ? formatLevel(cell) : "",
  },
  {
    id: "_timestamp",
    name: $_("Js.Time"),
    width: "15%",
    formatter: (cell) =>
    cell ?
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}"
      ) : "",
    convert: true,
  },
  {
    id: "score",
    name: $_("Js.Score"),
    width: "5%",
    formatter: (cell) => cell.toFixed(2),
  },
  {
    id: "all",
    name: $_('Js.Log'),
    width: "60%",
  },
  {
    id: "copy",
    name: "C",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Copy">
      <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"></path><path fill-rule="evenodd" d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"></path></svg>
    </button>`),
  },
  {
    id: "memo",
    name: "M",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Memo">
    <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>
    </button>`),
  },
  {
    id: "extractor",
    name: "E",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Extract">
    <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M7.429 1.525a6.593 6.593 0 011.142 0c.036.003.108.036.137.146l.289 1.105c.147.56.55.967.997 1.189.174.086.341.183.501.29.417.278.97.423 1.53.27l1.102-.303c.11-.03.175.016.195.046.219.31.41.641.573.989.014.031.022.11-.059.19l-.815.806c-.411.406-.562.957-.53 1.456a4.588 4.588 0 010 .582c-.032.499.119 1.05.53 1.456l.815.806c.08.08.073.159.059.19a6.494 6.494 0 01-.573.99c-.02.029-.086.074-.195.045l-1.103-.303c-.559-.153-1.112-.008-1.529.27-.16.107-.327.204-.5.29-.449.222-.851.628-.998 1.189l-.289 1.105c-.029.11-.101.143-.137.146a6.613 6.613 0 01-1.142 0c-.036-.003-.108-.037-.137-.146l-.289-1.105c-.147-.56-.55-.967-.997-1.189a4.502 4.502 0 01-.501-.29c-.417-.278-.97-.423-1.53-.27l-1.102.303c-.11.03-.175-.016-.195-.046a6.492 6.492 0 01-.573-.989c-.014-.031-.022-.11.059-.19l.815-.806c.411-.406.562-.957.53-1.456a4.587 4.587 0 010-.582c.032-.499-.119-1.05-.53-1.456l-.815-.806c-.08-.08-.073-.159-.059-.19a6.44 6.44 0 01.573-.99c.02-.029.086-.075.195-.045l1.103.303c.559.153 1.112.008 1.529-.27.16-.107.327-.204.5-.29.449-.222.851-.628.998-1.189l.289-1.105c.029-.11.101-.143.137-.146zM8 0c-.236 0-.47.01-.701.03-.743.065-1.29.615-1.458 1.261l-.29 1.106c-.017.066-.078.158-.211.224a5.994 5.994 0 00-.668.386c-.123.082-.233.09-.3.071L3.27 2.776c-.644-.177-1.392.02-1.82.63a7.977 7.977 0 00-.704 1.217c-.315.675-.111 1.422.363 1.891l.815.806c.05.048.098.147.088.294a6.084 6.084 0 000 .772c.01.147-.038.246-.088.294l-.815.806c-.474.469-.678 1.216-.363 1.891.2.428.436.835.704 1.218.428.609 1.176.806 1.82.63l1.103-.303c.066-.019.176-.011.299.071.213.143.436.272.668.386.133.066.194.158.212.224l.289 1.106c.169.646.715 1.196 1.458 1.26a8.094 8.094 0 001.402 0c.743-.064 1.29-.614 1.458-1.26l.29-1.106c.017-.066.078-.158.211-.224a5.98 5.98 0 00.668-.386c.123-.082.233-.09.3-.071l1.102.302c.644.177 1.392-.02 1.82-.63.268-.382.505-.789.704-1.217.315-.675.111-1.422-.364-1.891l-.814-.806c-.05-.048-.098-.147-.088-.294a6.1 6.1 0 000-.772c-.01-.147.039-.246.088-.294l.814-.806c.475-.469.679-1.216.364-1.891a7.992 7.992 0 00-.704-1.218c-.428-.609-1.176-.806-1.82-.63l-1.103.303c-.066.019-.176.011-.299-.071a5.991 5.991 0 00-.668-.386c-.133-.066-.194-.158-.212-.224L10.16 1.29C9.99.645 9.444.095 8.701.031A8.094 8.094 0 008 0zm1.5 8a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0zM11 8a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
    </button>`),
  },

  ];
}

const getTimeOnlyLogData = (r, filter, scoreField) => {
  if (!scoreField) {
    scoreField = "score";
  }
  const d = [];
  r.Logs.forEach((l) => {
    if (filter && filter.st) {
      if (l.Time < filter.st || l.Time > filter.et) {
        return;
      }
    }
    const score = l.KeyValue[scoreField] || 0.0;
    d.push([getLogLevel(l), l.Time, score, l.All,"copy","memo",l.All]);
  });
  return d;
};

const columnsSyslog = () => {
  return  [
  {
    id: "level",
    name: $_("Js.Level"),
    width: "6%",
    formatter: (cell) => cell ? formatLevel(cell) : "",
  },
  {
    id: "_timestamp",
    name: $_("Js.Time"),
    width: "15%",
    formatter: (cell) => cell ?
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}"
      ) : "",
    convert: true,
  },
  {
    id: "logsrc",
    name: $_("Js.SRC"),
    width: "14%",
  },
  {
    id: "tag",
    name: $_("Js.Tag"),
    width: "18%",
  },
  {
    id: "message",
    name: $_("Js.Message"),
    width: "38%",
  },
  {
    id: "copy",
    name: "C",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Copy">
      <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"></path><path fill-rule="evenodd" d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"></path></svg>
    </button>`),
  },
  {
    id: "memo",
    name: "M",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Memo">
    <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>
    </button>`),
  },
  {
    id: "extractor",
    name: "E",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Extract">
    <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M7.429 1.525a6.593 6.593 0 011.142 0c.036.003.108.036.137.146l.289 1.105c.147.56.55.967.997 1.189.174.086.341.183.501.29.417.278.97.423 1.53.27l1.102-.303c.11-.03.175.016.195.046.219.31.41.641.573.989.014.031.022.11-.059.19l-.815.806c-.411.406-.562.957-.53 1.456a4.588 4.588 0 010 .582c-.032.499.119 1.05.53 1.456l.815.806c.08.08.073.159.059.19a6.494 6.494 0 01-.573.99c-.02.029-.086.074-.195.045l-1.103-.303c-.559-.153-1.112-.008-1.529.27-.16.107-.327.204-.5.29-.449.222-.851.628-.998 1.189l-.289 1.105c-.029.11-.101.143-.137.146a6.613 6.613 0 01-1.142 0c-.036-.003-.108-.037-.137-.146l-.289-1.105c-.147-.56-.55-.967-.997-1.189a4.502 4.502 0 01-.501-.29c-.417-.278-.97-.423-1.53-.27l-1.102.303c-.11.03-.175-.016-.195-.046a6.492 6.492 0 01-.573-.989c-.014-.031-.022-.11.059-.19l.815-.806c.411-.406.562-.957.53-1.456a4.587 4.587 0 010-.582c.032-.499-.119-1.05-.53-1.456l-.815-.806c-.08-.08-.073-.159-.059-.19a6.44 6.44 0 01.573-.99c.02-.029.086-.075.195-.045l1.103.303c.559.153 1.112.008 1.529-.27.16-.107.327-.204.5-.29.449-.222.851-.628.998-1.189l.289-1.105c.029-.11.101-.143.137-.146zM8 0c-.236 0-.47.01-.701.03-.743.065-1.29.615-1.458 1.261l-.29 1.106c-.017.066-.078.158-.211.224a5.994 5.994 0 00-.668.386c-.123.082-.233.09-.3.071L3.27 2.776c-.644-.177-1.392.02-1.82.63a7.977 7.977 0 00-.704 1.217c-.315.675-.111 1.422.363 1.891l.815.806c.05.048.098.147.088.294a6.084 6.084 0 000 .772c.01.147-.038.246-.088.294l-.815.806c-.474.469-.678 1.216-.363 1.891.2.428.436.835.704 1.218.428.609 1.176.806 1.82.63l1.103-.303c.066-.019.176-.011.299.071.213.143.436.272.668.386.133.066.194.158.212.224l.289 1.106c.169.646.715 1.196 1.458 1.26a8.094 8.094 0 001.402 0c.743-.064 1.29-.614 1.458-1.26l.29-1.106c.017-.066.078-.158.211-.224a5.98 5.98 0 00.668-.386c.123-.082.233-.09.3-.071l1.102.302c.644.177 1.392-.02 1.82-.63.268-.382.505-.789.704-1.217.315-.675.111-1.422-.364-1.891l-.814-.806c-.05-.048-.098-.147-.088-.294a6.1 6.1 0 000-.772c-.01-.147.039-.246.088-.294l.814-.806c.475-.469.679-1.216.364-1.891a7.992 7.992 0 00-.704-1.218c-.428-.609-1.176-.806-1.82-.63l-1.103.303c-.066.019-.176.011-.299-.071a5.991 5.991 0 00-.668-.386c-.133-.066-.194-.158-.212-.224L10.16 1.29C9.99.645 9.444.095 8.701.031A8.094 8.094 0 008 0zm1.5 8a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0zM11 8a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
    </button>`),
  },
];
}

const columnsAccessLog = () => {
  return  [
  {
    id: "response",
    name: $_("Js.RespCode"),
    width: "6%",
    formatter: (cell) => cell ? formatCode(cell) : "",
  },
  {
    id: "_timestamp",
    name: $_("Js.Time"),
    width: "15%",
    formatter: (cell) => 
    cell ?
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ) : "",
    convert: true,
  },
  {
    id: "verb",
    name: $_("Js.Request"),
    width: "7%",
  },
  {
    id: "bytes",
    name: $_("Js.Size"),
    width: "6%",
  },
  {
    id: "clientip",
    name: $_("Js.Client"),
    width: "25%",
  },
  {
    id: "clientip_geo_country",
    name: $_("Js.Country"),
    width: "6%",
  },
  {
    id: "request",
    name: $_("Js.Path"),
    width: "26%",
  },
  {
    id: "copy",
    name: "C",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Copy">
      <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"></path><path fill-rule="evenodd" d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"></path></svg>
    </button>`),
  },
  {
    id: "memo",
    name: "M",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Memo">
    <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>
    </button>`),
  },
  {
    id: "extractor",
    name: "E",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Extract">
    <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M7.429 1.525a6.593 6.593 0 011.142 0c.036.003.108.036.137.146l.289 1.105c.147.56.55.967.997 1.189.174.086.341.183.501.29.417.278.97.423 1.53.27l1.102-.303c.11-.03.175.016.195.046.219.31.41.641.573.989.014.031.022.11-.059.19l-.815.806c-.411.406-.562.957-.53 1.456a4.588 4.588 0 010 .582c-.032.499.119 1.05.53 1.456l.815.806c.08.08.073.159.059.19a6.494 6.494 0 01-.573.99c-.02.029-.086.074-.195.045l-1.103-.303c-.559-.153-1.112-.008-1.529.27-.16.107-.327.204-.5.29-.449.222-.851.628-.998 1.189l-.289 1.105c-.029.11-.101.143-.137.146a6.613 6.613 0 01-1.142 0c-.036-.003-.108-.037-.137-.146l-.289-1.105c-.147-.56-.55-.967-.997-1.189a4.502 4.502 0 01-.501-.29c-.417-.278-.97-.423-1.53-.27l-1.102.303c-.11.03-.175-.016-.195-.046a6.492 6.492 0 01-.573-.989c-.014-.031-.022-.11.059-.19l.815-.806c.411-.406.562-.957.53-1.456a4.587 4.587 0 010-.582c.032-.499-.119-1.05-.53-1.456l-.815-.806c-.08-.08-.073-.159-.059-.19a6.44 6.44 0 01.573-.99c.02-.029.086-.075.195-.045l1.103.303c.559.153 1.112.008 1.529-.27.16-.107.327-.204.5-.29.449-.222.851-.628.998-1.189l.289-1.105c.029-.11.101-.143.137-.146zM8 0c-.236 0-.47.01-.701.03-.743.065-1.29.615-1.458 1.261l-.29 1.106c-.017.066-.078.158-.211.224a5.994 5.994 0 00-.668.386c-.123.082-.233.09-.3.071L3.27 2.776c-.644-.177-1.392.02-1.82.63a7.977 7.977 0 00-.704 1.217c-.315.675-.111 1.422.363 1.891l.815.806c.05.048.098.147.088.294a6.084 6.084 0 000 .772c.01.147-.038.246-.088.294l-.815.806c-.474.469-.678 1.216-.363 1.891.2.428.436.835.704 1.218.428.609 1.176.806 1.82.63l1.103-.303c.066-.019.176-.011.299.071.213.143.436.272.668.386.133.066.194.158.212.224l.289 1.106c.169.646.715 1.196 1.458 1.26a8.094 8.094 0 001.402 0c.743-.064 1.29-.614 1.458-1.26l.29-1.106c.017-.066.078-.158.211-.224a5.98 5.98 0 00.668-.386c.123-.082.233-.09.3-.071l1.102.302c.644.177 1.392-.02 1.82-.63.268-.382.505-.789.704-1.217.315-.675.111-1.422-.364-1.891l-.814-.806c-.05-.048-.098-.147-.088-.294a6.1 6.1 0 000-.772c-.01-.147.039-.246.088-.294l.814-.806c.475-.469.679-1.216.364-1.891a7.992 7.992 0 00-.704-1.218c-.428-.609-1.176-.806-1.82-.63l-1.103.303c-.066.019-.176.011-.299-.071a5.991 5.991 0 00-.668-.386c-.133-.066-.194-.158-.212-.224L10.16 1.29C9.99.645 9.444.095 8.701.031A8.094 8.094 0 008 0zm1.5 8a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0zM11 8a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
    </button>`),
  },
  ];
}

const formatWinLevel = (level) => {
  switch (level * 0) {
    case 1:
    case 2:
      return html(`<div class="color-fg-danger">${$_('Js.Error')}(${level})</div>`);
    case 3:
      return html(`<div class="color-fg-attention">${$_('Js.Warnning')}</div>`);
  }
  return html(`<div class="color-fg-default">${$_('Js.Normal')}</div>`);
};

const columnsWindowsLog = () => {
  return  [
  {
    id: "level",
    name: "Level",
    width: "8%",
    formatter: (cell) => cell ? formatWinLevel(cell) : "",
  },
  {
    id: "_timestamp",
    name: $_("Js.Time"),
    width: "15%",
    formatter: (cell) =>
    cell ?
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}"
      ) : "",
    convert: true,
  },
  {
    id: "winComputer",
    name: $_("Js.Computer"),
    width: "20%",
  },
  {
    id: "winEventID",
    name: $_("Js.EventID"),
    width: "9%",
  },
  {
    id: "winEventRecordID",
    name: $_("Js.RecordID"),
    width: "10%",
  },
  {
    id: "winChannel",
    name: $_("Js.Channel"),
    width: "15%",
  },
  {
    id: "winProvider",
    name: $_("Js.Provider"),
    width: "20%",
  },
  {
    id: "copy",
    name: "C",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Copy">
      <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"></path><path fill-rule="evenodd" d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"></path></svg>
    </button>`),
  },
  {
    id: "memo",
    name: "M",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Memo">
    <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>
    </button>`),
  },
  {
    id: "extractor",
    name: "E",
    sort: false,
    width: "3%",
    formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Extract">
    <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M7.429 1.525a6.593 6.593 0 011.142 0c.036.003.108.036.137.146l.289 1.105c.147.56.55.967.997 1.189.174.086.341.183.501.29.417.278.97.423 1.53.27l1.102-.303c.11-.03.175.016.195.046.219.31.41.641.573.989.014.031.022.11-.059.19l-.815.806c-.411.406-.562.957-.53 1.456a4.588 4.588 0 010 .582c-.032.499.119 1.05.53 1.456l.815.806c.08.08.073.159.059.19a6.494 6.494 0 01-.573.99c-.02.029-.086.074-.195.045l-1.103-.303c-.559-.153-1.112-.008-1.529.27-.16.107-.327.204-.5.29-.449.222-.851.628-.998 1.189l-.289 1.105c-.029.11-.101.143-.137.146a6.613 6.613 0 01-1.142 0c-.036-.003-.108-.037-.137-.146l-.289-1.105c-.147-.56-.55-.967-.997-1.189a4.502 4.502 0 01-.501-.29c-.417-.278-.97-.423-1.53-.27l-1.102.303c-.11.03-.175-.016-.195-.046a6.492 6.492 0 01-.573-.989c-.014-.031-.022-.11.059-.19l.815-.806c.411-.406.562-.957.53-1.456a4.587 4.587 0 010-.582c.032-.499-.119-1.05-.53-1.456l-.815-.806c-.08-.08-.073-.159-.059-.19a6.44 6.44 0 01.573-.99c.02-.029.086-.075.195-.045l1.103.303c.559.153 1.112.008 1.529-.27.16-.107.327-.204.5-.29.449-.222.851-.628.998-1.189l.289-1.105c.029-.11.101-.143.137-.146zM8 0c-.236 0-.47.01-.701.03-.743.065-1.29.615-1.458 1.261l-.29 1.106c-.017.066-.078.158-.211.224a5.994 5.994 0 00-.668.386c-.123.082-.233.09-.3.071L3.27 2.776c-.644-.177-1.392.02-1.82.63a7.977 7.977 0 00-.704 1.217c-.315.675-.111 1.422.363 1.891l.815.806c.05.048.098.147.088.294a6.084 6.084 0 000 .772c.01.147-.038.246-.088.294l-.815.806c-.474.469-.678 1.216-.363 1.891.2.428.436.835.704 1.218.428.609 1.176.806 1.82.63l1.103-.303c.066-.019.176-.011.299.071.213.143.436.272.668.386.133.066.194.158.212.224l.289 1.106c.169.646.715 1.196 1.458 1.26a8.094 8.094 0 001.402 0c.743-.064 1.29-.614 1.458-1.26l.29-1.106c.017-.066.078-.158.211-.224a5.98 5.98 0 00.668-.386c.123-.082.233-.09.3-.071l1.102.302c.644.177 1.392-.02 1.82-.63.268-.382.505-.789.704-1.217.315-.675.111-1.422-.364-1.891l-.814-.806c-.05-.048-.098-.147-.088-.294a6.1 6.1 0 000-.772c-.01-.147.039-.246.088-.294l.814-.806c.475-.469.679-1.216.364-1.891a7.992 7.992 0 00-.704-1.218c-.428-.609-1.176-.806-1.82-.63l-1.103.303c-.066.019-.176.011-.299-.071a5.991 5.991 0 00-.668-.386c-.133-.066-.194-.158-.212-.224L10.16 1.29C9.99.645 9.444.095 8.701.031A8.094 8.094 0 008 0zm1.5 8a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0zM11 8a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
    </button>`),
  },
  ];
}

const makeDataColumns = (fields) => {
  const colums = [];
  fields.forEach((f) => {
    if (f == "time" || f.startsWith("_") || f.endsWith("_geo")) {
      return;
    }
    colums.push({
      id: f,
      name: getFieldName(f),
    });
  });
  colums.sort((a,b)=> {
    return a.id < b.id ? -1 : 1;
  });
  colums.unshift({
    id: "time",
    name: $_("Js.Time"),
    formatter: (cell) =>
    cell ? 
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}"
      ) : "",
    convert: true,
  });
  colums.push(
    {
      id: "copy",
      name: "C",
      sort: false,
      width: "3%",
      formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Copy">
        <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"></path><path fill-rule="evenodd" d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"></path></svg>
      </button>`),
    }
  );
  colums.push(
      {
      id: "memo",
      name: "M",
      sort: false,
      width: "3%",
      formatter: (cell) => html(`<button class="btn-octicon" type="button" aria-label="Memo">
      <svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>
      </button>`),
    },
  );
  return colums;
};

export const getLogColums = (view, fields) => {
  switch (view) {
    case "syslog":
      return columnsSyslog();
    case "access":
      return columnsAccessLog();
    case "windows":
      return columnsWindowsLog();
    case "data":
    case "ex_data":
      return makeDataColumns(fields);
  }
  return columnsTimeOnly();
};

const getAccessLogData = (r, filter) => {
  const d = [];
  r.Logs.forEach((l) => {
    if (filter && filter.st) {
      if (l.Time < filter.st || l.Time > filter.et) {
        return;
      }
    }
    d.push([
      l.KeyValue.response,
      l.Time,
      l.KeyValue.verb,
      l.KeyValue.bytes,
      l.KeyValue.clientip_host
        ? l.KeyValue.clientip + "(" + l.KeyValue.clientip_host + ")"
        : l.KeyValue.clientip,
      l.KeyValue.clientip_geo_country || "",
      l.KeyValue.request,
      "copy",
      "memo",
      l.All,
    ]);
  });
  return d;
};

export const getLogLevel = (l) => {
  let suverity = l.KeyValue.suverity || l.KeyValue.priority;
  if (suverity && suverity != "") {
    suverity %= 8;
    return suverity < 4 ? "error" : suverity == 4 ? "warn" : "normal";
  }
  const code = l.KeyValue.response;
  if (code > 99) {
    return code < 300 ? "normal" : code < 400 ? "warn" : "error";
  }

  let winLevel = l.KeyValue.winLevel;
  if (winLevel != undefined) {
    return winLevel == 1 || winLevel == 2
      ? "error"
      : winLevel == 3
      ? "warn"
      : "normal";
  }
  const level = l.KeyValue.suverity_str || l.KeyValue.level || l.All;
  if (/(alert|error|crit|fatal|emerg|failure|err )/i.test(level)) {
    return "error";
  }
  if (/warn/i.test(level)) {
    return "warn";
  }
  return "normal";
};

const getSyslogData = (r, filter) => {
  const d = [];
  r.Logs.forEach((l) => {
    if (filter && filter.st) {
      if (l.Time < filter.st || l.Time > filter.et) {
        return;
      }
    }
    const message = l.KeyValue.message || "";
    const pid = l.KeyValue.pid || "";
    const tag =
      l.KeyValue.tag ||
      (l.KeyValue.program || "") + (pid ? "[" + pid + "]" : "");
    const src = l.KeyValue.logsource || "";
    const level = getLogLevel(l);
    d.push([level, l.Time, src, tag, message,"copy","memo",l.All]);
  });
  return d;
};

const getExtractData = (r, filter, fields) => {
  const d = [];
  fields.sort((a,b) => a > b);
  r.Logs.forEach((l) => {
    if (filter && filter.st) {
      if (l.Time < filter.st || l.Time > filter.et) {
        return;
      }
    }
    const ent = { time: l.Time };
    fields.forEach((k) => {
      if (k == "time" || k.startsWith("_") || k.endsWith("_geo")) {
        return;
      }
      ent[k] = l.KeyValue[k] || "";
    });
    d.push(ent);
  });
  return d;
};

const getWindowsLogData = (r, filter) => {
  const d = [];
  r.Logs.forEach((l) => {
    if (filter && filter.st) {
      if (l.Time < filter.st || l.Time > filter.et) {
        return;
      }
    }
    d.push([
      l.KeyValue.winLevel || 0,
      l.Time,
      l.KeyValue.winComputer || "",
      l.KeyValue.winEventID || 0,
      l.KeyValue.winEventRecordID || 0,
      l.KeyValue.winChannel || "",
      l.KeyValue.winProvider || "",
      "copy",
      "memo",
      l.All,
    ]);
  });
  return d;
};

export const getLogData = (r, view, filter) => {
  switch (view) {
    case "syslog":
      return getSyslogData(r, filter);
    case "access":
      return getAccessLogData(r, filter);
    case "windows":
      return getWindowsLogData(r, filter);
    case "data":
      return getExtractData(r, filter,r.Fields);
    case "ex_data":
      return getExtractData(r, filter,r.ExFields);
    case "anomaly":
      return getTimeOnlyLogData(r, filter, "anomalyScore");
  }
  return getTimeOnlyLogData(r, filter);
};

let timeIndex = 1;

const gridSearch = {
  enable: true,
  selector: (cell, rowIndex, cellIndex) =>
    cellIndex == timeIndex
      ? echarts.time.format(
          new Date(cell / (1000 * 1000)),
          "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}"
        )
      : cell,
};

export const getGridSearch = (view) => {
  timeIndex =  view == "data" ? 0 : 1;
  return gridSearch;
};
