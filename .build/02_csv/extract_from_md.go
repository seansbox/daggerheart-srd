package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

func main() {
	markdownPath := ".build/01_pdf/DH-SRD-2.0-2026-08-25.md"
	contentBytes, err := os.ReadFile(markdownPath)
	if err != nil {
		fmt.Printf("Error reading markdown file: %v\n", err)
		return
	}
	content := string(contentBytes)

	outputDir := ".build/02_csv"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	scanVariations(content)

	extractAbilities(content, outputDir)
	extractAdversaries(content, outputDir)
	extractAncestries(content, outputDir)
	extractArmor(content, outputDir)
	extractClasses(content, outputDir)
	extractCommunities(content, outputDir)
	extractConsumables(content, outputDir)
	extractDomains(content, outputDir)
	extractBeastforms(content, outputDir)
	extractEnvironments(content, outputDir)
	extractItems(content, outputDir)
	extractSubclasses(content, outputDir)
	extractWeapons(content, outputDir)
}

func normalizeHeaderText(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isAllCaps(s string) bool {
	hasLetters := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return false
		}
		if r >= 'A' && r <= 'Z' {
			hasLetters = true
		}
	}
	return hasLetters
}

func titleizeAllCaps(s string) string {
	smallWords := map[string]bool{
		"a": true, "an": true, "and": true, "as": true, "at": true,
		"but": true, "by": true, "for": true, "from": true, "in": true,
		"of": true, "on": true, "or": true, "the": true, "to": true,
		"via": true, "with": true, "over": true, "into": true,
	}
	acronyms := map[string]bool{
		"GM": true, "SRD": true, "HP": true, "XP": true, "PC": true, "NPC": true, "ATK": true,
	}
	words := strings.Fields(s)
	for i, w := range words {
		upper := strings.ToUpper(w)
		lower := strings.ToLower(w)
		if i > 0 && smallWords[lower] {
			words[i] = lower
			continue
		}
		if w == upper && acronyms[w] {
			words[i] = w
			continue
		}
		segments := strings.Split(lower, "-")
		for j, seg := range segments {
			runes := []rune(seg)
			if len(runes) == 0 {
				continue
			}
			runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
			segments[j] = string(runes)
		}
		words[i] = strings.Join(segments, "-")
	}
	return strings.Join(words, " ")
}

func titleizeIfAllCaps(s string) string {
	if isAllCaps(s) {
		return titleizeAllCaps(s)
	}
	return s
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

func headerLevel(line string) int {
	level := 0
	trimmed := strings.TrimLeft(line, " ")
	for _, c := range trimmed {
		if c == '#' {
			level++
		} else {
			break
		}
	}
	return level
}

func findHeaderLine(content, headerLine string) (string, bool) {
	level := headerLevel(headerLine)
	target := normalizeHeaderText(strings.TrimSpace(strings.TrimLeft(headerLine, "#")))
	re := regexp.MustCompile(`(?m)^(#{1,6})\s*(\S.*)$`)
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		if len(m[1]) != level {
			continue
		}
		if normalizeHeaderText(m[2]) == target {
			return m[0], true
		}
	}
	return "", false
}

// getSection returns the content of a section starting with the given header line,
// until the next header of equal or higher level (fewer hashes).
func getSection(content, headerLine string) string {
	// Find exact start of the header line
	reStart := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(headerLine) + `\s*$`)
	loc := reStart.FindStringIndex(content)
	if loc == nil {
		if alt, ok := findHeaderLine(content, headerLine); ok {
			headerLine = alt
			reStart = regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(headerLine) + `\s*$`)
			loc = reStart.FindStringIndex(content)
		}
	}
	if loc == nil {
		return ""
	}
	startIdx := loc[0]

	// Determine header level
	level := headerLevel(headerLine)

	// End is next header with 1 to 'level' hashes
	reEnd := regexp.MustCompile(fmt.Sprintf(`(?m)^#{1,%d}\s`, level))

	// Start search after the current header line
	rest := content[startIdx:]
	firstLineEnd := strings.Index(rest, "\n")
	if firstLineEnd == -1 {
		return rest
	}

	searchBody := rest[firstLineEnd+1:]
	endLoc := reEnd.FindStringIndex(searchBody)
	if endLoc == nil {
		return rest
	}

	return rest[:firstLineEnd+1+endLoc[0]]
}

func warnMissing(entity, name, field string) {
	if name == "" {
		name = "<unknown>"
	}
	fmt.Printf("Warning: %s '%s' missing %s\n", entity, name, field)
}

func warnFormat(entity, name, detail string) {
	if name == "" {
		name = "<unknown>"
	}
	fmt.Printf("Notice: %s '%s' format: %s\n", entity, name, detail)
}

func parseMarkdownTableRow(line string, expectedCols int) ([]string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return nil, false
	}
	if strings.Contains(trimmed, "---") {
		return nil, false
	}
	trimmed = strings.Trim(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if expectedCols > 0 && len(parts) < expectedCols {
		return nil, false
	}
	return parts, true
}

func cleanInlineText(s string) string {
	return strings.TrimSpace(strings.Trim(s, "*_ "))
}

func findHeaderBlockStart(fullText string, re *regexp.Regexp) (int, string) {
	loc := re.FindStringIndex(fullText)
	if loc == nil {
		return -1, ""
	}
	header := strings.TrimSpace(fullText[loc[0]:loc[1]])
	return loc[0], header
}

func findNextHeaderIndex(content string, start int) int {
	if start < 0 || start >= len(content) {
		return -1
	}
	rest := content[start:]
	firstLineEnd := strings.Index(rest, "\n")
	if firstLineEnd == -1 {
		return -1
	}
	search := rest[firstLineEnd+1:]
	re := regexp.MustCompile(`(?m)^#{1,6}\s`)
	loc := re.FindStringIndex(search)
	if loc == nil {
		return -1
	}
	return start + firstLineEnd + 1 + loc[0]
}

func findHeaderIndex(content, headerLine string) int {
	reStart := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(headerLine) + `\s*$`)
	loc := reStart.FindStringIndex(content)
	if loc == nil {
		if alt, ok := findHeaderLine(content, headerLine); ok {
			reStart = regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(alt) + `\s*$`)
			loc = reStart.FindStringIndex(content)
		}
	}
	if loc == nil {
		return -1
	}
	return loc[0]
}

func extractFeaturePairs(block string) []string {
	lines := strings.Split(block, "\n")
	var feats []string
	currN, currT := "", ""
	reColon := regexp.MustCompile(`^_([^_]+):_\s*(.*)`)
	reHead := regexp.MustCompile(`^_(.+?)_\s*(.*)`)
	reDot := regexp.MustCompile(`^_(.+?)\._\s*(.*)`)
	for i, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "#") {
			if i == 0 {
				continue
			}
			break
		}
		if m := reColon.FindStringSubmatch(l); len(m) > 2 {
			if currN != "" {
				feats = append(feats, currN, currT)
			}
			currN = strings.TrimSpace(strings.Trim(m[1], "*_ "))
			currN = strings.TrimSuffix(currN, ":")
			currN = titleizeIfAllCaps(currN)
			currT = strings.TrimSpace(m[2])
			continue
		}
		if m := reDot.FindStringSubmatch(l); len(m) > 2 {
			if currN != "" {
				feats = append(feats, currN, currT)
			}
			currN = strings.TrimSpace(strings.Trim(m[1], "*_ "))
			currN = titleizeIfAllCaps(currN)
			currT = strings.TrimSpace(m[2])
			continue
		}
		if m := reHead.FindStringSubmatch(l); len(m) > 2 && strings.Contains(m[1], ":") {
			if currN != "" {
				feats = append(feats, currN, currT)
			}
			currN = strings.TrimSpace(strings.Trim(m[1], "*_ "))
			currN = strings.TrimSuffix(currN, ":")
			currN = titleizeIfAllCaps(currN)
			currT = strings.TrimSpace(m[2])
			continue
		}
		if strings.HasPrefix(l, "_") && strings.Contains(l, ":") {
			if currN != "" {
				feats = append(feats, currN, currT)
			}
			parts := strings.SplitN(strings.TrimPrefix(l, "_"), ":", 2)
			currN = strings.TrimSpace(strings.Trim(parts[0], "*_ "))
			currN = titleizeIfAllCaps(currN)
			currT = ""
			if len(parts) > 1 {
				currT = strings.TrimSpace(strings.Trim(parts[1], "*_ "))
			}
			continue
		}
		if currN != "" {
			if l == "" {
				if currT != "" {
					currT += "\n"
				}
				continue
			}
			if currT != "" {
				currT += "\n"
			}
			currT += l
		}
	}
	if currN != "" {
		feats = append(feats, currN, strings.Trim(currT, "\n"))
	}
	for i := 1; i < len(feats); i += 2 {
		feats[i] = strings.Trim(feats[i], "\n")
	}
	return feats
}

