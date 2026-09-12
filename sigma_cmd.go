package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/twsnmp/TWLogAIAN/pkg/datastore"
	"github.com/twsnmp/TWLogAIAN/pkg/sigma"
	"gopkg.in/yaml.v3"
)

func runSigmaCLI(args []string) {
	if len(args) == 0 {
		printSigmaHelp()
		return
	}

	action := args[0]
	restArgs := args[1:]

	fs := flag.NewFlagSet("sigma", flag.ExitOnError)
	var (
		rulesPath     string
		packsStr      string
		configPath    string
		datastorePath string
		filterPack    string
		filterCustom  bool
		outputDir     string
	)

	fs.StringVar(&rulesPath, "s", "", "Sigma rules path (file or dir)")
	fs.StringVar(&rulesPath, "rules", "", "Sigma rules path (file or dir)")
	fs.StringVar(&packsStr, "p", "", "Comma-separated Sigma rule packs (e.g. windows-essential,linux-auth or all)")
	fs.StringVar(&packsStr, "packs", "", "Comma-separated Sigma rule packs (e.g. windows-essential,linux-auth or all)")
	fs.StringVar(&configPath, "c", "", "Config directory path")
	fs.StringVar(&configPath, "config", "", "Config directory path")
	fs.StringVar(&datastorePath, "d", "", "Datastore path (logs.parquet, twlogaian.badger, twlogaian_logs.db)")
	fs.StringVar(&datastorePath, "datastore", "", "Datastore path")
	fs.StringVar(&filterPack, "pack", "", "Filter rules by pack name (for list)")
	fs.BoolVar(&filterCustom, "custom", false, "Filter custom rules only (for list)")
	fs.StringVar(&outputDir, "o", "./converted-rules", "Output directory for converted Wazuh rules")

	switch action {
	case "help", "-h", "--help":
		printSigmaHelp()
		return
	case "packs":
		showCLISigmaPacks()
		return
	case "list":
		_ = fs.Parse(restArgs)
		packs := parsePacks(packsStr)
		runCLISigmaList(rulesPath, packs, configPath, filterPack, filterCustom)
		return
	case "stat":
		_ = fs.Parse(restArgs)
		packs := parsePacks(packsStr)
		runCLISigmaStat(rulesPath, packs, configPath)
		return
	case "logsrc":
		_ = fs.Parse(restArgs)
		packs := parsePacks(packsStr)
		runCLISigmaLogSrc(rulesPath, packs, configPath)
		return
	case "field":
		_ = fs.Parse(restArgs)
		packs := parsePacks(packsStr)
		runCLISigmaField(rulesPath, packs, configPath)
		return
	case "test":
		_ = fs.Parse(restArgs)
		testLogs := fs.Args()
		if len(testLogs) == 0 {
			fmt.Println("Usage: twlogaian sigma test [flags] <log1> [log2...]")
			return
		}
		packs := parsePacks(packsStr)
		runCLISigmaTest(rulesPath, packs, configPath, testLogs)
		return
	case "convert-wazuh":
		_ = fs.Parse(restArgs)
		xmlArgs := fs.Args()
		if len(xmlArgs) == 0 {
			fmt.Println("Usage: twlogaian sigma convert-wazuh [flags] <wazuh-rule-file-or-dir>")
			return
		}
		runCLISigmaConvertWazuh(xmlArgs[0], outputDir)
		return
	case "convert-wazuh-decoder":
		_ = fs.Parse(restArgs)
		xmlArgs := fs.Args()
		if len(xmlArgs) == 0 {
			fmt.Println("Usage: twlogaian sigma convert-wazuh-decoder <wazuh-decoder-file-or-dir>")
			return
		}
		runCLISigmaConvertWazuhDecoder(xmlArgs[0])
		return
	case "scan":
		_ = fs.Parse(restArgs)
		packs := parsePacks(packsStr)
		runCLISigmaScan(rulesPath, packs, configPath, datastorePath)
		return
	default:
		// If action doesn't match standard keywords, try parsing as flags for scan
		if strings.HasPrefix(action, "-") {
			_ = fs.Parse(args)
			packs := parsePacks(packsStr)
			runCLISigmaScan(rulesPath, packs, configPath, datastorePath)
			return
		}
		fmt.Printf("unknown sigma action: %s\n\n", action)
		printSigmaHelp()
	}
}

