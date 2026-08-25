// import_srd_2 imports SRD 2.0-only structured records from the authoritative
// PDF text extraction.  It intentionally complements rather than replaces the
// mature 1.0 extractor: the 2.0 PDF's tagged, two-column layout is not the
// Marker Markdown format consumed by extract_from_md.go.
package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const sourcePath = ".build/01_pdf/DH-SRD-2.0-2026-08-25.md"
const outputDir = ".build/02_csv"

func main() {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		panic(fmt.Errorf("read SRD 2.0 source: %w", err))
	}
	if err := importDreadAbilities(string(source)); err != nil {
		panic(err)
	}
	if err := importDomains(string(source)); err != nil {
		panic(err)
	}
	if err := importTransformations(string(source)); err != nil {
		panic(err)
	}
	if err := importExpansionSourceRecords(string(source)); err != nil {
		panic(err)
	}
	if err := writePageCatalog(string(source)); err != nil {
		panic(err)
	}
}

func importDreadAbilities(source string) error {
	start := strings.Index(source, "DREAD DOMAIN\nBLIGHTING STRIKE")
	end := strings.Index(source, "<!-- PDF page 215 -->")
	if start < 0 || end < 0 || end <= start {
		return fmt.Errorf("could not locate the SRD 2.0 Dread domain appendix")
	}
	section := source[start:end]
	re := regexp.MustCompile(`(?m)^([A-Z][A-Z -]+)\nLevel ([0-9]+) Dread (Spell|Ability)\nRecall Cost: ([0-9]+)\n`)
	matchIndexes := re.FindAllStringSubmatchIndex(section, -1)
	if len(matchIndexes) != 21 {
		return fmt.Errorf("expected 21 Dread cards, found %d", len(matchIndexes))
	}

	path := outputDir + "/abilities.csv"
	header, rows, err := readCSV(path)
	if err != nil {
		return err
	}
	filtered := make([][]string, 0, len(rows)+len(matchIndexes))
	for _, row := range rows {
		if len(row) > 2 && !strings.EqualFold(strings.TrimSpace(row[2]), "Dread") {
			filtered = append(filtered, row)
		}
	}
	dreadNames := []string{"Blighting Strike", "Umbral Veil", "Voice of Dread", "Hideous Retribution", "Siphon Essence", "Shared Trauma", "Terrify", "Chains of Affliction", "Summon Horror", "Dire Strike", "Spectral Mist", "Darkfire", "Jump Scare", "Dread-Touched", "Wall of Hunger", "Dark Army", "Eldritch Flesh", "Damnation", "Savor the Anguish", "Avatar of Terror", "Invoke Torment"}
	for i, match := range matchIndexes {
		bodyEnd := len(section)
		if i+1 < len(matchIndexes) {
			bodyEnd = matchIndexes[i+1][0]
		}
		filtered = append(filtered, []string{
			dreadNames[i], section[match[4]:match[5]], "Dread", section[match[6]:match[7]], section[match[8]:match[9]], cleanBody(section[match[1]:bodyEnd]),
		})
	}
	return writeCSV(path, header, filtered)
}

func importTransformations(source string) error {
	start := strings.Index(source, "<!-- PDF page 43 -->")
	end := strings.Index(source, "<!-- PDF page 46 -->")
	if start < 0 || end < 0 || end <= start {
		return fmt.Errorf("could not locate SRD 2.0 transformation pages")
	}
	section := source[start:end]
	names := []string{"DEMIGOD", "GHOST", "REANIMATED", "SHAPESHIFTER", "VAMPIRE", "WEREWOLF"}
	positions := make([]int, len(names))
	for i, name := range names {
		positions[i] = strings.Index(section, "\n"+name+"\n")
		if positions[i] < 0 {
			return fmt.Errorf("could not locate transformation %s", name)
		}
	}
	rows := make([][]string, 0, len(names))
	for i, name := range names {
		end := len(section)
		if i+1 < len(names) {
			end = positions[i+1]
		}
		block := section[positions[i]:end]
		description, features, questions, err := parseTransformation(block)
		if err != nil {
			return fmt.Errorf("parse %s: %w", name, err)
		}
		row := []string{titleCase(name), description}
		for _, feature := range features {
			row = append(row, feature[0], feature[1])
		}
		for len(row) < 14 { // six feature columns (three name/text pairs)
			row = append(row, "")
		}
		for _, question := range questions {
			row = append(row, question)
		}
		for len(row) < 20 { // six question columns
			row = append(row, "")
		}
		rows = append(rows, row)
	}
	header := []string{
		"Name", "Description",
		"Feature 1 Name", "Feature 1 Text", "Feature 2 Name", "Feature 2 Text", "Feature 3 Name", "Feature 3 Text",
		"Feature 4 Name", "Feature 4 Text", "Feature 5 Name", "Feature 5 Text", "Feature 6 Name", "Feature 6 Text",
		"Question 1 Text", "Question 2 Text", "Question 3 Text", "Question 4 Text", "Question 5 Text", "Question 6 Text",
	}
	return writeCSV(outputDir+"/transformations.csv", header, rows)
}