func scanVariations(content string) {
	type sectionRange struct {
		start int
		end   int
	}
	buildScanRanges := func(text string) []sectionRange {
		reH2 := regexp.MustCompile(`(?m)^##\s+(.+)$`)
		matches := reH2.FindAllStringSubmatchIndex(text, -1)
		type h2 struct {
			start int
			end   int
			title string
		}
		headers := make([]h2, 0, len(matches))
		for i, m := range matches {
			start := m[0]
			end := len(text)
			if i+1 < len(matches) {
				end = matches[i+1][0]
			}
			title := strings.TrimSpace(text[m[2]:m[3]])
			headers = append(headers, h2{start: start, end: end, title: title})
		}

		normalize := func(s string) string {
			return normalizeHeaderText(s)
		}

		ranges := []sectionRange{}

		// Classes span from ## CLASSES to ## ANCESTRIES (or end).
		classStart := -1
		classEnd := -1
		for i, h := range headers {
			if normalize(h.title) == "classes" {
				classStart = h.start
				for j := i + 1; j < len(headers); j++ {
					if normalize(headers[j].title) == "ancestries" {
						classEnd = headers[j].start
						break
					}
				}
				if classEnd == -1 {
					classEnd = h.end
				}
				break
			}
		}
		if classStart != -1 && classEnd != -1 {
			ranges = append(ranges, sectionRange{start: classStart, end: classEnd})
		}

		isDataSection := func(title string) bool {
			n := normalize(title)
			return strings.Contains(n, "adversar") ||
				strings.Contains(n, "ancestr") ||
				strings.Contains(n, "communit") ||
				strings.Contains(n, "environment") ||
				strings.Contains(n, "weapon") ||
				strings.Contains(n, "armortable") ||
				strings.Contains(n, "domaincardreference") ||
				strings.Contains(n, "domains") ||
				strings.Contains(n, "consumable") ||
				strings.Contains(n, "loot")
		}

		for _, h := range headers {
			if normalize(h.title) == "classes" {
				continue
			}
			if isDataSection(h.title) {
				ranges = append(ranges, sectionRange{start: h.start, end: h.end})
			}
		}

		if len(ranges) == 0 {
			return ranges
		}
		// Merge overlapping ranges.
		sort.Slice(ranges, func(i, j int) bool {
			return ranges[i].start < ranges[j].start
		})
		merged := []sectionRange{ranges[0]}
		for _, r := range ranges[1:] {
			last := &merged[len(merged)-1]
			if r.start <= last.end {
				if r.end > last.end {
					last.end = r.end
				}
				continue
			}
			merged = append(merged, r)
		}
		return merged
	}

	isInRanges := func(pos int, ranges []sectionRange) bool {
		for _, r := range ranges {
			if pos >= r.start && pos < r.end {
				return true
			}
		}
		return false
	}

	lines := strings.Split(content, "\n")
	ranges := buildScanRanges(content)
	ancestrySingular := []string{}
	adversaryTierNoDigit := []string{}
	weaponTierColon := []string{}
	featureHeadingVariants := []string{}

	reAdversaryTier := regexp.MustCompile(`[*][*]_Tier\s+([^_]+)_[*][*]`)
	reWeaponTier := regexp.MustCompile(`(?i)^####\s+TIER`)

	knownFeatureHeadings := map[string]bool{
		"##### FEATURES":                    true,
		"##### ANCESTRY FEATURES":           true,
		"##### ANCESTRY FEATURE":            true,
		"##### COMMUNITY FEATURE":           true,
		"##### CLASS FEATURE":               true,
		"##### CLASS FEATURES":              true,
		"##### FOUNDATION FEATURE":          true,
		"##### FOUNDATION FEATURES":         true,
		"##### SPECIALIZATION FEATURE":      true,
		"##### SPECIALIZATION FEATURES":     true,
		"##### MASTERY FEATURE":             true,
		"##### MASTERY FEATURES":            true,
		"##### SPELLCAST TRAIT":             true,
		"##### HOPE FEATURE":                true,
		"##### USING FEATURES AFTER A ROLL": true,
	}

	pos := 0
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		if len(ranges) > 0 && !isInRanges(pos, ranges) {
			pos += len(line) + 1
			continue
		}

		if strings.Contains(trimmed, "ANCESTRY FEATURE") && !strings.Contains(trimmed, "FEATURES") {
			ancestrySingular = append(ancestrySingular, fmt.Sprintf("%d: %s", lineNum, trimmed))
		}

		if m := reAdversaryTier.FindStringSubmatch(trimmed); len(m) > 1 {
			if !regexp.MustCompile(`\d`).MatchString(m[1]) {
				adversaryTierNoDigit = append(adversaryTierNoDigit, fmt.Sprintf("%d: %s", lineNum, trimmed))
			}
		}

		if reWeaponTier.MatchString(trimmed) && strings.Contains(trimmed, ":") {
			weaponTierColon = append(weaponTierColon, fmt.Sprintf("%d: %s", lineNum, trimmed))
		}

		if strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, "FEATURE") {
			isClassHope := strings.HasPrefix(trimmed, "##### ") && strings.HasSuffix(trimmed, "'S HOPE FEATURE")
			if !knownFeatureHeadings[trimmed] && !isClassHope {
				featureHeadingVariants = append(featureHeadingVariants, fmt.Sprintf("%d: %s", lineNum, trimmed))
				continue
			}
			if !strings.HasPrefix(trimmed, "#####") {
				featureHeadingVariants = append(featureHeadingVariants, fmt.Sprintf("%d: %s", lineNum, trimmed))
			}
		}
		pos += len(line) + 1
	}

	if len(ancestrySingular) > 0 {
		fmt.Println("Notice: Ancestry feature header singular variants:")
		for _, v := range ancestrySingular {
			fmt.Println(" ", v)
		}
	}
	if len(adversaryTierNoDigit) > 0 {
		fmt.Println("Notice: Adversary tier lines without digits:")
		for _, v := range adversaryTierNoDigit {
			fmt.Println(" ", v)
		}
	}
	if len(weaponTierColon) > 0 {
		fmt.Println("Notice: Weapon tier headers with colon variants:")
		for _, v := range weaponTierColon {
			fmt.Println(" ", v)
		}
	}
	if len(featureHeadingVariants) > 0 {
		fmt.Println("Notice: Feature heading variants:")
		for _, v := range featureHeadingVariants {
			fmt.Println(" ", v)
		}
	}
}

func writeCSV(filename string, header []string, records [][]string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating CSV file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		fmt.Printf("Error writing header to %s: %v\n", filename, err)
		return
	}

	if err := writer.WriteAll(records); err != nil {
		fmt.Printf("Error writing records to %s: %v\n", filename, err)
		return
	}
	fmt.Printf("Generated %s with %d records\n", filename, len(records))
}

