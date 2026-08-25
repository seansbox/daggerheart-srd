package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

func main() {
	jsonDir := ".build/03_json"
	templateDir := ".build/04_md/templates"
	outputDir := ".build/04_md/docs"
	srdBasePath := ".build/01_pdf/DH-SRD-2.0-2026-08-25.md"
	srdPath := "README.md"

	entries, err := os.ReadDir(jsonDir)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", jsonDir, err)
		return
	}

	funcs := template.FuncMap{
		"upper":            strings.ToUpper,
		"urlEncode":        url.PathEscape,
		"featureQuestions": featureQuestions,
		"fileName":         sanitizeFilename,
		"abilityLink":      abilityLink,
		"optionAt":         optionAt,
		"add1":             add1,
	}

	var beastforms []map[string]any
	var beastformTiers []map[string]any
	if bf, err := loadJSON(filepath.Join(jsonDir, "beastforms.json")); err == nil {
		beastforms = bf
		beastformTiers = groupBeastformsByTier(beastforms)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		base := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		templatePath := filepath.Join(templateDir, base+".md")
		if _, err := os.Stat(templatePath); err != nil {
			continue
		}
		jsonPath := filepath.Join(jsonDir, entry.Name())
		items, err := loadJSON(jsonPath)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", jsonPath, err)
			continue
		}
		templateName := filepath.Base(templatePath)
		tmpl, err := template.New(templateName).Funcs(funcs).ParseFiles(templatePath)
		if err != nil {
			fmt.Printf("Error parsing template %s: %v\n", templatePath, err)
			continue
		}

		categoryDir := filepath.Join(outputDir, base)
		if err := os.MkdirAll(categoryDir, 0755); err != nil {
			fmt.Printf("Error creating %s: %v\n", categoryDir, err)
			continue
		}

		for _, item := range items {
			name, _ := item["name"].(string)
			if strings.TrimSpace(name) == "" {
				continue
			}
			if base == "classes" && len(beastformTiers) > 0 {
				item["beastform_tiers"] = beastformTiers
			}
			normalizeItem(item)
			outPath := filepath.Join(categoryDir, sanitizeFilename(name)+".md")
			if err := renderTemplate(tmpl, templateName, outPath, item); err != nil {
				fmt.Printf("Error rendering %s (%s): %v\n", outPath, name, err)
				continue
			}
		}
	}

	if err := generateSRD(srdBasePath, srdPath, jsonDir); err != nil {
		fmt.Printf("Error generating %s: %v\n", srdPath, err)
	}
}

func loadJSON(path string) ([]map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) >= 3 && raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF {
		raw = raw[3:]
	}
	var out []map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func renderTemplate(tmpl *template.Template, name, outPath string, data map[string]any) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := tmpl.ExecuteTemplate(f, name, data); err != nil {
		return err
	}
	if err := collapseExtraBlankLines(outPath); err != nil {
		return err
	}
	return ensureCanonicalFilename(outPath)
}

func normalizeItem(item map[string]any) {
	ensureSlice(item, "feature")
	ensureSlice(item, "background")
	ensureSlice(item, "connection")
	ensureSlice(item, "question")
	ensureSlice(item, "foundation")
	ensureSlice(item, "specialization")
	ensureSlice(item, "mastery")
	ensureString(item, "suggested_secondary")
	stripEmbeddedFeatureQuestions(item)
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "’", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, ":", "")
	name = strings.ReplaceAll(name, "&", "and")
	return name
}

func ensureSlice(item map[string]any, key string) {
	if _, ok := item[key]; !ok {
		item[key] = []any{}
	}
}

func ensureString(item map[string]any, key string) {
	if _, ok := item[key]; !ok {
		item[key] = ""
	}
}

func stripEmbeddedFeatureQuestions(item map[string]any) {
	features, ok := item["feature"].([]any)
	if !ok {
		return
	}
	for _, entry := range features {
		f, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		text, _ := f["text"].(string)
		question, _ := f["question"].(string)
		if text == "" || question == "" {
			continue
		}
		if strings.Contains(text, question) {
			f["question"] = ""
		}
	}
}