func importDomains(source string) error {
	start := strings.Index(source, "<!-- PDF page 7 -->")
	end := strings.Index(source, "<!-- PDF page 8 -->")
	if start < 0 || end < 0 || end <= start {
		return fmt.Errorf("could not locate the SRD 2.0 domains overview")
	}
	section := source[start:end]
	names := []string{"Arcana", "Blade", "Bone", "Codex", "Dread", "Grace", "Midnight", "Sage", "Splendor", "Valor"}
	positions := map[string]int{}
	for _, name := range names {
		position := strings.Index(section, "\n"+strings.ToUpper(name)+"\n")
		if position < 0 {
			return fmt.Errorf("could not locate domain %s", name)
		}
		positions[name] = position + 1
	}
	header, rows, err := readCSV(outputDir + "/domains.csv")
	if err != nil {
		return err
	}
	for _, name := range names {
		start := positions[name] + len(strings.ToUpper(name)) + 1
		finish := len(section)
		for _, candidate := range names {
			if positions[candidate] > positions[name] && positions[candidate] < finish {
				finish = positions[candidate]
			}
		}
		row := make([]string, len(header))
		row[0] = name
		row[1] = cleanBody(section[start:finish])
		if name == "Dread" {
			cardsByLevel := map[int][]string{}
			abilitiesHeader, abilities, err := readCSV(outputDir + "/abilities.csv")
			if err != nil {
				return err
			}
			nameIndex, levelIndex, domainIndex := csvColumn(abilitiesHeader, "Name"), csvColumn(abilitiesHeader, "Level"), csvColumn(abilitiesHeader, "Domain")
			for _, ability := range abilities {
				if domainIndex >= 0 && domainIndex < len(ability) && ability[domainIndex] == "Dread" {
					level := 0
					fmt.Sscanf(ability[levelIndex], "%d", &level)
					cardsByLevel[level] = append(cardsByLevel[level], ability[nameIndex])
				}
			}
			for level := 1; level <= 10; level++ {
				for option, card := range cardsByLevel[level] {
					index := 2 + (level-1)*3 + option
					if index < len(row) {
						row[index] = card
					}
				}
			}
		} else {
			for _, existing := range rows {
				if len(existing) > 0 && existing[0] == name {
					copy(row[2:], existing[2:])
					break
				}
			}
		}
		rows = replaceNamedRow(rows, name, row)
	}
	return writeCSV(outputDir+"/domains.csv", header, rows)
}

func csvColumn(header []string, want string) int {
	for i, field := range header {
		if field == want {
			return i
		}
	}
	return -1
}

func parseTransformation(block string) (string, [][2]string, []string, error) {
	featuresAt := strings.Index(block, "TRANSFORMATION FEATURES\n")
	questionsAt := strings.Index(block, "TRANSFORMATION QUESTIONS\n")
	if featuresAt < 0 || questionsAt < 0 || questionsAt <= featuresAt {
		return "", nil, nil, fmt.Errorf("missing feature or question section")
	}
	nameEnd := strings.Index(block[1:], "\n") + 2
	description := cleanBody(block[nameEnd:featuresAt])
	featureText := cleanBody(block[featuresAt+len("TRANSFORMATION FEATURES\n") : questionsAt])
	questionsText := block[questionsAt+len("TRANSFORMATION QUESTIONS\n"):]
	featureRE := regexp.MustCompile(`(?m)^([^:\n]+):\s*`)
	featureMatches := featureRE.FindAllStringSubmatchIndex(featureText, -1)
	if len(featureMatches) < 2 {
		return "", nil, nil, fmt.Errorf("expected at least two features")
	}
	features := make([][2]string, 0, len(featureMatches))
	for i, match := range featureMatches {
		textEnd := len(featureText)
		if i+1 < len(featureMatches) {
			textEnd = featureMatches[i+1][0]
		}
		features = append(features, [2]string{strings.TrimSpace(featureText[match[2]:match[3]]), cleanBody(featureText[match[1]:textEnd])})
	}
	questions := []string{}
	for _, line := range strings.Split(questionsText, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "•") {
			questions = append(questions, strings.TrimSpace(strings.TrimPrefix(trimmed, "•")))
			continue
		}
		if trimmed != "" && len(questions) > 0 && !strings.HasPrefix(trimmed, "<!--") && !strings.HasSuffix(trimmed, "Daggerheart SRD") && !regexp.MustCompile(`^[0-9]+$`).MatchString(trimmed) {
			questions[len(questions)-1] += " " + trimmed
		}
	}
	if len(questions) != 6 {
		return "", nil, nil, fmt.Errorf("expected six questions, found %d", len(questions))
	}
	return description, features, questions, nil
}