func extractAbilities(content, outputDir string) {
	section := getSection(content, "## DOMAIN CARD REFERENCE")
	if section == "" {
		fmt.Println("Warning: Domain Card Reference section not found")
		return
	}

	// Ability blocks start with #####
	chunks := strings.Split(section, "\n##### ")
	var records [][]string

	reLevel := regexp.MustCompile(`[*][*]Level\s+(\d+)\s+(.*?)\s+(Ability|Grimoire|Spell)[*][*]`)
	reRecall := regexp.MustCompile(`[*][*]Recall Cost:\s+(\d+)[*][*]`)

	for i, chunk := range chunks {
		if i == 0 {
			continue
		} // Intro text
		lines := strings.Split(chunk, "\n")
		name := titleizeIfAllCaps(strings.TrimSpace(lines[0]))

		fullText := strings.Join(lines, "\n")

		level, domain, typeStr, recall, text := "", "", "", "", ""

		if m := reLevel.FindStringSubmatch(fullText); len(m) > 3 {
			level = m[1]
			domain = strings.TrimSpace(m[2])
			typeStr = m[3]
		}
		if m := reRecall.FindStringSubmatch(fullText); len(m) > 1 {
			recall = m[1]
		}

		// Find where descriptive text begins (after stats)
		statsEnd := 0
		if loc := reRecall.FindStringIndex(fullText); loc != nil {
			statsEnd = loc[1]
		} else if loc := reLevel.FindStringIndex(fullText); loc != nil {
			statsEnd = loc[1]
		}

		if statsEnd > 0 && statsEnd < len(fullText) {
			text = strings.TrimSpace(fullText[statsEnd:])
		} else {
			var textLines []string
			for _, l := range lines[1:] {
				if !strings.HasPrefix(strings.TrimSpace(l), "**") {
					textLines = append(textLines, l)
				}
			}
			text = strings.TrimSpace(strings.Join(textLines, "\n"))
		}

		// Ensure we don't leak into next major header if split was loose
		if idx := strings.Index(text, "\n#"); idx != -1 {
			text = strings.TrimSpace(text[:idx])
		}

		records = append(records, []string{name, level, domain, typeStr, recall, text})
	}

	writeCSV(outputDir+"/abilities.csv", []string{"Name", "Level", "Domain", "Type", "Recall", "Text"}, records)
}

func extractAdversaries(content, outputDir string) {
	// Adversaries are in sections like "## TIER 1 ADVERSARIES" or similar
	// We find all L2 headers that contain "ADVERSARIES" but not "USING" or "ENVIRONMENT"
	reL2 := regexp.MustCompile(`(?m)^## (.*ADVERSARIES.*)`)
	matches := reL2.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		fmt.Println("Warning: No Adversary sections found")
		return
	}

	var records [][]string
	reSectionTier := regexp.MustCompile(`(?i)Tier\s+(\d+)`)
	for _, m := range matches {
		h := m[1]
		if strings.Contains(h, "USING") || strings.Contains(h, "ENVIRONMENT") {
			continue
		}
		section := getSection(content, "## "+h)
		if section == "" {
			continue
		}
		defaultTier := ""
		if tm := reSectionTier.FindStringSubmatch(h); len(tm) > 1 {
			defaultTier = tm[1]
		}

		chunks := strings.Split(section, "\n### ")
		for i, chunk := range chunks {
			if i == 0 {
				continue
			}
			lines := strings.Split(chunk, "\n")
			name := titleizeIfAllCaps(strings.TrimSpace(lines[0]))
			fullText := strings.Join(lines, "\n")

			// Adversaries must have a Tier block: **_Tier X Type_**
			reTierFull := regexp.MustCompile(`[*][*]_Tier\s+(\d+)\s+([^_]+?)_[*][*]`)
			reTierType := regexp.MustCompile(`[*][*]_Tier\s+([A-Za-z]+)_[*][*]`)
			tier, typeStr := defaultTier, ""
			if tierMatch := reTierFull.FindStringSubmatch(fullText); len(tierMatch) > 2 {
				tier = strings.TrimSpace(tierMatch[1])
				typeStr = strings.TrimSpace(tierMatch[2])
			} else if tierMatch := reTierType.FindStringSubmatch(fullText); len(tierMatch) > 1 {
				typeStr = strings.TrimSpace(tierMatch[1])
			}
			typeStr = titleizeIfAllCaps(typeStr)
			if tier == "" {
				warnMissing("Adversary", name, "Tier")
			}
			if typeStr == "" {
				warnMissing("Adversary", name, "Type")
			}

			desc := ""
			reDesc := regexp.MustCompile(`[*][*]_.*?_[*][*]\n_(.*?)_`)
			if m := reDesc.FindStringSubmatch(fullText); len(m) > 1 {
				desc = strings.TrimSpace(m[1])
			}
			if desc == "" {
				warnMissing("Adversary", name, "Description")
			}

			motives, diff, thresh, hp, stress, atk, attackName, rng, dmg, exp := "", "", "", "", "", "", "", "", "", ""

			// 1. Motives
			if m := regexp.MustCompile(`(?i)[*][*]Motives & Tactics:[*][*]\s*(.*)`).FindStringSubmatch(fullText); len(m) > 1 {
				motives = strings.TrimSpace(m[1])
			} else {
				warnMissing("Adversary", name, "Motives & Tactics")
			}

			// 2. Stats (Difficulty, Thresholds, HP, Stress) - can be in one line or separate
			if m := regexp.MustCompile(`(?i)[*][*]Difficulty:[*][*]\s*(\d+)`).FindStringSubmatch(fullText); len(m) > 1 {
				diff = m[1]
			} else {
				warnMissing("Adversary", name, "Difficulty")
			}

			if m := regexp.MustCompile(`(?i)[*][*]Thresholds:[*][*]\s*(\d+/(?:\d+|None)|None)`).FindStringSubmatch(fullText); len(m) > 1 {
				thresh = m[1]
			}

			if m := regexp.MustCompile(`(?i)[*][*]HP:[*][*]\s*(\d+)`).FindStringSubmatch(fullText); len(m) > 1 {
				hp = m[1]
			} else {
				warnMissing("Adversary", name, "HP")
			}

			if m := regexp.MustCompile(`(?i)[*][*]Stress:[*][*]\s*(\d+|-)`).FindStringSubmatch(fullText); len(m) > 1 {
				stress = m[1]
			}

			// 3. Attack (ATK, Name, Range, Damage)
			reAtkLine := regexp.MustCompile(`(?i)[*][*]ATK:[*][*]\s*([+\-](?:\d*d\d+|\d+))\s*\|\s*[*][*](.*?):[*][*]\s*(.*?)\s*\|\s*(.*)`)
			if m := reAtkLine.FindStringSubmatch(fullText); len(m) > 4 {
				atk = strings.TrimSpace(m[1])
				attackName = strings.TrimSpace(m[2])
				rng = strings.TrimSpace(m[3])
				dmg = strings.TrimSpace(m[4])
			} else if m := regexp.MustCompile(`(?i)[*][*]ATK:[*][*]\s*([+\-](?:\d*d\d+|\d+))`).FindStringSubmatch(fullText); len(m) > 1 {
				atk = strings.TrimSpace(m[1])
			}
			if atk == "" {
				warnMissing("Adversary", name, "ATK")
			}
			if attackName == "" || rng == "" || dmg == "" {
				warnMissing("Adversary", name, "Attack Line")
			}

			// 4. Experience
			if m := regexp.MustCompile(`(?i)[*][*]Experience:[*][*]\s*(.*)`).FindStringSubmatch(fullText); len(m) > 1 {
				exp = strings.TrimSpace(m[1])
			}

			// Features
			var feats []string
			reFeat := regexp.MustCompile(`(?m)^#{4,6}\s+FEATURE(?:S|\(S\))?\s*$`)
			if idx, header := findHeaderBlockStart(fullText, reFeat); idx != -1 {
				if header != "##### FEATURES" {
					warnFormat("Adversary", name, fmt.Sprintf("features heading '%s'", header))
				}
				fLines := strings.Split(fullText[idx:], "\n")
				currN, currT := "", ""
				for i, l := range fLines {
					l = strings.TrimSpace(l)
					if strings.HasPrefix(l, "#") {
						if i == 0 {
							continue
						}
						break
					}
					if strings.HasPrefix(l, "_") && strings.Contains(l, ":_") {
						if currN != "" {
							feats = append(feats, currN, strings.Trim(currT, "\n"))
						}
						parts := strings.SplitN(l, ":_", 2)
						currN = strings.TrimPrefix(parts[0], "_")
						currN = titleizeIfAllCaps(currN)
						currT = strings.TrimSpace(parts[1])
					} else if currN != "" && !strings.HasPrefix(l, "#####") {
						if l == "" {
							currT += "\n"
						} else {
							currT += "\n" + l
						}
					}
				}
				if currN != "" {
					feats = append(feats, currN, strings.Trim(currT, "\n"))
				}
			} else {
				warnMissing("Adversary", name, "Features")
			}

			row := []string{name, tier, typeStr, desc, motives, diff, thresh, hp, stress, atk, attackName, rng, dmg, exp}
			for i := 0; i < 14; i++ {
				if i < len(feats) {
					row = append(row, feats[i])
				} else {
					row = append(row, "")
				}
			}
			records = append(records, row)
		}
	}
	writeCSV(outputDir+"/adversaries.csv", []string{"Name", "Tier", "Type", "Description", "Motives and Tactics", "Difficulty", "Thresholds", "HP", "Stress", "ATK", "Attack", "Range", "Damage", "Experience", "Feature 1 Name", "Feature 1 Text", "Feature 2 Name", "Feature 2 Text", "Feature 3 Name", "Feature 3 Text", "Feature 4 Name", "Feature 4 Text", "Feature 5 Name", "Feature 5 Text", "Feature 6 Name", "Feature 6 Text", "Feature 7 Name", "Feature 7 Text"}, records)
}