func collapseExtraBlankLines(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := normalizeMarkdown(strings.ReplaceAll(string(raw), "\r\n", "\n"))
	return os.WriteFile(path, []byte(out), 0644)
}

type linkTarget struct {
	name     string
	path     string
	category string
}

func generateSRD(basePath, outPath, jsonDir string) error {
	baseRaw, err := os.ReadFile(basePath)
	if err != nil {
		return err
	}
	content := strings.ReplaceAll(string(baseRaw), "\r\n", "\n")

	nameLinks, sectionLinks, categoryLinks, err := loadSRDLinkMaps(jsonDir)
	if err != nil {
		return err
	}

	content = replaceTableNamesWithLinks(content, sectionLinks)
	content = replaceListItemsWithLinks(content, categoryLinks)
	content = replaceHeadingBlocksWithLinks(content, nameLinks)
	content = replaceClassDomainLinks(content, categoryLinks["classes"], categoryLinks["domains"])
	content = replaceClassMentionList(content, categoryLinks["classes"])
	content = replaceAncestryMentionList(content, categoryLinks["ancestries"])
	content = insertCommunityParagraph(content, categoryLinks["communities"])
	content = insertDomainCardReferenceList(content, categoryLinks["domains"])
	content = removeDomainCardReferenceSections(content, categoryLinks["domains"])
	content = removeSectionsByHeadingPrefix(content, 2, []string{
		"TIER 1 ENVIRONMENTS",
		"TIER 2 ENVIRONMENTS",
		"TIER 3 ENVIRONMENTS",
		"TIER 4 ENVIRONMENTS",
		"TIER 1 ADVERSARIES",
		"TIER 2 ADVERSARIES",
		"TIER 3 ADVERSARIES",
		"TIER 4 ADVERSARIES",
	})
	content = normalizeMarkdown(content)
	content = insertContents(content)
	content = demoteHeadings(content)
	return os.WriteFile(outPath, []byte(content), 0644)
}

func loadSRDLinkMaps(jsonDir string) (map[string]linkTarget, map[string]map[string]linkTarget, map[string]map[string]linkTarget, error) {
	nameLinks := map[string]linkTarget{}
	sectionLinks := map[string]map[string]linkTarget{}
	categoryLinks := map[string]map[string]linkTarget{}
	sectionMap := map[string]string{
		"WEAPONS":     "weapons",
		"ARMOR":       "armor",
		"CONSUMABLES": "consumables",
		"ITEMS":       "items",
		"LOOT":        "items",
	}

	entries, err := os.ReadDir(jsonDir)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		base := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		items, err := loadJSON(filepath.Join(jsonDir, entry.Name()))
		if err != nil {
			return nil, nil, nil, err
		}
		for _, item := range items {
			name, _ := item["name"].(string)
			if strings.TrimSpace(name) == "" {
				continue
			}
			link := buildSRDLink(base, name)
			key := normalizeHeading(name)
			if _, exists := nameLinks[key]; !exists {
				nameLinks[key] = linkTarget{name: name, path: link, category: base}
			}
			for section, sectionBase := range sectionMap {
				if base == sectionBase {
					if sectionLinks[section] == nil {
						sectionLinks[section] = map[string]linkTarget{}
					}
					sectionLinks[section][key] = linkTarget{name: name, path: link, category: base}
				}
			}
			if categoryLinks[base] == nil {
				categoryLinks[base] = map[string]linkTarget{}
			}
			categoryLinks[base][key] = linkTarget{name: name, path: link, category: base}
		}
	}
	return nameLinks, sectionLinks, categoryLinks, nil
}

func buildSRDLink(category, name string) string {
	file := url.PathEscape(sanitizeFilename(name))
	return fmt.Sprintf("%s/%s.md", category, file)
}