// importExpansionSourceRecords appends SRD 2.0's new character options without
// changing the legacy columns consumers already use. The complete source block
// is stored in Description; this preserves every word of the authoritative PDF
// while the standard fields retain the values that existing integrations expect.
func importExpansionSourceRecords(source string) error {
	if err := upsertSourceBlocks(source, "classes.csv", []string{
		"Assassin", "Bard", "Brawler", "Druid", "Guardian", "Ranger", "Rogue", "Seraph", "Sorcerer", "Warlock", "Warrior", "Witch", "Wizard",
	}, map[string]map[string]string{
		"Assassin": {"Domain 1": "Blade", "Domain 2": "Midnight", "Evasion": "12", "HP": "5", "Subclass 1": "Executioners Guild", "Subclass 2": "Poisoners Guild"},
		"Brawler":  {"Domain 1": "Valor", "Domain 2": "Bone", "Evasion": "10", "HP": "6", "Subclass 1": "Juggernaut", "Subclass 2": "Martial Artist"},
		"Warlock":  {"Domain 1": "Dread", "Domain 2": "Grace", "Evasion": "11", "HP": "5", "Subclass 1": "Pact of the Endless", "Subclass 2": "Pact of the Wrathful"},
		"Witch":    {"Domain 1": "Dread", "Domain 2": "Sage", "Evasion": "10", "HP": "6", "Subclass 1": "Hedge", "Subclass 2": "Moon"},
	}, []string{"Assassin", "Brawler", "Warlock", "Witch"}); err != nil {
		return err
	}
	if err := upsertSourceBlocks(source, "ancestries.csv", []string{
		"Aetheris", "Clank", "Drakona", "Dwarf", "Earthkin", "Elf", "Emberkin", "Faerie", "Faun", "Firbolg", "Fungril", "Galapa", "Giant", "Gnome", "Goblin", "Halfling", "Human", "Infernis", "Katari", "Orc", "Ribbet", "Simiah", "Skykin", "Tidekin",
	}, nil, []string{"Aetheris", "Earthkin", "Emberkin", "Gnome", "Skykin", "Tidekin"}); err != nil {
		return err
	}
	if err := upsertSourceBlocks(source, "communities.csv", []string{
		"Duneborne", "Freeborne", "Frostborne", "Hearthborne", "Highborne", "Loreborne", "Orderborne", "Reborne", "Ridgeborne", "Seaborne", "Slyborne", "Underborne", "Wanderborne", "Warborne", "Wildborne",
	}, nil, []string{"Duneborne", "Freeborne", "Frostborne", "Hearthborne", "Reborne", "Warborne"}); err != nil {
		return err
	}
	return upsertStaticDescriptions("subclasses.csv", map[string]string{
		"Executioners Guild":   sourceSegment(source, "EXECUTIONERS GUILD\n", "\nPOISONERS GUILD\n"),
		"Poisoners Guild":      sourceSegment(source, "POISONERS GUILD\n", "\nBACKGROUND QUESTIONS\n"),
		"Juggernaut":           sourceSegment(source, "JUGGERNAUT\n", "\nMARTIAL ARTIST\n"),
		"Martial Artist":       sourceSegment(source, "MARTIAL ARTIST\n", "\nBACKGROUND QUESTIONS\n"),
		"Pact of the Endless":  sourceSegment(source, "PACT OF THE ENDLESS\n", "\nPACT OF THE WRATHFUL\n"),
		"Pact of the Wrathful": sourceSegment(source, "PACT OF THE WRATHFUL\n", "\nBACKGROUND QUESTIONS\n"),
		"Hedge":                sourceSegment(source, "HEDGE\n", "\nMOON\n"),
		"Moon":                 sourceSegment(source, "MOON\n", "\nBACKGROUND QUESTIONS\n"),
	})
}