func printSigmaHelp() {
	fmt.Print(`TWLogAIAN Sigma Threat Detection & Compliance Audit Tool
Usage:
  twlogaian sigma <action> [options]

Actions:
  packs                      List available embedded rule packs
  list                       List active rules with metadata
  stat                       Show rule, source, and tag statistics
  logsrc                     List log sources
  field                      List log fields
  test <log1> [log2...]      Test log strings against active rules
  scan [options]             Scan datastore against active rules
  convert-wazuh <xml>        Convert Wazuh XML rules to Sigma YAML
  convert-wazuh-decoder <xml> Convert Wazuh XML decoders to regex patterns

Options:
  -s, --rules <path>         Custom Sigma rules path (file or dir)
  -p, --packs <packs>        Rule packs (comma-separated, e.g. windows-essential,linux-auth or all)
  -c, --config <path>        Custom Sigma field mapping config directory
  -d, --datastore <path>     Datastore path (parquet, badger, bbolt, or work directory)
  --pack <pack>              Filter rules by pack name (for list)
  --custom                   Filter custom rules only (for list)
  -o <dir>                   Output directory for convert-wazuh (default: ./converted-rules)
`)
}

func parsePacks(p string) []string {
	if p == "" {
		return nil
	}
	parts := strings.Split(p, ",")
	var res []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			res = append(res, part)
		}
	}
	return res
}

func initEngine(rulesPath string, packs []string, configPath string) (*sigma.Engine, error) {
	cfg := sigma.Config{
		RulesPath: rulesPath,
		Packs:     packs,
		ConfigDir: configPath,
		Strict:    false,
	}
	return sigma.NewEngine(cfg)
}

func showCLISigmaPacks() {
	packs := sigma.GetAvailableSigmaPacks()
	fmt.Printf("Available Rule Packs (%d):\n", len(packs))
	for _, p := range packs {
		count := 0
		if info, err := sigma.GetSigmaPackInfo(p, false); err == nil {
			count = info.RuleCount
		}
		desc := sigma.GetSigmaPackDescription(p)
		fmt.Printf("  - %-20s (rules: %2d) : %s\n", p, count, desc)
	}
}

func runCLISigmaList(rulesPath string, packs []string, configPath, filterPack string, filterCustom bool) {
	eng, err := initEngine(rulesPath, packs, configPath)
	if err != nil {
		log.Fatalf("failed to initialize sigma engine: %v", err)
	}
	for _, entry := range eng.RuleEntries {
		if filterPack != "" && entry.Source != "pack:"+filterPack {
			continue
		}
		if filterCustom && strings.HasPrefix(entry.Source, "pack:") {
			continue
		}
		e := entry.Evaluator
		lsk := fmt.Sprintf("%s:%s:%s", e.Logsource.Product, e.Logsource.Category, e.Logsource.Service)
		tagsStr := strings.Join(e.Rule.Tags, ",")
		fmt.Printf("%-24s\t%s\t%-8s\t%-20s\t%-30s\t%s\n", entry.Source, e.ID, e.Level, lsk, tagsStr, e.Title)
	}
}