func replaceHeadingBlocksWithLinks(content string, nameLinks map[string]linkTarget) string {
	lines := strings.Split(content, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		level, title := parseHeading(line)
		if level == 0 {
			out = append(out, line)
			continue
		}
		key := normalizeHeading(title)
		target, ok := nameLinks[key]
		if !ok {
			out = append(out, line)
			continue
		}
		if target.category == "adversaries" || target.category == "environments" {
			i = skipHeadingBlock(lines, i, level)
			continue
		}
		if target.category == "classes" {
			i = skipHeadingBlock(lines, i, level)
			continue
		}
		if target.category == "ancestries" {
			i = skipHeadingBlock(lines, i, level)
			continue
		}
		if target.category == "communities" {
			i = skipHeadingBlock(lines, i, level)
			continue
		}
		if target.category == "domains" {
			out = append(out, fmt.Sprintf("- [%s](%s)", target.name, target.path))
			out = append(out, "")
			i = skipHeadingBlock(lines, i, level)
			continue
		}
		out = append(out, line)
		out = append(out, "")
		out = append(out, fmt.Sprintf("See: [%s](%s)", target.name, target.path))
		out = append(out, "")
		i = skipHeadingBlock(lines, i, level)
	}
	return strings.Join(out, "\n")
}

func replaceTableNamesWithLinks(content string, sectionLinks map[string]map[string]linkTarget) string {
	lines := strings.Split(content, "\n")
	currentSection := ""
	for i, line := range lines {
		level, title := parseHeading(line)
		if level == 2 {
			currentSection = strings.ToUpper(strings.TrimSpace(title))
		}
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		links := sectionLinks[currentSection]
		if len(links) == 0 {
			continue
		}
		columnIndex := 1
		if currentSection == "LOOT" || currentSection == "CONSUMABLES" {
			columnIndex = 2
		}
		updated, changed := replaceTableCell(line, links, columnIndex)
		if changed {
			lines[i] = updated
		}
	}
	return strings.Join(lines, "\n")
}

func replaceListItemsWithLinks(content string, categoryLinks map[string]map[string]linkTarget) string {
	lines := strings.Split(content, "\n")
	inAdversaryList := false
	inEnvironmentSection := false
	for i, line := range lines {
		level, title := parseHeading(line)
		if level > 0 {
			if level == 3 && strings.EqualFold(strings.TrimSpace(title), "ADVERSARIES BY TIER") {
				inAdversaryList = true
			} else if level <= 3 {
				inAdversaryList = false
			}
			if level == 2 {
				inEnvironmentSection = strings.EqualFold(strings.TrimSpace(title), "USING ENVIRONMENTS")
			}
		}
		trimmed := strings.TrimLeft(line, " \t")
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		itemText := strings.TrimSpace(trimmed[2:])
		var updated string
		var changed bool
		if inAdversaryList {
			updated, changed = linkifyListItem(itemText, categoryLinks["adversaries"], false)
		} else if inEnvironmentSection {
			updated, changed = linkifyListItem(itemText, categoryLinks["environments"], true)
		}
		if changed {
			lines[i] = strings.Replace(line, trimmed, "- "+updated, 1)
		}
	}
	return strings.Join(lines, "\n")
}

func replaceClassDomainLinks(content string, classLinks, domainLinks map[string]linkTarget) string {
	if len(classLinks) == 0 || len(domainLinks) == 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	inClassDomains := false
	for i, line := range lines {
		level, title := parseHeading(line)
		if level == 2 {
			inClassDomains = strings.EqualFold(strings.TrimSpace(title), "CLASS DOMAINS")
		}
		if !inClassDomains {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- **") || !strings.Contains(trimmed, ":") {
			continue
		}
		start := strings.Index(trimmed, "**")
		if start == -1 {
			continue
		}
		rest := trimmed[start+2:]
		end := strings.Index(rest, "**")
		if end == -1 {
			continue
		}
		className := strings.TrimSuffix(rest[:end], ":")
		after := strings.TrimSpace(rest[end+2:])
		domainPart := strings.TrimSpace(strings.TrimPrefix(after, ":"))
		if domainPart == "" {
			continue
		}
		linkedDomains := linkDomainList(domainPart, domainLinks)
		linkedClass := className
		if target, ok := classLinks[normalizeHeading(className)]; ok {
			linkedClass = fmt.Sprintf("[%s](%s)", target.name, target.path)
		}
		lines[i] = fmt.Sprintf("- **%s:** %s", linkedClass, linkedDomains)
	}
	return strings.Join(lines, "\n")
}