func upsertSourceBlocks(source, filename string, boundaries []string, values map[string]map[string]string, additions []string) error {
	path := outputDir + "/" + filename
	header, rows, err := readCSV(path)
	if err != nil {
		return err
	}
	positions := map[string]int{}
	for _, name := range boundaries {
		needle := "\n" + strings.ToUpper(name) + "\n"
		position := strings.Index(source, needle)
		if position < 0 {
			return fmt.Errorf("could not locate %s in SRD 2.0 source", name)
		}
		positions[name] = position + 1
	}
	for _, name := range additions {
		start := positions[name]
		end := len(source)
		for _, candidate := range boundaries {
			if positions[candidate] > start && positions[candidate] < end {
				end = positions[candidate]
			}
		}
		content := cleanBody(source[start:end])
		row := make([]string, len(header))
		row[0] = name
		for i, field := range header {
			if field == "Description" {
				row[i] = content
			}
			if values != nil && values[name] != nil {
				if value, ok := values[name][field]; ok {
					row[i] = value
				}
			}
		}
		rows = replaceNamedRow(rows, name, row)
	}
	return writeCSV(path, header, rows)
}

func replaceNamedRow(rows [][]string, name string, replacement []string) [][]string {
	out := make([][]string, 0, len(rows)+1)
	for _, row := range rows {
		if len(row) > 0 && strings.EqualFold(strings.TrimSpace(row[0]), name) {
			continue
		}
		out = append(out, row)
	}
	return append(out, replacement)
}

func upsertStaticDescriptions(filename string, additions map[string]string) error {
	path := outputDir + "/" + filename
	header, rows, err := readCSV(path)
	if err != nil {
		return err
	}
	for name, description := range additions {
		row := make([]string, len(header))
		row[0] = name
		for i, field := range header {
			if field == "Description" {
				row[i] = description
			}
		}
		rows = replaceNamedRow(rows, name, row)
	}
	return writeCSV(path, header, rows)
}

func sourceSegment(source, startText, endText string) string {
	start := strings.Index(source, startText)
	if start < 0 {
		panic(fmt.Sprintf("could not locate SRD 2.0 source segment %q", strings.TrimSpace(startText)))
	}
	endRelative := strings.Index(source[start+len(startText):], endText)
	if endRelative < 0 {
		panic(fmt.Sprintf("could not locate end of SRD 2.0 source segment %q", strings.TrimSpace(startText)))
	}
	return cleanBody(source[start : start+len(startText)+endRelative])
}

// writePageCatalog provides a lossless machine-readable bridge for every PDF
// page, including equipment, adversaries, environments, and campaign material
// whose 2.0 layout no longer matches the legacy column extractor.
func writePageCatalog(source string) error {
	marker := regexp.MustCompile(`(?m)^<!-- PDF page ([0-9]+) -->\n`)
	matches := marker.FindAllStringSubmatchIndex(source, -1)
	if len(matches) != 224 {
		return fmt.Errorf("expected 224 SRD 2.0 pages, found %d", len(matches))
	}
	rows := make([][]string, 0, len(matches))
	for i, match := range matches {
		end := len(source)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		rows = append(rows, []string{source[match[2]:match[3]], cleanBody(source[match[1]:end])})
	}
	return writeCSV(outputDir+"/srd_2_content.csv", []string{"PDF Page", "Text"}, rows)
}

func readCSV(path string) ([]string, [][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, nil, err
	}
	rows, err := r.ReadAll()
	return header, rows, err
}

func writeCSV(path string, header []string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.Write(header); err != nil {
		return err
	}
	if err := w.WriteAll(rows); err != nil {
		return err
	}
	return w.Error()
}

func cleanBody(value string) string {
	value = strings.ReplaceAll(value, "\n\n", "\n")
	return strings.TrimSpace(value)
}

func titleCase(value string) string {
	words := strings.Fields(strings.ToLower(value))
	for i, word := range words {
		if word == "of" || word == "the" || word == "and" {
			continue
		}
		segments := strings.Split(word, "-")
		for j, segment := range segments {
			if segment != "" {
				segments[j] = strings.ToUpper(segment[:1]) + segment[1:]
			}
		}
		words[i] = strings.Join(segments, "-")
	}
	return strings.Join(words, " ")
}