func extractAncestries(content, outputDir string) {
	section := getSection(content, "## ANCESTRIES")
	chunks := strings.Split(section, "\n### ")
	var records [][]string
	for i, chunk := range chunks {
		if i == 0 {
			continue
		}
		lines := strings.Split(chunk, "\n")
		name := titleizeIfAllCaps(strings.TrimSpace(lines[0]))
		desc, feats := "", []string{}

		reFeat := regexp.MustCompile(`(?m)^#{4,6}\s+ANCESTRY FEATURE(S)?\b.*$`)
		splitIdx, header := findHeaderBlockStart(chunk, reFeat)
		if splitIdx == -1 {
			warnMissing("Ancestry", name, "Ancestry Features")
			continue
		}
		if header != "##### ANCESTRY FEATURES" && header != "##### ANCESTRY FEATURE" {
			warnFormat("Ancestry", name, fmt.Sprintf("ancestry feature heading '%s'", header))
		}
		descPart := chunk[len(lines[0]):splitIdx]
		desc = strings.Trim(descPart, "\n")
		featPart := chunk[splitIdx:]
		currN, currT := "", ""
		for i, l := range strings.Split(featPart, "\n") {
			l = strings.TrimSpace(l)
			if strings.HasPrefix(l, "#") {
				if i == 0 {
					continue
				}
				break
			}
			if strings.HasPrefix(l, "_") && strings.Contains(l, ":_") {
				if currN != "" {
					feats = append(feats, currN, strings.Trim(currT, "\n"))
				}
				p := strings.SplitN(l, ":_", 2)
				currN = strings.TrimPrefix(p[0], "_")
				currN = titleizeIfAllCaps(currN)
				currT = strings.TrimSpace(p[1])
			} else if currN != "" && !strings.HasPrefix(l, "#####") {
				if l == "" {
					currT += "\n"
				} else {
					currT += "\n" + l
				}
			}
		}
		if currN != "" {
			feats = append(feats, currN, strings.Trim(currT, "\n"))
		}
		if len(feats) == 0 {
			warnMissing("Ancestry", name, "Features")
		}
		row := []string{name, strings.TrimSpace(desc)}
		for k := 0; k < 4; k++ {
			if k < len(feats) {
				row = append(row, feats[k])
			} else {
				row = append(row, "")
			}
		}
		records = append(records, row)
	}
	writeCSV(outputDir+"/ancestries.csv", []string{"Name", "Description", "Feature 1 Name", "Feature 1 Text", "Feature 2 Name", "Feature 2 Text"}, records)
}

func extractArmor(content, outputDir string) {
	section := getSection(content, "## ARMOR")
	var records [][]string
	currentTier := ""
	reTier := regexp.MustCompile(`(?i)^####\s+TIER\s+(\d+)`)
	for _, line := range strings.Split(section, "\n") {
		if m := reTier.FindStringSubmatch(line); len(m) > 1 {
			currentTier = strings.TrimSpace(m[1])
		}
		cells, ok := parseMarkdownTableRow(line, 4)
		if !ok {
			continue
		}
		if strings.EqualFold(cells[0], "name") || strings.EqualFold(cells[1], "thresholds") {
			continue
		}
		name, thresh, score, featFull := cells[0], cells[1], cells[2], cells[3]
		name = titleizeIfAllCaps(name)
		if name == "" {
			continue
		}
		featN, featT := "", ""
		if strings.Contains(featFull, ":") {
			p := strings.SplitN(featFull, ":", 2)
			featN = strings.Trim(p[0], "*_ ")
			featN = titleizeIfAllCaps(featN)
			featT = cleanInlineText(p[1])
		} else {
			featT = cleanInlineText(featFull)
		}
		if featT == "—" {
			featN, featT = "", ""
		}
		if thresh == "" {
			warnMissing("Armor", name, "Base Thresholds")
		}
		if score == "" {
			warnMissing("Armor", name, "Base Score")
		}
		records = append(records, []string{name, currentTier, thresh, score, featN, featT})
	}
	writeCSV(outputDir+"/armor.csv", []string{"Name", "Tier", "Base Thresholds", "Base Score", "Feature 1 Name", "Feature 1 Text"}, records)
}