func runCLISigmaStat(rulesPath string, packs []string, configPath string) {
	eng, err := initEngine(rulesPath, packs, configPath)
	if err != nil {
		log.Fatalf("failed to initialize sigma engine: %v", err)
	}
	logSrcMap := make(map[string]int)
	fieldMap := make(map[string]int)
	sourceMap := make(map[string]int)
	tagMap := make(map[string]int)

	for _, entry := range eng.RuleEntries {
		e := entry.Evaluator
		sourceMap[entry.Source]++
		lsk := fmt.Sprintf("%s:%s:%s", e.Logsource.Product, e.Logsource.Category, e.Logsource.Service)
		logSrcMap[lsk]++
		for _, tag := range e.Rule.Tags {
			tagMap[tag]++
		}
		for _, dsv := range e.Rule.Detection.Searches {
			for _, ems := range dsv.EventMatchers {
				for _, em := range ems {
					fk := lsk + ":" + em.Field
					fieldMap[fk]++
				}
			}
		}
	}

	fmt.Printf("Total Rules: %d\n", len(eng.RuleEntries))
	fmt.Printf("Log Sources: %d, Fields: %d, Unique Tags: %d\n\n", len(logSrcMap), len(fieldMap), len(tagMap))
	fmt.Println("Sources:")
	for s, c := range sourceMap {
		fmt.Printf("  %-24s : %d\n", s, c)
	}
	fmt.Println("\nTop Tags (MITRE ATT&CK / Compliance):")
	type kv struct {
		k string
		v int
	}
	var tagList []kv
	for k, v := range tagMap {
		tagList = append(tagList, kv{k, v})
	}
	sort.Slice(tagList, func(i, j int) bool {
		return tagList[i].v > tagList[j].v
	})
	limit := 25
	if len(tagList) < limit {
		limit = len(tagList)
	}
	for i := 0; i < limit; i++ {
		fmt.Printf("  %-35s : %d\n", tagList[i].k, tagList[i].v)
	}
}

func runCLISigmaLogSrc(rulesPath string, packs []string, configPath string) {
	eng, err := initEngine(rulesPath, packs, configPath)
	if err != nil {
		log.Fatalf("failed to initialize sigma engine: %v", err)
	}
	logSrcMap := make(map[string]int)
	for _, entry := range eng.RuleEntries {
		e := entry.Evaluator
		lsk := fmt.Sprintf("%s_%s_%s", e.Logsource.Product, e.Logsource.Category, e.Logsource.Service)
		logSrcMap[lsk]++
	}
	for k, c := range logSrcMap {
		fmt.Printf("%s=%d\n", k, c)
	}
}

func runCLISigmaField(rulesPath string, packs []string, configPath string) {
	eng, err := initEngine(rulesPath, packs, configPath)
	if err != nil {
		log.Fatalf("failed to initialize sigma engine: %v", err)
	}
	fieldMap := make(map[string]int)
	for _, entry := range eng.RuleEntries {
		e := entry.Evaluator
		lsk := fmt.Sprintf("%s_%s_%s", e.Logsource.Product, e.Logsource.Category, e.Logsource.Service)
		for _, dsv := range e.Rule.Detection.Searches {
			for _, ems := range dsv.EventMatchers {
				for _, em := range ems {
					fk := lsk + ":" + em.Field
					fieldMap[fk]++
				}
			}
		}
	}
	for k, c := range fieldMap {
		fmt.Printf("%s=%d\n", k, c)
	}
}

func runCLISigmaTest(rulesPath string, packs []string, configPath string, testLogs []string) {
	eng, err := initEngine(rulesPath, packs, configPath)
	if err != nil {
		log.Fatalf("failed to initialize sigma engine: %v", err)
	}
	hit := false
	now := time.Now().UnixNano()
	for _, l := range testLogs {
		if entry := eng.Match(l, now); entry != nil {
			e := entry.Evaluator
			lsk := fmt.Sprintf("%s:%s:%s", e.Logsource.Product, e.Logsource.Category, e.Logsource.Service)
			tags := strings.Join(e.Rule.Tags, ", ")
			fmt.Printf("===\nHIT: %s\nRule ID : %s\nLevel   : %s\nSource  : %s\nLogSrc  : %s\nTags    : %s\nTitle   : %s\n",
				l, e.ID, e.Level, entry.Source, lsk, tags, e.Title)
			hit = true
		}
	}
	if !hit {
		fmt.Println("===\nno rule match")
	}
}