func replaceClassMentionList(content string, classLinks map[string]linkTarget) string {
	if len(classLinks) == 0 {
		return content
	}
	prefix := "There are 9 classes in the Daggerheart core materials:"
	start := strings.Index(content, prefix)
	if start == -1 {
		return content
	}
	end := strings.Index(content[start:], ".")
	if end == -1 {
		return content
	}
	end += start
	listText := content[start+len(prefix) : end]
	parts := strings.Split(listText, ",")
	if len(parts) < 2 {
		return content
	}
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	lastPart := parts[len(parts)-1]
	if strings.HasPrefix(lastPart, "and ") {
		lastPart = strings.TrimSpace(strings.TrimPrefix(lastPart, "and "))
		parts[len(parts)-1] = lastPart
	}
	for i, name := range parts {
		target, ok := classLinks[normalizeHeading(name)]
		if !ok {
			continue
		}
		parts[i] = fmt.Sprintf("[%s](%s)", target.name, target.path)
	}
	if len(parts) > 1 {
		parts[len(parts)-1] = "and " + parts[len(parts)-1]
	}
	repl := prefix + " " + strings.Join(parts, ", ") + "."
	content = content[:start] + repl + content[end+1:]
	return content
}

func replaceAncestryMentionList(content string, ancestryLinks map[string]linkTarget) string {
	if len(ancestryLinks) == 0 {
		return content
	}
	prefix := "The core ruleset includes the following ancestries:"
	start := strings.Index(content, prefix)
	if start == -1 {
		return content
	}
	end := strings.Index(content[start:], ".")
	if end == -1 {
		return content
	}
	end += start
	listText := content[start+len(prefix) : end]
	parts := strings.Split(listText, ",")
	if len(parts) < 2 {
		return content
	}
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	lastPart := parts[len(parts)-1]
	if strings.HasPrefix(lastPart, "and ") {
		lastPart = strings.TrimSpace(strings.TrimPrefix(lastPart, "and "))
		parts[len(parts)-1] = lastPart
	}
	for i, name := range parts {
		target, ok := ancestryLinks[normalizeHeading(name)]
		if !ok {
			continue
		}
		parts[i] = fmt.Sprintf("[%s](%s)", target.name, target.path)
	}
	if len(parts) > 1 {
		parts[len(parts)-1] = "and " + parts[len(parts)-1]
	}
	repl := prefix + " " + strings.Join(parts, ", ") + "."
	content = content[:start] + repl + content[end+1:]
	return content
}

func insertCommunityParagraph(content string, communityLinks map[string]linkTarget) string {
	if len(communityLinks) == 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		level, title := parseHeading(line)
		if level == 2 && strings.EqualFold(strings.TrimSpace(title), "COMMUNITIES") {
			end := i + 1
			for end < len(lines) {
				nextLevel, _ := parseHeading(lines[end])
				if nextLevel > 0 && nextLevel <= 2 {
					break
				}
				end++
			}
			var names []string
			for _, target := range communityLinks {
				names = append(names, target.name)
			}
			sort.Strings(names)
			var linked []string
			for _, name := range names {
				target := communityLinks[normalizeHeading(name)]
				linked = append(linked, fmt.Sprintf("[%s](%s)", target.name, target.path))
			}
			paragraph := "The core ruleset includes the following communities: " + strings.Join(linked, ", ") + "."
			block := []string{"", paragraph, ""}
			out := append([]string{}, lines[:end]...)
			out = append(out, block...)
			out = append(out, lines[end:]...)
			return strings.Join(out, "\n")
		}
	}
	return content
}

func insertDomainCardReferenceList(content string, domainLinks map[string]linkTarget) string {
	if len(domainLinks) == 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		level, title := parseHeading(line)
		if level == 2 && strings.EqualFold(strings.TrimSpace(title), "DOMAIN CARD REFERENCE") {
			insertAt := i + 1
			var names []string
			for _, target := range domainLinks {
				names = append(names, target.name)
			}
			sort.Strings(names)
			var block []string
			block = append(block, "")
			for _, name := range names {
				target := domainLinks[normalizeHeading(name)]
				block = append(block, fmt.Sprintf("- [%s](%s)", target.name, target.path))
			}
			block = append(block, "")
			out := append([]string{}, lines[:insertAt]...)
			out = append(out, block...)
			out = append(out, lines[insertAt:]...)
			return strings.Join(out, "\n")
		}
	}
	return content
}