func extractClasses(content, outputDir string) {
	idx := findHeaderIndex(content, "## CLASSES")
	if idx == -1 {
		return
	}
	// We consume all L2 sections until we hit one that doesn't look like a class
	chunks := regexp.MustCompile(`(?m)^## `).Split(content[idx:], -1)
	var records [][]string
	for i, chunk := range chunks {
		if i == 0 || strings.HasPrefix(chunk, "CLASSES") {
			continue
		}
		if !strings.Contains(chunk, "CLASS FEATURE") && !strings.Contains(chunk, "CLASS FEATURES") && !strings.Contains(chunk, "DOMAINS -") {
			break
		}

		lines := strings.Split(strings.TrimSpace(chunk), "\n")
		name := titleizeIfAllCaps(lines[0])
		doms, evasion, hp, items := "", "", "", ""
		sugTraits, sugPrimary, sugSecondary, sugArmor := "", "", "", ""
		reDomains := regexp.MustCompile(`(?i)\*\*DOMAINS\s*-\*\*\s*(.*)`)
		reEvasion := regexp.MustCompile(`(?i)\*\*STARTING EVASION\s*-\*\*\s*(\d+)`)
		reHP := regexp.MustCompile(`(?i)\*\*STARTING HIT POINTS\s*-\*\*\s*(\d+)`)
		reItems := regexp.MustCompile(`(?i)\*\*CLASS ITEMS\s*-\*\*\s*(.*)`)
		reSugTraits := regexp.MustCompile(`(?i)\*\*SUGGESTED TRAITS\s*-\*\*\s*(.*)`)
		reSugPrimary := regexp.MustCompile(`(?i)\*\*SUGGESTED PRIMARY\s*-\*\*\s*(.*)`)
		reSugSecondary := regexp.MustCompile(`(?i)\*\*SUGGESTED SECONDARY\s*-\*\*\s*(.*)`)
		reSugArmor := regexp.MustCompile(`(?i)\*\*SUGGESTED ARMOR\s*-\*\*\s*(.*)`)
		if m := reDomains.FindStringSubmatch(chunk); len(m) > 1 {
			doms = strings.TrimSpace(m[1])
			doms = strings.ReplaceAll(doms, "&", ",")
			if strings.Contains(doms, " and ") {
				warnFormat("Class", name, "domains uses 'and' instead of '&'")
				doms = strings.ReplaceAll(doms, " and ", ",")
			}
		}
		if m := reEvasion.FindStringSubmatch(chunk); len(m) > 1 {
			evasion = strings.TrimSpace(m[1])
		}
		if m := reHP.FindStringSubmatch(chunk); len(m) > 1 {
			hp = strings.TrimSpace(m[1])
		}
		if m := reItems.FindStringSubmatch(chunk); len(m) > 1 {
			items = strings.TrimSpace(m[1])
		}
		if m := reSugTraits.FindStringSubmatch(chunk); len(m) > 1 {
			sugTraits = strings.TrimSpace(m[1])
		}
		if m := reSugPrimary.FindStringSubmatch(chunk); len(m) > 1 {
			sugPrimary = strings.TrimSpace(m[1])
		}
		if m := reSugSecondary.FindStringSubmatch(chunk); len(m) > 1 {
			sugSecondary = strings.TrimSpace(m[1])
		}
		if m := reSugArmor.FindStringSubmatch(chunk); len(m) > 1 {
			sugArmor = strings.TrimSpace(m[1])
		}
		d1, d2 := "", ""
		ds := strings.Split(doms, ",")
		if len(ds) > 0 {
			d1 = titleizeIfAllCaps(strings.TrimSpace(ds[0]))
		}
		if len(ds) > 1 {
			d2 = titleizeIfAllCaps(strings.TrimSpace(ds[1]))
		}

		hopeN, hopeT := "", ""
		reHope := regexp.MustCompile(`(?m)^#{4,6}\s+.*HOPE FEATURE\b.*$`)
		if hi, header := findHeaderBlockStart(chunk, reHope); hi != -1 {
			if !strings.HasPrefix(header, "#####") {
				warnFormat("Class", name, fmt.Sprintf("hope feature heading '%s'", header))
			}
			feats := extractFeaturePairs(chunk[hi:])
			if len(feats) >= 2 {
				hopeN, hopeT = feats[0], feats[1]
			}
		}

		cfFeats := []string{}
		reCF := regexp.MustCompile(`(?m)^#{4,6}\s+CLASS FEATURE(S)?\b.*$`)
		if ci, header := findHeaderBlockStart(chunk, reCF); ci != -1 {
			if header != "##### CLASS FEATURE" && header != "##### CLASS FEATURES" {
				warnFormat("Class", name, fmt.Sprintf("class feature heading '%s'", header))
			}
			cfFeats = extractFeaturePairs(chunk[ci:])
		}

		bg := []string{"", "", ""}
		if bi := strings.Index(chunk, "#### BACKGROUND QUESTIONS"); bi != -1 {
			c := 0
			for _, l := range strings.Split(chunk[bi:], "\n") {
				if strings.HasPrefix(strings.TrimSpace(l), "- ") {
					bg[c] = strings.TrimPrefix(strings.TrimSpace(l), "- ")
					c++
					if c == 3 {
						break
					}
				}
			}
		}
		cn := []string{"", "", ""}
		if ci := strings.Index(chunk, "#### CONNECTIONS"); ci != -1 {
			c := 0
			for _, l := range strings.Split(chunk[ci:], "\n") {
				if strings.HasPrefix(strings.TrimSpace(l), "- ") {
					cn[c] = strings.TrimPrefix(strings.TrimSpace(l), "- ")
					c++
					if c == 3 {
						break
					}
				}
			}
		}

		sub1, sub2 := "", ""
		reS := regexp.MustCompile(`(?m)^#### (.*)`)
		sms := reS.FindAllStringSubmatch(chunk, -1)
		validSubs := []string{}
		for _, sm := range sms {
			if !strings.Contains(sm[1], "QUESTIONS") && !strings.Contains(sm[1], "CONNECTIONS") {
				validSubs = append(validSubs, strings.TrimSpace(sm[1]))
			}
		}
		if len(validSubs) > 0 {
			sub1 = titleizeIfAllCaps(validSubs[0])
		}
		if len(validSubs) > 1 {
			sub2 = titleizeIfAllCaps(validSubs[1])
		}

		descLines := []string{}
		for _, l := range lines[1:] {
			t := strings.TrimRight(l, " ")
			if strings.HasPrefix(strings.TrimSpace(t), "---") || strings.Contains(t, "DOMAINS -") {
				break
			}
			descLines = append(descLines, t)
		}
		desc := strings.Trim(strings.Join(descLines, "\n"), "\n")

		cf1n, cf1t, cf2n, cf2t, cf3n, cf3t := "", "", "", "", "", ""
		if len(cfFeats) >= 2 {
			cf1n, cf1t = cfFeats[0], cfFeats[1]
		}
		if len(cfFeats) >= 4 {
			cf2n, cf2t = cfFeats[2], cfFeats[3]
		}
		if len(cfFeats) >= 6 {
			cf3n, cf3t = cfFeats[4], cfFeats[5]
		}

		records = append(records, []string{name, strings.TrimSpace(desc), d1, d2, evasion, hp, items, sugTraits, sugPrimary, sugSecondary, sugArmor, hopeN, hopeT, cf1n, cf1t, cf2n, cf2t, cf3n, cf3t, bg[0], bg[1], bg[2], cn[0], cn[1], cn[2], sub1, sub2})
	}
	writeCSV(outputDir+"/classes.csv", []string{"Name", "Description", "Domain 1", "Domain 2", "Evasion", "HP", "Items", "Suggested Traits", "Suggested Primary", "Suggested Secondary", "Suggested Armor", "Hope Feature Name", "Hope Feature Text", "Feature 1 Name", "Feature 1 Text", "Feature 2 Name", "Feature 2 Text", "Feature 3 Name", "Feature 3 Text", "Background 1 Question", "Background 2 Question", "Background 3 Question", "Connection 1 Question", "Connection 2 Question", "Connection 3 Question", "Subclass 1", "Subclass 2"}, records)
}

func extractCommunities(content, outputDir string) {
	section := getSection(content, "## COMMUNITIES")
	chunks := strings.Split(section, "\n### ")
	var records [][]string
	for i, chunk := range chunks {
		if i == 0 {
			continue
		}
		name := titleizeIfAllCaps(strings.TrimSpace(strings.Split(chunk, "\n")[0]))
		desc, note, fN, fT := "", "", "", ""
		fi := strings.Index(chunk, "##### COMMUNITY FEATURE")
		if fi != -1 {
			descRaw := strings.Trim(chunk[len(name):fi], "\n")
			parts := regexp.MustCompile(`\n\s*\n`).Split(descRaw, -1)
			if len(parts) > 0 {
				last := strings.TrimSpace(parts[len(parts)-1])
				if strings.HasPrefix(last, "_") {
					end := strings.LastIndex(last, "_")
					if end > 0 {
						note = strings.TrimSpace(last[1:end])
						desc = strings.Trim(strings.Join(parts[:len(parts)-1], "\n\n"), "\n")
					} else {
						warnFormat("Community", name, "note paragraph missing closing underscore")
						desc = descRaw
					}
				} else {
					desc = descRaw
					warnFormat("Community", name, "note paragraph not italicized")
				}
			} else {
				desc = descRaw
			}
			for _, l := range strings.Split(chunk[fi:], "\n") {
				if strings.Contains(l, ":") && !strings.Contains(l, "FEATURE") {
					line := strings.TrimSpace(l)
					p := strings.SplitN(line, ":_", 2)
					if len(p) < 2 {
						p = strings.SplitN(line, ":", 2)
					}
					if len(p) < 2 {
						continue
					}
					fN = strings.Trim(p[0], "_* ")
					fN = titleizeIfAllCaps(fN)
					fT = strings.TrimSpace(p[1])
					if strings.HasPrefix(fT, "_") {
						fT = strings.TrimSpace(strings.TrimPrefix(fT, "_"))
					}
					break
				}
			}
		}
		records = append(records, []string{name, strings.TrimSpace(desc), note, fN, fT})
	}
	writeCSV(outputDir+"/communities.csv", []string{"Name", "Description", "Note", "Feature 1 Name", "Feature 1 Text"}, records)
}