func runCLISigmaConvertWazuh(inputPath string, outputDir string) {
	var xmlFiles []string
	info, err := os.Stat(inputPath)
	if err != nil {
		log.Fatalf("cannot access input path: %v", err)
	}
	if info.IsDir() {
		_ = filepath.WalkDir(inputPath, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(strings.ToLower(p), ".xml") {
				xmlFiles = append(xmlFiles, p)
			}
			return nil
		})
	} else {
		xmlFiles = append(xmlFiles, inputPath)
	}

	var allRules []sigma.WazuhRuleXML
	for _, f := range xmlFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		rules, err := sigma.ParseWazuhRulesXML(data)
		if err == nil {
			allRules = append(allRules, rules...)
		}
	}
	fmt.Printf("Parsed %d raw rules from %d XML file(s)\n", len(allRules), len(xmlFiles))

	resolved := sigma.ResolveRuleHierarchy(allRules, "")
	opts := sigma.WazuhConvertOptions{
		MinLevel: 3,
	}

	_ = os.MkdirAll(outputDir, 0755)
	count := 0
	for _, r := range resolved {
		rule, err := sigma.ConvertWazuhRuleToSigma(r, opts)
		if err == nil && rule != nil {
			out, err := yaml.Marshal(rule)
			if err == nil {
				filename := filepath.Join(outputDir, fmt.Sprintf("wazuh-%s.yaml", r.Rule.ID))
				_ = os.WriteFile(filename, out, 0644)
				count++
			}
		}
	}
	fmt.Printf("Converted %d rules to %s\n", count, outputDir)
}

func runCLISigmaConvertWazuhDecoder(inputPath string) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("failed to read %s: %v", inputPath, err)
	}
	decoders, err := sigma.ParseWazuhDecodersXML(data)
	if err != nil {
		log.Fatalf("failed to parse decoders: %v", err)
	}
	for _, dec := range decoders {
		pattern, err := sigma.ConvertDecoderToNamedRegex(dec)
		if err == nil && pattern != "" {
			fmt.Printf("# Decoder: %s\n%s\n\n", dec.Name, pattern)
		}
	}
}

func runCLISigmaScan(rulesPath string, packs []string, configPath, dsPath string) {
	if dsPath == "" {
		// Auto detect in current directory
		candidates := []string{"logs.parquet", "twlogaian.badger", "twlogaian_logs.db"}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				dsPath = c
				break
			}
		}
	}
	if dsPath == "" {
		log.Fatalln("datastore not found. Please specify -d <path>")
	}

	ds, err := datastore.Open(dsPath)
	if err != nil {
		log.Fatalf("failed to open datastore: %v", err)
	}
	defer ds.Close()

	eng, err := initEngine(rulesPath, packs, configPath)
	if err != nil {
		log.Fatalf("failed to initialize sigma engine: %v", err)
	}

	fmt.Printf("Scanning logs in %s using %d rules...\n", dsPath, len(eng.RuleEntries))
	st := time.Now()
	total := 0
	hits := 0
	ruleHits := make(map[string]int)

	_ = ds.ForEach(0, time.Now().UnixNano(), func(entry *datastore.LogEntry) bool {
		total++
		matched := eng.Match(entry.Log, entry.Time)
		if matched != nil {
			hits++
			ruleHits[matched.Evaluator.Title]++
		}
		return true
	})

	dur := time.Since(st)
	fmt.Printf("\nScan Complete in %v\n", dur)
	fmt.Printf("Total Logs : %d\n", total)
	fmt.Printf("Threat Hits: %d\n", hits)
	if hits > 0 {
		fmt.Println("\nTop Detected Rules:")
		type kv struct {
			k string
			v int
		}
		var list []kv
		for k, v := range ruleHits {
			list = append(list, kv{k, v})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].v > list[j].v })
		for i := 0; i < len(list) && i < 15; i++ {
			fmt.Printf("  [%4d] %s\n", list[i].v, list[i].k)
		}
	}
}