func removeDomainCardReferenceSections(content string, domainLinks map[string]linkTarget) string {
	if len(domainLinks) == 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		level, title := parseHeading(lines[i])
		if level == 3 && strings.HasSuffix(strings.ToUpper(strings.TrimSpace(title)), " DOMAIN") {
			name := strings.TrimSpace(strings.TrimSuffix(title, "DOMAIN"))
			name = strings.TrimSpace(strings.TrimSuffix(name, "Domain"))
			if _, ok := domainLinks[normalizeHeading(name)]; ok {
				i = skipHeadingBlock(lines, i, level)
				continue
			}
		}
		out = append(out, lines[i])
	}
	return strings.Join(out, "\n")
}

func linkDomainList(value string, domainLinks map[string]linkTarget) string {
	parts := strings.Split(value, "&")
	if len(parts) == 1 {
		return linkSingleDomain(value, domainLinks)
	}
	for i, part := range parts {
		parts[i] = linkSingleDomain(part, domainLinks)
	}
	return strings.Join(parts, " & ")
}

func linkSingleDomain(value string, domainLinks map[string]linkTarget) string {
	name := strings.TrimSpace(value)
	if name == "" {
		return value
	}
	if target, ok := domainLinks[normalizeHeading(name)]; ok {
		return fmt.Sprintf("[%s](%s)", target.name, target.path)
	}
	return name
}

func linkifyListItem(item string, links map[string]linkTarget, allowSuffix bool) (string, bool) {
	if len(links) == 0 {
		return item, false
	}
	parts := strings.Split(item, "•")
	changed := false
	for i, part := range parts {
		part = strings.TrimSpace(part)
		base := part
		suffix := ""
		if allowSuffix {
			if idx := strings.Index(part, " ("); idx != -1 {
				base = strings.TrimSpace(part[:idx])
				suffix = part[idx:]
			}
		}
		key := normalizeHeading(base)
		target, ok := links[key]
		if ok {
			parts[i] = fmt.Sprintf("[%s](%s)%s", target.name, target.path, suffix)
			changed = true
		} else {
			parts[i] = part
		}
	}
	return strings.Join(parts, " • "), changed
}

func replaceTableCell(line string, links map[string]linkTarget, columnIndex int) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return line, false
	}
	parts := strings.Split(line, "|")
	if len(parts) <= columnIndex {
		return line, false
	}
	if isTableSeparatorRow(parts) {
		return line, false
	}
	cell := strings.TrimSpace(parts[columnIndex])
	cellPlain := stripMarkdownEmphasis(cell)
	if strings.EqualFold(cellPlain, "name") || strings.EqualFold(cellPlain, "roll") || strings.EqualFold(cellPlain, "loot") {
		return line, false
	}
	key := normalizeHeading(cell)
	target, ok := links[key]
	if !ok {
		return line, false
	}
	parts[columnIndex] = " [" + target.name + "](" + target.path + ") "
	return strings.Join(parts, "|"), true
}

func isTableSeparatorRow(parts []string) bool {
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		for _, r := range trimmed {
			if r != '-' {
				return false
			}
		}
		return true
	}
	return false
}

func stripMarkdownEmphasis(value string) string {
	out := strings.ReplaceAll(value, "**", "")
	out = strings.ReplaceAll(out, "__", "")
	out = strings.ReplaceAll(out, "*", "")
	out = strings.ReplaceAll(out, "_", "")
	return strings.TrimSpace(out)
}

func abilityLink(name string) string {
	clean := strings.TrimSpace(name)
	if clean == "" || clean == "—" {
		return "—"
	}
	return fmt.Sprintf("[%s](../abilities/%s.md)", clean, url.PathEscape(sanitizeFilename(clean)))
}

