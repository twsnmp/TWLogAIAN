---
title: How to Write Search Queries
layout: default
---

# How to Write Search Queries

This article serves as a reference/help guide for the search query syntax in the Log Analysis Tool. The help button next to the search query input box links directly to this document.

[Back](./index.html)

## Search Feature Differences by Storage Engine

In TWLogAIAN v2.1.0, you can choose **Parquet**, **Badger**, or **Bbolt** as the log storage engine in addition to **Bluge** (the traditional indexed engine).
Depending on the selected storage engine, search capabilities behave as follows:

| Storage Engine | Full-Text Search Mode | Regular Expression (Regexp) Behavior | Features & Recommended Use |
| :--- | :---: | :---: | :--- |
| **Bluge** (Default) | **Available** | △ Limited (token-based) | Best for fast keyword, field, and range queries using the full-text search index. However, regular expressions match against tokenized terms in the index, which may not match across arbitrary log structures as expected. |
| **Non-Bluge**<br>(Parquet / Badger / Bbolt) | **Not Available**<br>(Hidden) | **◎ Fully & Accurately Works** | **Because full-text indexes are not built, Full-Text Search Mode is unavailable.** However, queries are evaluated directly against the raw log string using Go's standard regular expression engine, meaning **regular expressions work accurately and reliably.** Ideal for strict and complex regex pattern filtering. |

*Note: When using a storage engine other than Bluge, "Full Text Search" is automatically omitted from the search mode selection menu.*

## Search Query Mode Switching
You can switch search query modes from the Advanced Settings menu inside the search settings pane:

- **Simple**: Intuitive filtering using space-separated AND keywords, NOT conditions, and prefix wildcards.
- **Regexp**: Direct regular expression pattern matching (works with full fidelity on non-Bluge engines).
- **AI**: Describe what you are looking for in natural language (Japanese or English), and the LLM will generate an optimal regular expression.
- **Full Text**: Advanced query syntax using the Bluge query string parser (Bluge only).

## Simple Mode
Simple Mode is available across all storage engines and follows only a few rules:
- Key phrases separated by spaces function as an **AND** condition.
- Prefixing a keyword with `!` functions as a **NOT** condition.
- Suffixing a keyword with `*` functions as a prefix (wildcard) search.

For example:
```
test !mode sta*
```
translates to the following condition:
> Contains `test`, does not contain `mode`, and contains a word starting with `sta`.

*Note: With Bluge, due to word boundary tokenization, words containing `_` or `:` might not match wildcard queries.*

## Regular Expression Mode
Performs a search by treating the input query directly as a regular expression.

> [!NOTE]
> **Behavior by Storage Engine**:
> - **When using Parquet / Badger / Bbolt (Non-Bluge)**:
>   Matching is evaluated against the entire raw log line using Go's standard regex engine (RE2-compatible), so **regular expressions function completely and accurately**.
>   Examples: `^.*ERROR.*\[code=\d+\].*$` or `(Failed|Invalid) password for (invalid user )?\w+`
> - **When using Bluge**:
>   Regular expressions are evaluated against individual tokens within the search index, so patterns expecting full raw log structure may not match as intended.

## AI Regular Expression Mode (Added in v2.1.0)
Available when an LLM provider (the embedded local LLM tensai, or Ollama / Gemini / OpenAI / Anthropic) is enabled in settings.

Enter a description of what you want to find in **natural language (English or Japanese)** into the search bar. The AI will formulate an optimal regular expression and immediately run the search.
You do not need to memorize complex regex syntax. For example:

- `authentication failed logs`
- `access from IP address starting with 192.168.`
- `HTTP status code 500 series errors`
- `timeout or disconnected connection`

The generated regular expression is populated into the search query input and can also be inspected and tweaked in the Advanced Settings pane.

## Full-Text Search Mode (Bluge Only)
This mode is available only when the storage engine is set to **Bluge**. It uses the standard Bleve query string query syntax (same as Bluge). It supports the following specifications:

### Term Search
A single term with no special syntax matches documents (logs) containing that term.
For example:
```
water
```
searches for documents containing the term `water` in any field (`_all`).

### Phrase Search
To search for an exact sequence of words, wrap them in double quotes.
For example:
```
"light beer"
```
searches for the exact phrase "light beer".

### Field Specification
You can limit the search to a specific field by prefixing the search term with the field name followed by a colon.
For example:
```
description:water
```
searches for documents where the `description` field contains `water`.

### Wildcard
Used to match parts of a word using `*`.
For example:
```
mart*
```
searches for documents containing words starting with `mart`. This can also be used with field specifications.

### Regular Expressions
You can use regular expressions by wrapping the pattern with `/`.
For example:
```
/light (beer|wine)/
```
To specify a field:
```
description:/wat.*/
```

### Required, Optional, and Excluded
By default, terms are optional.
- Prefixing a term with `+` makes it **Required**.
- Prefixing a term with `-` makes it **Excluded** (must not be present).

For example:
```
+description:water -light beer
```
means that `water` is required in the `description` field, `light` must not be present anywhere in the document, and `beer` is optional.

### Boosting (Priority Specification)
You can influence the relevance score of specific query parts using `^` followed by a numeric value.
For example:
```
description:water name:water^5
```
This increases the score of documents containing `water` in the `name` field by a factor of 5 compared to those containing it in the `description` field.

### Fuzziness
You can perform a fuzzy search by suffixing a term with `~` followed by a number.
For example:
```
watex~2
```
*Note: The exact usage details are currently unspecified.*

### Numeric Range
You can search numeric fields using `>`, `>=`, `<`, `<=` operators.
For example:
```
abv:>10
```
searches for documents where the `abv` field value is greater than 10.

### Date/Time Range
You can specify date and time ranges using `>`, `>=`, `<`, `<=` operators.
For example:
```
created:>"2016-09-21"
```
searches for documents where the `created` field date is after September 21, 2016.

### Escape Characters
The following characters must be escaped with a backslash `\` if you want to search them literally:
`"+-=&|><!(){}[]^\"~*?:\\/ "`

For example:
```
my\ name
```
or
```
"contains a\" character"
```
This allows spaces or double quotes to be included within search terms.

### References
- [Query String Query -- Bleve](https://blevesearch.com/docs/Query-String-Query/)
- [blugelabs/query_string parser tests on GitHub](https://github.com/blugelabs/query_string/blob/master/query_string_parser_test.go)

## Time Range Specification
Separately from the search query, you can specify the time range in the graphical user interface.

### Range
Allows you to specify the start and end dates/times down to the minute.

### Target
Allows you to search for logs within a specified duration (in seconds) around a target date and time.
