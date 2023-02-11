# TWLogAIAN
TWSNMP`s Log AI Analyzer
AIアシスト付きのログ分析ツール

## Overview

TWSNMP FCで開発したsyslogの分析機能を拡張してデスクトップアプリケーションとして開発するプロジェクトです。
ログ分析はログサーバーにアクセスして行うよりもファイルを提供してもらって自分のパソコンで行うほうが圧倒的に
多いと思います。Unixのコマンドの達人ならばコマンドを駆使してログを検索したり整形したりして分析を行っていると
と思います。エディターに読み込んでエディタの検索機能などで作業している人も多いと思います。Excelを使っている
人もいると思います。このツールは、そのような人を少しだけ手助けするためのツールです。提供されたログ・ファイルを圧縮された状態から直接読み込んだり、リモートのサーバーから直接読み込んだり、DockerやKubernetesのコマンドから直接読み込んで全文検索エンジンでインデックスを作成し検索可能にします。インデックスにはログから抽出したIPアドレスから
位置情報やホスト名などの補足情報を含めることもできます。検索した結果もWebの技術を使って簡単にビジュアライズできます。分析が終わったログデータはフォルダごと削除すればよいだけです。

![2022-03-01_06-22-51](https://user-images.githubusercontent.com/5225950/156246976-ca92f7eb-686c-4bc5-bafd-0053a74f3b88.png)


分析対象のログファイルは、

- ローカルファイル
- ローカルディレクトリ
- ローカルでのコマンド実行結果(DockerやKubernetes)
- リモートサーバーからSCPで転送
- リモートサーバーでのSSHによるコマンド実行結果(DockerやKubernetes)
- TWSNMP FC連携 (v1.1.0から)
- Windowsイベントログ(v1.1.0から)

により取得できます。

※ Gravwell連携(v1.1.0で追加しましたが、v1.7.０で削除しました。）


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

今後、

- 機械学習やAIの適用を強化する
- ログの分析のためのビジュアライズ機能を強化する

を予定しています。


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

## Status

v1.0.0(2022/3/2) 最初のリリース
v1.1.0(2022/3/14) 外部連携
v1.2.0(2022/3/21) メモ機能対応
v1.3.0(2022/4/3) AIアシスト対応
v1.4.0(2022/4/11) Grokパターン編集機能の改善
v1.5.0(2022/4/24) 検索時データ抽出に対応、Grokパターン編集機能の改善
v1.6.0(2022/10/29) Grokパターン、フィールド編集機能の改善、ログ種別の自動判定
v1.7.0(2022/1/15) 英語対応、ログ検索機能の改善
v1.8.0(2022/2/5) Grokパターン編集と選択の改善、タイムスタンプ処理の改善
v1.9.0(2022/2/12) Windowsイベントログ処理の改善、ハイライト表示
## Build

ビルドのためには、ｗails v2のインストールが必要です。
https://wails.io/docs/gettingstarted/installation/

ビルドはmakeで行います。
```
$make
```
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
を実行すれば、MacOS,Windows用の実行ファイルが、`build/bin`のディレクトリに作成されます。

デバッグ用に起動するためには
```
$make dev
```
を実行します。


## Copyright

see ./LICENSE

```
Copyright 2022 Masayuki Yamai
```