func optionAt(options any, index int) string {
	if index <= 0 {
		return ""
	}
	switch typed := options.(type) {
	case []any:
		if len(typed) < index {
			return ""
		}
		if val, ok := typed[index-1].(string); ok {
			return val
		}
	case []string:
		if len(typed) < index {
			return ""
		}
		return typed[index-1]
	}
	return ""
}

func add1(value int) int {
	return value + 1
}

func parseHeading(line string) (int, string) {
	trimmed := strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(trimmed, "#") {
		return 0, ""
	}
	hashes := 0
	for hashes < len(trimmed) && trimmed[hashes] == '#' {
		hashes++
	}
	if hashes == 0 || hashes >= len(trimmed) || trimmed[hashes] != ' ' {
		return 0, ""
	}
	return hashes, strings.TrimSpace(trimmed[hashes:])
}

func skipHeadingBlock(lines []string, start, level int) int {
	for i := start + 1; i < len(lines); i++ {
		nextLevel, _ := parseHeading(lines[i])
		if nextLevel > 0 && nextLevel <= level {
			return i - 1
		}
	}
	return len(lines) - 1
}

func normalizeHeading(text string) string {
	clean := strings.ToLower(strings.TrimSpace(text))
	clean = strings.ReplaceAll(clean, "’", "")
	clean = strings.ReplaceAll(clean, "'", "")
	var b strings.Builder
	space := false
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			space = false
			continue
		}
		if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

func removeSectionsByHeadingPrefix(content string, level int, prefixes []string) string {
	lines := strings.Split(content, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		lvl, title := parseHeading(lines[i])
		if lvl == level && hasHeadingPrefix(title, prefixes) {
			i = skipHeadingBlock(lines, i, lvl)
			continue
		}
		out = append(out, lines[i])
	}
	return strings.Join(out, "\n")
}