func extractConsumables(content, outputDir string) {
	section := getSection(content, "## CONSUMABLES")
	var records [][]string
	for _, line := range strings.Split(section, "\n") {
		cells, ok := parseMarkdownTableRow(line, 3)
		if !ok {
			continue
		}
		if strings.EqualFold(cells[0], "roll") {
			continue
		}
		roll, name, desc := cells[0], strings.Trim(cells[1], "* "), cells[2]
		name = titleizeIfAllCaps(name)
		name = titleizeIfAllCaps(name)
		if roll == "" || name == "" {
			warnMissing("Consumable", name, "Roll or Name")
		}
		records = append(records, []string{roll, name, desc})
	}
	writeCSV(outputDir+"/consumables.csv", []string{"Roll", "Name", "Description"}, records)
}

func extractDomains(content, outputDir string) {
	section := getSection(content, "## DOMAINS")
	cardMap := extractDomainCardReference(content)
	chunks := strings.Split(section, "\n#### ")
	var records [][]string
	headers := []string{"Name", "Description"}
	for level := 1; level <= 10; level++ {
		headers = append(headers,
			fmt.Sprintf("Card %d.1", level),
			fmt.Sprintf("Card %d.2", level),
			fmt.Sprintf("Card %d.3", level),
		)
	}
	for i, chunk := range chunks {
		if i == 0 || strings.HasPrefix(chunk, "##") {
			continue
		}
		lines := strings.Split(chunk, "\n")
		name := titleizeIfAllCaps(strings.TrimSpace(lines[0]))
		desc := strings.Trim(strings.Join(lines[1:], "\n"), "\n")
		row := []string{name, desc}
		cardKey := normalizeHeading(name)
		cards := cardMap[cardKey]
		expected := 21
		if len(cards) != expected {
			fmt.Printf("Warning: Domain '%s' has %d cards (expected %d)\n", name, len(cards), expected)
		}
		idx := 0
		for level := 1; level <= 10; level++ {
			opts := []string{"", "", ""}
			need := 2
			if level == 1 {
				need = 3
			}
			for j := 0; j < need; j++ {
				if idx < len(cards) {
					opts[j] = cards[idx]
				}
				idx++
			}
			row = append(row, opts...)
		}
		if idx < len(cards) {
			fmt.Printf("Warning: Domain '%s' has %d extra cards\n", name, len(cards)-idx)
		}
		records = append(records, row)
	}
	writeCSV(outputDir+"/domains.csv", headers, records)
}

func extractDomainCardReference(content string) map[string][]string {
	section := getSection(content, "## DOMAIN CARD REFERENCE")
	chunks := strings.Split(section, "\n### ")
	results := map[string][]string{}
	reCard := regexp.MustCompile(`(?m)^#{4,6}\s+(.+)$`)
	for i, chunk := range chunks {
		if i == 0 {
			continue
		}
		lines := strings.Split(chunk, "\n")
		if len(lines) == 0 {
			continue
		}
		domainLine := strings.TrimSpace(lines[0])
		domainName := strings.TrimSpace(strings.TrimSuffix(domainLine, "DOMAIN"))
		domainName = strings.TrimSpace(strings.TrimSuffix(domainName, "Domain"))
		domainName = titleizeIfAllCaps(domainName)
		if domainName == "" {
			continue
		}
		var cards []string
		matches := reCard.FindAllStringSubmatch(chunk, -1)
		for _, match := range matches {
			cardName := titleizeIfAllCaps(strings.TrimSpace(match[1]))
			if cardName == "" {
				continue
			}
			cards = append(cards, cardName)
		}
		results[normalizeHeading(domainName)] = cards
	}
	return results
}

func extractEnvironments(content, outputDir string) {
	reL2 := regexp.MustCompile(`(?m)^## (.*ENVIRONMENTS.*)`)
	matches := reL2.FindAllStringSubmatch(content, -1)
	var sb strings.Builder
	for _, m := range matches {
		if strings.Contains(m[1], "USING") {
			continue
		}
		sb.WriteString(getSection(content, "## "+m[1]))
		sb.WriteString("\n")
	}
	section := sb.String()
	chunks := strings.Split(section, "\n### ")
	var records [][]string
	for i, chunk := range chunks {
		if i == 0 {
			continue
		}
		name := titleizeIfAllCaps(strings.TrimSpace(strings.Split(chunk, "\n")[0]))
		fullText := chunk
		reT := regexp.MustCompile(`[*][*]_(Tier\s+\d+\s+.*?)_[*][*]`)
		tm := reT.FindStringSubmatch(fullText)
		if tm == nil {
			// Check if it has Impulses anyway
			if !strings.Contains(fullText, "Impulses:") {
				continue
			}
		}
		tier, typ := "", ""
		if tm != nil {
			ps := strings.SplitN(tm[1], " ", 3)
			if len(ps) >= 2 {
				tier = ps[1]
			}
			if len(ps) > 2 {
				typ = titleizeIfAllCaps(ps[2])
			}
		}
		desc := ""
		reD := regexp.MustCompile(`[*][*]_.*?_[*][*]\n_(.*?)_`)
		if m := reD.FindStringSubmatch(fullText); len(m) > 1 {
			desc = strings.TrimSpace(m[1])
		}
		if desc == "" {
			fmt.Printf("Warning: Environment '%s' missing Description\n", name)
		}
		imp, diff, adv := "", "", ""
		reImp := regexp.MustCompile(`[*][*]Impulses:[*][*]\s*(.*)`)
		if m := reImp.FindStringSubmatch(fullText); len(m) > 1 {
			imp = strings.TrimSpace(m[1])
		}
		reDiff := regexp.MustCompile(`[*][*]Difficulty:[*][*]\s*([^\n]+)`)
		if m := reDiff.FindStringSubmatch(fullText); len(m) > 1 {
			diff = strings.TrimSpace(m[1])
		}
		reAdv := regexp.MustCompile(`[*][*]Potential Adversaries:[*][*]\s*(.*)`)
		if m := reAdv.FindStringSubmatch(fullText); len(m) > 1 {
			adv = strings.TrimSpace(m[1])
		}

		var feats []string
		reFeat := regexp.MustCompile(`(?m)^#{4,6}\s+FEATURE(?:S|\(S\))?\s*$`)
		if idx, header := findHeaderBlockStart(fullText, reFeat); idx != -1 {
			if header != "##### FEATURES" {
				warnFormat("Environment", name, fmt.Sprintf("features heading '%s'", header))
			}
			fLines := strings.Split(fullText[idx:], "\n")
			currN, currT := "", ""
			for i, l := range fLines {
				l = strings.TrimSpace(l)
				if strings.HasPrefix(l, "#") {
					if i == 0 {
						continue
					}
					break
				}
				if strings.HasPrefix(l, "_") && strings.Contains(l, ":_") {
					if currN != "" {
						feats = append(feats, currN, strings.Trim(currT, "\n"))
					}
					parts := strings.SplitN(l, ":_", 2)
					currN = strings.TrimPrefix(parts[0], "_")
					currN = titleizeIfAllCaps(currN)
					currT = strings.TrimSpace(parts[1])
				} else if currN != "" && !strings.HasPrefix(l, "#####") {
					if l == "" {
						currT += "\n"
					} else {
						currT += "\n" + l
					}
				}
			}
			if currN != "" {
				feats = append(feats, currN, strings.Trim(currT, "\n"))
			}
		}
		row := []string{name, tier, typ, desc, imp, diff, adv}
		for k := 0; k < 6; k++ {
			featName := ""
			featText := ""
			if k*2 < len(feats) {
				featName = feats[k*2]
			}
			if k*2+1 < len(feats) {
				featText = feats[k*2+1]
			}
			featQuestions := ""
			if featText != "" {
				qs := extractQuestions(featText)
				if len(qs) > 0 {
					featQuestions = strings.Join(qs, " ")
				}
			}
			row = append(row, featName, featText, featQuestions)
		}
		records = append(records, row)
	}
	writeCSV(outputDir+"/environments.csv", []string{"Name", "Tier", "Type", "Description", "Impulses", "Difficulty", "Potential Adversaries", "Feature 1 Name", "Feature 1 Text", "Feature 1 Question", "Feature 2 Name", "Feature 2 Text", "Feature 2 Question", "Feature 3 Name", "Feature 3 Text", "Feature 3 Question", "Feature 4 Name", "Feature 4 Text", "Feature 4 Question", "Feature 5 Name", "Feature 5 Text", "Feature 5 Question", "Feature 6 Name", "Feature 6 Text", "Feature 6 Question"}, records)
}

