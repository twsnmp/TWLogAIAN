# TWLogAIAN
TWSNMP`s Log AI Analyzer
AIアシスト付きのログ分析ツール

## Overview
TWSNMP FCで開発したsyslogの分析機能を拡張してデスクトップアプリケーションとして開発するプロジェクトです。

分析対象のログデーターは、
- ローカルファイル
- SFTP経由のリモートファイル
- TWSNMP FCに保存したログ
- Splunk/Gravwellなどのログサーバー
- 何らかのデータベース？
を予定しています。

ログの分析機能としては、
- ログの種類を自動判別する
- タイムスタンプを自動取得する
- 正規表現や簡易な方法でフィルターできる
- ログの中から特定パターンのデータを抽出できる
- ログの件数や抽出したデータをグラフ表示できる
- AIを使って異常検知、予測などができる
- 分析結果をCSVやEXCELに出力できる
を予定しています。

デスクトップアプリを開発するための使う予定の技術は
- GO言語
- Wails v2
- Svelte
- p5.js
- Apache echarts
- Primer/CSS,Octicons
- TensorFlow.js
です。

バックエンドをGO言語により並列/高速処理し、JS/CSS/HTMLのフロントエンドにより豊かな表現力を実現します。

## Status

開発を始めたばかりで、説明に目標を書いただけです。


## Copyright

see ./LICENSE

```
Copyright 2022 Masayuki Yamai
```
