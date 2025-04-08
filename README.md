# TWLogAIAN

TWSNMP`s Log AI Analyzer
AIアシスト付きのログ分析ツール

## Overview

This is a project that expands the analysis function of syslog developed with TWSNMP FC and develops it as a desktop application.
Log analysis is overwhelmingly more effective than having files provided and doing it on your own computer than by accessing the log server.
I think there are a lot.If you are a Unix command expert, you use commands to search and format logs to perform analysis.
I think so.I'm sure there are many people who load it into the editor and work with the editor search function.I'm using Excel
I think there are people too.This tool is a tool to help such people a little bit.Load the provided log files directly from compressed state, directly from a remote server, or directly from a Docker or Kubernetes command, and create indexes and make them searchable using a full-text search engine.The index is from the IP address extracted from the log.
It can also include additional information such as location information and hostname.You can also easily visualize the results of your search using web technology.All you have to do is delete the log data as a whole folder after the analysis.

TWSNMP FCで開発したsyslogの分析機能を拡張してデスクトップアプリケーションとして開発するプロジェクトです。
ログ分析はログサーバーにアクセスして行うよりもファイルを提供してもらって自分のパソコンで行うほうが圧倒的に
多いと思います。Unixのコマンドの達人ならばコマンドを駆使してログを検索したり整形したりして分析を行っていると
と思います。エディターに読み込んでエディタの検索機能などで作業している人も多いと思います。Excelを使っている
人もいると思います。このツールは、そのような人を少しだけ手助けするためのツールです。提供されたログ・ファイルを圧縮された状態から直接読み込んだり、リモートのサーバーから直接読み込んだり、DockerやKubernetesのコマンドから直接読み込んで全文検索エンジンでインデックスを作成し検索可能にします。インデックスにはログから抽出したIPアドレスから
位置情報やホスト名などの補足情報を含めることもできます。検索した結果もWebの技術を使って簡単にビジュアライズできます。分析が終わったログデータはフォルダごと削除すればよいだけです。

![2022-03-01_06-22-51](https://user-images.githubusercontent.com/5225950/156246976-ca92f7eb-686c-4bc5-bafd-0053a74f3b88.png)


The log file to be analyzed is

- Local files
- Local Directory
- Local command execution results (Docker and Kubernetes)
- SCP forwarding from a remote server
- Results of command execution using SSH on a remote server (Docker and Kubernetes)
- TWSNMP FC integration (from v1.1.0)
- Windows Event Log (from v1.1.0)

can be obtained by:

The log analysis function is

- Automatically determine the log type
- Automatically retrieve timestamps
- Filter with regular expressions and simple methods
- Extract data from a specific pattern from the log
- Location information and host name can be estimated from extracted IP addresses
- Vendor name can be estimated from extracted MAC addresses
- Logs and extracted data can be indexed using a full-text search engine
- Full-text search engine allows you to search by time range, keyword, numeric range, and geographical location range
- Visualize the number of logs and extracted data on graphs and world maps
- Histogram
- Cluster
- Time series
- World Map
- Globe
- Heatmap (v1.1.0)
- You can output analysis results to CSV or EXCEL
- Select logs to create chronologically arranged notes (v1.2.0)
- Anomaly logs can be detected with AI (machine learning) assistance (from v1.3.0)
- Data can be extracted during searching

It has been realized.


The techniques used to develop desktop apps

- GO Language
- Wails v2: GO Language GUI creation tool
- Svelte: Web Application Framework
- Bluge: GO Language Full Text Search Engine
- p5.js : Visualize
- Apache echarts: graph display
- Primer/CSS,Octicons: Screen design
- TensorFlow.js : AI

is.


The backend is processed in parallel/fastly using GO language, and the JS/CSS/HTML frontend achieves rich expressiveness.

分析対象のログファイルは、

- ローカルファイル
- ローカルディレクトリ
- ローカルでのコマンド実行結果(DockerやKubernetes)
- リモートサーバーからSCPで転送
- リモートサーバーでのSSHによるコマンド実行結果(DockerやKubernetes)
- TWSNMP FC連携 (v1.1.0から)
- Windowsイベントログ(v1.1.0から)

により取得できます。

ログの分析機能としては、

- ログの種類を自動判別する
- タイムスタンプを自動取得する
- 正規表現や簡易な方法でフィルターできる
- ログの中から特定パターンのデータを抽出できる
- 抽出したIPアドレスから位置情報、ホスト名を推定できる
- 抽出したMACアドレスからベンダ名を推定できる
- ログと抽出したデータを全文検索エンジンでインデックス作成できる
- 全文検索エンジンで、時間範囲、キワード、数値範囲、地理的な位置範囲で検索できる
- ログの件数や抽出したデータをグラフや世界地図上にビジュアライズできる
	- ヒストグラム
	- クラスター
	- 時系列
	- 世界地図
	- 地球儀
	- ヒートマップ(v1.1.0)
- 分析結果をCSVやEXCELに出力できる
- ログを選択して時系列に並べたメモを作成できる(v1.2.0)
- AI(機械学習)のアシストにより異常ログを検知できる(v1.3.0から)
- 検索時にデータ抽出できる

を実現しました。


デスクトップアプリを開発するための使った技術は

- GO言語
- Wails v2 : GO言語版のGUI作成ツール
- Svelte : Webアプリケーションフレームワーク
- Bluge : GO言語版全文検索エンジン
- p5.js : ビジュアライズ
- Apache echarts : グラフ表示
- Primer/CSS,Octicons : 画面デザイン
- TensorFlow.js : AI

です。

バックエンドをGO言語により並列/高速処理し、JS/CSS/HTMLのフロントエンドにより豊かな表現力を実現します。

## Document

https://twsnmp.github.io/TWLogAIAN/

## Status

v1.0.0(2022/3/2) First release
v1.1.0 (2022/3/14) External link
v1.2.0 (2022/3/21) Supports memo function
v1.3.0 (2022/4/3) AI assist compatible
v1.4.0 (2022/4/11) Improved Grok pattern editing function
v1.5.0 (2022/4/24) Compatible with data extraction during search, improved Grok pattern editing function
v1.6.0 (2022/10/29) Grok pattern, improved field editing function, automatic log type determination
v1.7.0 (2023/1/15) English support, improved log search function
v1.8.0 (2/5/2023) Improved Grok pattern editing and selection, improved timestamp processing
v1.9.0 (2/12/2023) Improved Windows event log processing, highlighting
v1.10.0 (2023/6/12) Improved Windows event log processing, abnormality detection using TFIDF

v1.0.0(2022/3/2) 最初のリリース
v1.1.0(2022/3/14) 外部連携
v1.2.0(2022/3/21) メモ機能対応
v1.3.0(2022/4/3) AIアシスト対応
v1.4.0(2022/4/11) Grokパターン編集機能の改善
v1.5.0(2022/4/24) 検索時データ抽出に対応、Grokパターン編集機能の改善
v1.6.0(2022/10/29) Grokパターン、フィールド編集機能の改善、ログ種別の自動判定
v1.7.0(2023/1/15) 英語対応、ログ検索機能の改善
v1.8.0(2023/2/5) Grokパターン編集と選択の改善、タイムスタンプ処理の改善
v1.9.0(2023/2/12) Windowsイベントログ処理の改善、ハイライト表示
v1.10.0(2023/6/12) Windowsイベントログ処理の改善、TFIDFによる異常検知

## Build

Wails v2 installation is required for the build.

https://wails.io/docs/gettingstarted/installation/

The build will be done with make.

ビルドのためには、ｗails v2のインストールが必要です。

https://wails.io/docs/gettingstarted/installation/

ビルドはmakeで行います。

```
$make
```

The following targets can be specified:
````
all Build all executable files (optional)
Building an executable file for mac Mac
Build an executable file for windows
windebug Build debug version executable file for Windows
clean Delete the built executable file
dev Starting the debugging environment
````

以下のターゲットが指定できます。
```
	all        全実行ファイルのビルド（省略可能）
	mac        Mac用の実行ファイルのビルド
	windows    Windows用の実行ファイルのビルド
	windebug   Windows用のデバック版の実行ファイルのビルド
	clean      ビルドした実行ファイルの削除
	dev        デバッグ環境の起動
```

```
$make
```

An executable file for MacOS and Windows will be created in the `build/bin` directory.

To start for debugging
```
$make dev
```


を実行すれば、MacOS,Windows用の実行ファイルが、`build/bin`のディレクトリに作成されます。

デバッグ用に起動するためには
```
$make dev
```
を実行します。


## Copyright

see ./LICENSE

```
Copyright 2022-2025 Masayuki Yamai
```