func extractQuestions(text string) []string {
	cleanLine := func(line string) string {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, ">") {
			line = strings.TrimSpace(strings.TrimPrefix(line, ">"))
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			line = strings.TrimSpace(line[2:])
		}
		if m := regexp.MustCompile(`^\d+\.\s+`).FindString(line); m != "" {
			line = strings.TrimSpace(strings.TrimPrefix(line, m))
		}
		return strings.TrimSpace(line)
	}

	var out []string
	seen := map[string]bool{}
	addQuestion := func(q string) {
		q = strings.TrimSpace(q)
		if q != "" {
			q = strings.Join(strings.Fields(q), " ")
		}
		if q == "" || seen[q] {
			return
		}
		seen[q] = true
		out = append(out, q)
	}

	lines := strings.Split(text, "\n")
	reSentence := regexp.MustCompile(`[^?]*\?`)
	reStart := regexp.MustCompile(`(?i)^(who|what|when|where|why|how|which|can|do|did|is|are|should|would|could)\b`)

	extractFromSegment := func(segment string) {
		for _, match := range reSentence.FindAllString(segment, -1) {
			q := cleanLine(match)
			q = strings.Trim(q, `"'`)
			if q == "" || !strings.HasSuffix(q, "?") {
				continue
			}
			if reStart.MatchString(strings.ToLower(q)) {
				addQuestion(q)
			}
		}
	}

	italicMatches := regexp.MustCompile(`_(.+?)_`).FindAllStringSubmatch(text, -1)
	for _, m := range italicMatches {
		if len(m) < 2 {
			continue
		}
		segment := strings.TrimSpace(m[1])
		if !strings.Contains(segment, "?") {
			continue
		}
		extractFromSegment(segment)
	}
	if len(out) > 0 {
		return out
	}

	for _, raw := range lines {
		line := cleanLine(raw)
		if line == "" {
			continue
		}
		if strings.Contains(line, "?") {
			if strings.HasSuffix(line, "?") && !strings.Contains(line[:len(line)-1], "?") {
				addQuestion(line)
				continue
			}
			extractFromSegment(line)
		}
	}

	if len(out) == 0 {
		extractFromSegment(text)
	}

	return out
}

func extractBeastforms(content, outputDir string) {
	idx := findHeaderIndex(content, "#### BEASTFORM OPTIONS")
	if idx == -1 {
		fmt.Println("Warning: Beastform options section not found")
		return
	}

	rest := content[idx:]
	firstLineEnd := strings.Index(rest, "\n")
	if firstLineEnd == -1 {
		return
	}
	search := rest[firstLineEnd+1:]
	reEnd := regexp.MustCompile(`(?m)^##\s`)
	if loc := reEnd.FindStringIndex(search); loc != nil {
		rest = rest[:firstLineEnd+1+loc[0]]
	}
	section := rest

	reTier := regexp.MustCompile(`(?m)^####\s+TIER\s+(\d+)`)
	tiers := reTier.FindAllStringSubmatchIndex(section, -1)
	if len(tiers) == 0 {
		fmt.Println("Warning: No Beastform tiers found")
		return
	}

	var records [][]string
	for i, loc := range tiers {
		tier := section[loc[2]:loc[3]]
		start := loc[0]
		end := len(section)
		if i+1 < len(tiers) {
			end = tiers[i+1][0]
		}
		tierBlock := section[start:end]
		chunks := strings.Split(tierBlock, "\n##### ")
		for j, chunk := range chunks {
			if j == 0 {
				continue
			}
			lines := strings.Split(chunk, "\n")
			if len(lines) == 0 {
				continue
			}
			name := titleizeIfAllCaps(strings.TrimSpace(lines[0]))
			examples := ""
			statsLine := ""
			advantages := ""

			for _, raw := range lines[1:] {
				line := strings.TrimSpace(raw)
				if line == "" {
					continue
				}
				if examples == "" && strings.HasPrefix(line, "(") && strings.HasSuffix(line, ")") {
					examples = line
					continue
				}
				if statsLine == "" && strings.HasPrefix(line, "_") && strings.Contains(line, "|") {
					statsLine = strings.Trim(line, "_")
					continue
				}
				if advantages == "" && strings.HasPrefix(line, "**Gain advantage on:**") {
					advantages = strings.TrimSpace(strings.TrimPrefix(line, "**Gain advantage on:**"))
					continue
				}
			}

			traitBonus, evasionBonus, attack := "", "", ""
			if statsLine != "" {
				parts := strings.Split(statsLine, "|")
				if len(parts) >= 1 {
					traitBonus = strings.TrimSpace(parts[0])
				}
				if len(parts) >= 2 {
					evasionBonus = strings.TrimSpace(parts[1])
				}
				if len(parts) >= 3 {
					attack = strings.TrimSpace(parts[2])
				}
				if traitBonus == "" || evasionBonus == "" || attack == "" {
					warnFormat("Beastform", name, fmt.Sprintf("stats line '%s'", statsLine))
				}
			}

			feats := extractFeaturePairs(chunk)
			if len(feats) == 0 {
				warnMissing("Beastform", name, "Features")
			}

			row := []string{name, tier, examples, traitBonus, evasionBonus, attack, advantages}
			for k := 0; k < 3; k++ {
				if k*2 < len(feats) {
					row = append(row, feats[k*2])
				} else {
					row = append(row, "")
				}
				if k*2+1 < len(feats) {
					row = append(row, feats[k*2+1])
				} else {
					row = append(row, "")
				}
			}
			records = append(records, row)
		}
	}

	writeCSV(outputDir+"/beastforms.csv", []string{"Name", "Tier", "Examples", "Trait Bonus", "Evasion Bonus", "Attack", "Advantages", "Feature 1 Name", "Feature 1 Text", "Feature 2 Name", "Feature 2 Text", "Feature 3 Name", "Feature 3 Text"}, records)
}