func hasHeadingPrefix(title string, prefixes []string) bool {
	upper := strings.ToUpper(strings.TrimSpace(title))
	for _, prefix := range prefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

func rtrimLines(input string) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

func removeBlankLinesBetweenListItems(input string) string {
	lines := strings.Split(input, "\n")
	out := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" && i > 0 && i+1 < len(lines) {
			if isListItem(lines[i-1]) && isListItem(lines[i+1]) {
				continue
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func isListItem(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(trimmed, "- ") {
		return true
	}
	dot := strings.IndexByte(trimmed, '.')
	if dot <= 0 {
		return false
	}
	for i := 0; i < dot; i++ {
		if trimmed[i] < '0' || trimmed[i] > '9' {
			return false
		}
	}
	return len(trimmed) > dot+1 && trimmed[dot+1] == ' '
}

func normalizeMarkdown(input string) string {
	out := rtrimLines(input)
	for strings.Contains(out, "\n\n\n") {
		out = strings.ReplaceAll(out, "\n\n\n", "\n\n")
	}
	out = removeBlankLinesBetweenListItems(out)
	return out
}

func ensureCanonicalFilename(path string) error {
	dir := filepath.Dir(path)
	want := filepath.Base(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.EqualFold(name, want) || name == want {
			continue
		}
		tmp := want + ".casefix"
		oldPath := filepath.Join(dir, name)
		tmpPath := filepath.Join(dir, tmp)
		if err := os.Rename(oldPath, tmpPath); err != nil {
			return err
		}
		return os.Rename(tmpPath, path)
	}
	return nil
}

func insertContents(content string) string {
	lines := strings.Split(content, "\n")
	var toc []string
	seenFirst := false
	for _, line := range lines {
		level, title := parseHeading(line)
		if level == 0 || level > 2 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(title), "contents") {
			continue
		}
		if level == 1 && !seenFirst {
			seenFirst = true
			continue
		}
		anchor := headingAnchor(title)
		if anchor == "" {
			continue
		}
		display := titleCaseHeading(title)
		if level == 1 {
			if len(toc) > 0 && toc[len(toc)-1] != "" {
				toc = append(toc, "")
			}
			toc = append(toc, fmt.Sprintf("**[%s](#%s)**", display, anchor), "")
			continue
		}
		toc = append(toc, fmt.Sprintf("- [%s](#%s)", display, anchor))
	}
	if len(toc) == 0 {
		return content
	}
	for i, line := range lines {
		level, title := parseHeading(line)
		if level == 1 && strings.EqualFold(strings.TrimSpace(title), "INTRODUCTION") {
			block := []string{
				"**REFERENCE SITE**",
				"",
				"[seansbox.github.io/daggerheart-srd](https://seansbox.github.io/daggerheart-srd/)",
				"",
				"# CONTENTS",
				"",
			}
			block = append(block, toc...)
			block = append(block, "")
			out := append([]string{}, lines[:i]...)
			out = append(out, block...)
			out = append(out, lines[i:]...)
			return strings.Join(out, "\n")
		}
	}
	block := append([]string{
		"#### REFERENCE SITE",
		"",
		"[seansbox.github.io/daggerheart-srd](https://seansbox.github.io/daggerheart-srd/)",
		"",
		"# CONTENTS",
		"",
	}, toc...)
	block = append(block, "", "")
	return strings.Join(append(block, lines...), "\n")
}

func demoteHeadings(content string) string {
	lines := strings.Split(content, "\n")
	first := true
	for i, line := range lines {
		level, title := parseHeading(line)
		if level == 0 {
			continue
		}
		if first {
			first = false
			continue
		}
		level++
		if level > 6 {
			level = 6
		}
		lines[i] = strings.Repeat("#", level) + " " + strings.TrimSpace(title)
	}
	return strings.Join(lines, "\n")
}

func headingAnchor(title string) string {
	clean := strings.ToLower(strings.TrimSpace(title))
	var b strings.Builder
	dash := false
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	anchor := strings.Trim(b.String(), "-")
	return anchor
}

func titleCaseHeading(title string) string {
	smallWords := map[string]bool{
		"a": true, "an": true, "and": true, "as": true, "at": true,
		"but": true, "by": true, "for": true, "from": true, "in": true,
		"of": true, "on": true, "or": true, "the": true, "to": true,
		"via": true, "with": true, "over": true, "into": true,
	}
	acronyms := map[string]bool{
		"GM": true, "SRD": true, "HP": true, "XP": true, "PC": true, "NPC": true, "ATK": true,
	}
	parts := strings.Fields(title)
	for i, part := range parts {
		upper := strings.ToUpper(part)
		lower := strings.ToLower(part)
		if i > 0 && smallWords[lower] {
			parts[i] = lower
			continue
		}
		if part == upper && acronyms[part] {
			parts[i] = part
			continue
		}
		parts[i] = titleizeWord(lower)
	}
	return strings.Join(parts, " ")
}

func titleizeWord(word string) string {
	if word == "" {
		return word
	}
	segments := strings.Split(word, "-")
	for i, seg := range segments {
		if seg == "" {
			continue
		}
		runes := []rune(seg)
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		segments[i] = string(runes)
	}
	return strings.Join(segments, "-")
}

func groupBeastformsByTier(beastforms []map[string]any) []map[string]any {
	tierBuckets := map[int][]map[string]any{}
	for _, beast := range beastforms {
		tier := tierFromValue(beast["tier"])
		tierBuckets[tier] = append(tierBuckets[tier], beast)
	}
	ordered := []int{1, 2, 3, 4}
	seen := map[int]bool{}
	for _, t := range ordered {
		seen[t] = true
	}
	var tiers []map[string]any
	for _, t := range ordered {
		items := tierBuckets[t]
		if len(items) == 0 {
			continue
		}
		tiers = append(tiers, map[string]any{
			"tier":  t,
			"items": items,
		})
	}
	for t, items := range tierBuckets {
		if seen[t] || len(items) == 0 {
			continue
		}
		tiers = append(tiers, map[string]any{
			"tier":  t,
			"items": items,
		})
	}
	return tiers
}

func tierFromValue(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		var out int
		fmt.Sscanf(v, "%d", &out)
		if out > 0 {
			return out
		}
	}
	return 0
}

func featureQuestions(features any) []string {
	list, ok := features.([]any)
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, entry := range list {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		q, ok := m["question"].(string)
		if !ok {
			continue
		}
		q = strings.TrimSpace(q)
		if q == "" || seen[q] {
			continue
		}
		seen[q] = true
		out = append(out, q)
	}
	return out
}
