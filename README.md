# TWLogAIAN

TWSNMP`s Log AI Analyzer

[日本語のドキュメント (README_ja.md)](README_ja.md)

## Overview

This is a project that expands the analysis function of syslog developed with TWSNMP FC and develops it as a desktop application.
Log analysis is overwhelmingly more effective than having files provided and doing it on your own computer than by accessing the log server.
I think there are a lot.If you are a Unix command expert, you use commands to search and format logs to perform analysis.
I think so.I'm sure there are many people who load it into the editor and work with the editor search function.I'm using Excel
I think there are people too.This tool is a tool to help such people a little bit.Load the provided log files directly from compressed state, directly from a remote server, or directly from a Docker or Kubernetes command, and create indexes and make them searchable using a full-text search engine.The index is from the IP address extracted from the log.
It can also include additional information such as location information and hostname.You can also easily visualize the results of your search using web technology.All you have to do is delete the log data as a whole folder after the analysis.

![Log Analysis Flow](docs/images/en/log_analyzer.png)

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
v2.0.0 (2026/6/20) LLM

## Build

Wails v2 installation is required for the build.

https://wails.io/docs/gettingstarted/installation/

The build will be done with make.

```
$make
```

The following targets can be specified:
```
all Build all executable files (optional)
mac Building an executable file for mac
windows Build an executable file for windows
windebug Build debug version executable file for Windows
clean Delete the built executable file
dev Starting the debugging environment
```

An executable file for MacOS and Windows will be created in the `build/bin` directory.

To start for debugging
```
$make dev
```

## Copyright

see ./LICENSE

```
Copyright 2022-2026 Masayuki Yamai
```