func extractItems(content, outputDir string) {
	section := getSection(content, "## LOOT")
	var records [][]string
	for _, line := range strings.Split(section, "\n") {
		cells, ok := parseMarkdownTableRow(line, 3)
		if !ok {
			continue
		}
		if strings.EqualFold(cells[0], "roll") {
			continue
		}
		roll, name, desc := cells[0], strings.Trim(cells[1], "* "), cells[2]
		if roll == "" || name == "" {
			warnMissing("Item", name, "Roll or Name")
		}
		records = append(records, []string{roll, name, desc})
	}
	writeCSV(outputDir+"/items.csv", []string{"Roll", "Name", "Description"}, records)
}

func extractSubclasses(content, outputDir string) {
	idx := findHeaderIndex(content, "## CLASSES")
	if idx == -1 {
		return
	}
	chunks := regexp.MustCompile(`(?m)^## `).Split(content[idx:], -1)
	var records [][]string
	for i, chunk := range chunks {
		if i == 0 || strings.HasPrefix(chunk, "CLASSES") {
			continue
		}
		if !strings.Contains(chunk, "CLASS FEATURE") && !strings.Contains(chunk, "CLASS FEATURES") && !strings.Contains(chunk, "DOMAINS -") {
			break
		}

		subChunks := regexp.MustCompile(`(?m)^#### `).Split(chunk, -1)
		for j, subChunk := range subChunks {
			if j == 0 || !strings.Contains(subChunk, "FOUNDATION FEATURE") {
				continue
			}
			lines := strings.Split(strings.TrimSpace(subChunk), "\n")
			name := titleizeIfAllCaps(lines[0])
			spellcast := ""
			reSpell := regexp.MustCompile(`(?m)^#{4,6}\s+SPELLCAST TRAIT\b.*$`)
			if si, header := findHeaderBlockStart(subChunk, reSpell); si != -1 {
				if header != "##### SPELLCAST TRAIT" {
					warnFormat("Subclass", name, fmt.Sprintf("spellcast heading '%s'", header))
				}
				for _, l := range strings.Split(subChunk[si:], "\n") {
					t := strings.TrimSpace(l)
					if t == "" || strings.HasPrefix(t, "#") {
						continue
					}
					spellcast = t
					break
				}
			}
			descLines := []string{}
			for _, l := range lines[1:] {
				if strings.HasPrefix(strings.TrimSpace(l), "#####") {
					break
				}
				descLines = append(descLines, strings.TrimRight(l, " "))
			}
			desc := strings.Trim(strings.Join(descLines, "\n"), "\n")

			f1n, f1t, f2n, f2t := "", "", "", ""
			reFoundation := regexp.MustCompile(`(?m)^#{4,6}\s+FOUNDATION FEATURE(S)?\b.*$`)
			if fi, header := findHeaderBlockStart(subChunk, reFoundation); fi != -1 {
				if header != "##### FOUNDATION FEATURE" && header != "##### FOUNDATION FEATURES" {
					warnFormat("Subclass", name, fmt.Sprintf("foundation heading '%s'", header))
				}
				ff := extractFeaturePairs(subChunk[fi:])
				if len(ff) >= 2 {
					f1n, f1t = ff[0], ff[1]
				}
				if len(ff) >= 4 {
					f2n, f2t = ff[2], ff[3]
				}
			}

			s1n, s1t, s2n, s2t := "", "", "", ""
			reSpec := regexp.MustCompile(`(?m)^#{4,6}\s+SPECIALIZATION FEATURE(S)?\b.*$`)
			if si, header := findHeaderBlockStart(subChunk, reSpec); si != -1 {
				if header != "##### SPECIALIZATION FEATURE" && header != "##### SPECIALIZATION FEATURES" {
					warnFormat("Subclass", name, fmt.Sprintf("specialization heading '%s'", header))
				}
				sf := extractFeaturePairs(subChunk[si:])
				if len(sf) >= 2 {
					s1n, s1t = sf[0], sf[1]
				}
				if len(sf) >= 4 {
					s2n, s2t = sf[2], sf[3]
				}
			}

			m1n, m1t, m2n, m2t := "", "", "", ""
			reMastery := regexp.MustCompile(`(?m)^#{4,6}\s+MASTERY FEATURE(S)?\b.*$`)
			if mi, header := findHeaderBlockStart(subChunk, reMastery); mi != -1 {
				if header != "##### MASTERY FEATURE" && header != "##### MASTERY FEATURES" {
					warnFormat("Subclass", name, fmt.Sprintf("mastery heading '%s'", header))
				}
				mf := extractFeaturePairs(subChunk[mi:])
				if len(mf) >= 2 {
					m1n, m1t = mf[0], mf[1]
				}
				if len(mf) >= 4 {
					m2n, m2t = mf[2], mf[3]
				}
			}
			records = append(records, []string{name, strings.TrimSpace(desc), spellcast, f1n, f1t, f2n, f2t, s1n, s1t, s2n, s2t, m1n, m1t, m2n, m2t})
		}
	}

	writeCSV(outputDir+"/subclasses.csv", []string{"Name", "Description", "Spellcast Trait", "Foundation 1 Name", "Foundation 1 Text", "Foundation 2 Name", "Foundation 2 Text", "Specialization 1 Name", "Specialization 1 Text", "Specialization 2 Name", "Specialization 2 Text", "Mastery 1 Name", "Mastery 1 Text", "Mastery 2 Name", "Mastery 2 Text"}, records)
}

func extractWeapons(content, outputDir string) {
	section := getSection(content, "## WEAPONS")

	var records [][]string
	tier, category, damageType := "", "Primary", ""
	reTier := regexp.MustCompile(`(?i)^####\s+TIER\s+(\d+)(?:.*?(Physical|Magic))?`)
	for _, line := range strings.Split(section, "\n") {
		if strings.Contains(strings.ToUpper(line), "PRIMARY WEAPON TABLES") {
			category = "Primary"
		}
		if strings.Contains(strings.ToUpper(line), "SECONDARY WEAPON TABLES") {
			category = "Secondary"
		}
		if m := reTier.FindStringSubmatch(line); len(m) > 1 {
			tier = strings.TrimSpace(m[1])
			damageType = ""
			if len(m) > 2 && strings.TrimSpace(m[2]) != "" {
				damageType = "Physical"
				if strings.EqualFold(strings.TrimSpace(m[2]), "Magic") {
					damageType = "Magical"
				}
			}
		}

		cells, ok := parseMarkdownTableRow(line, 6)
		if !ok {
			continue
		}
		if strings.EqualFold(cells[0], "name") || strings.EqualFold(cells[1], "trait") {
			continue
		}
		n, trait, rng, dmg, burden, featFull := cells[0], cells[1], cells[2], cells[3], cells[4], cells[5]
		n = titleizeIfAllCaps(n)
		trait = titleizeIfAllCaps(trait)
		if n == "" {
			continue
		}
		featN, featT := "", ""
		if strings.Contains(featFull, ":") {
			p := strings.SplitN(featFull, ":", 2)
			featN = strings.Trim(p[0], "*_ ")
			featN = titleizeIfAllCaps(featN)
			featT = cleanInlineText(p[1])
		} else {
			featT = cleanInlineText(featFull)
		}
		if featT == "—" {
			featN, featT = "", ""
		}
		if damageType == "" {
			lower := strings.ToLower(dmg)
			if strings.Contains(lower, "mag") {
				damageType = "Magical"
			} else if strings.Contains(lower, "phy") {
				damageType = "Physical"
			}
		}
		if tier == "" || damageType == "" {
			warnMissing("Weapon", n, "Tier or Damage Type")
		}
		if trait == "" || rng == "" || dmg == "" || burden == "" {
			warnMissing("Weapon", n, "Trait/Range/Damage/Burden")
		}

		records = append(records, []string{n, category, tier, damageType, trait, rng, dmg, burden, featN, featT})
	}
	writeCSV(outputDir+"/weapons.csv", []string{"Name", "Primary or Secondary", "Tier", "Physical or Magical", "Trait", "Range", "Damage", "Burden", "Feature 1 Name", "Feature 1 Text"}, records)
}
