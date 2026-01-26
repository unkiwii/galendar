package galendar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

func LoadSpecialDaysFromFile(filename string, cfg Config) (SpecialDays, error) {
	if filename == "" {
		return nil, nil
	}

	var file specialDaysTomlFile

	_, err := toml.DecodeFile(filename, &file)
	if err != nil {
		return nil, fmt.Errorf("can't decode toml file %q: %w", filename, err)
	}

	days := SpecialDays{}
	for _, day := range file.Day {
		key, err := specialDaysKeyFromString(file.DateFormat, day.When, cfg)
		if err != nil {
			return nil, fmt.Errorf("invalid 'when' value %q: %w", day.When, err)
		}

		// Create the date for this special day (using calendar year)
		date := time.Date(cfg.Year, time.Month(key.month), key.day, 0, 0, 0, 0, time.UTC)

		// Evaluate expressions in string properties
		// We need to check if any expression evaluates to ≤ 0 to skip the day
		evaluatedText, shouldSkip, err := evaluateExpressionsWithSkip(day.Text, cfg, date)
		if err != nil {
			return nil, fmt.Errorf("error evaluating text for day %q: %w", day.When, err)
		}
		if shouldSkip {
			continue
		}

		evaluatedIcon, shouldSkip, err := evaluateExpressionsWithSkip(day.Icon, cfg, date)
		if err != nil {
			return nil, fmt.Errorf("error evaluating icon for day %q: %w", day.When, err)
		}
		if shouldSkip {
			continue
		}

		evaluatedFont, shouldSkip, err := evaluateExpressionsWithSkip(day.Font, cfg, date)
		if err != nil {
			return nil, fmt.Errorf("error evaluating font for day %q: %w", day.When, err)
		}
		if shouldSkip {
			continue
		}

		specialDay := SpecialDay{
			Date:    date,
			Holiday: day.Holiday,
			Icon:    evaluatedIcon,
			Note: SpecialDayNote{
				Text: evaluatedText,
				Font: evaluatedFont,
				Size: day.Size,
			},
		}

		days[key] = specialDay
	}

	return days, nil
}

type specialDaysTomlFile struct {
	DateFormat string `toml:"date_format"`
	Day        []struct {
		When    string
		Holiday bool
		Icon    string
		Text    string
		Font    string
		Size    float64
	}
}

type specialDaysKey struct {
	month int
	day   int
}

func (key specialDaysKey) String() string {
	return fmt.Sprintf("%d/%d", key.month, key.day)
}

func specialDaysKeyFromString(layout, s string, cfg Config) (specialDaysKey, error) {
	// Check if it's a relative date pattern: ((ordinal weekday))/month
	if key, err := parseRelativeDate(s, cfg); err == nil {
		return key, nil
	}

	// Try to parse as fixed date
	t, err := time.Parse(layout, s)
	if err != nil {
		return specialDaysKey{}, fmt.Errorf("can't parse %q as %q or relative date: %w", s, layout, err)
	}

	return specialDaysKeyFromTime(t), nil
}

func specialDaysKeyFromTime(t time.Time) specialDaysKey {
	return specialDaysKey{
		month: int(t.Month()),
		day:   t.Day(),
	}
}

// evaluateExpressionsWithSkip finds and evaluates all ((expression)) patterns in a string
// Returns the evaluated string, a boolean indicating if the day should be skipped (expression ≤ 0), and an error
func evaluateExpressionsWithSkip(text string, cfg Config, date time.Time) (string, bool, error) {
	if text == "" {
		return text, false, nil
	}

	// Pattern to match ((...))
	pattern := regexp.MustCompile(`\(\(([^)]+)\)\)`)
	shouldSkip := false

	// Find all matches first to check values
	matches := pattern.FindAllStringSubmatch(text, -1)
	values := make(map[string]int)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		expr := match[1] // The expression inside (())

		value, err := evaluateArithmetic(expr, cfg, date)
		if err != nil {
			return "", false, fmt.Errorf("expression evaluation error: %w", err)
		}

		// Check if the result is ≤ 0 - if so, mark this day to be skipped
		if value <= 0 {
			shouldSkip = true
		}

		// Store the value for replacement
		values[match[0]] = value
	}

	// Now replace all matches with their values
	result := pattern.ReplaceAllStringFunc(text, func(match string) string {
		if value, ok := values[match]; ok {
			return strconv.Itoa(value)
		}
		return match // Should not happen, but fallback
	})

	return result, shouldSkip, nil
}

// evaluateArithmetic evaluates a simple arithmetic expression with + and - operators
func evaluateArithmetic(expr string, cfg Config, date time.Time) (int, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, fmt.Errorf("empty expression")
	}

	// Parse the expression by splitting on + and - while preserving operators
	// We'll use a simple tokenizer approach
	tokens := tokenizeExpression(expr)
	if len(tokens) == 0 {
		return 0, fmt.Errorf("no tokens in expression")
	}

	// Evaluate left-to-right (no operator precedence for + and -)
	result, err := resolveValue(tokens[0], cfg, date)
	if err != nil {
		return 0, err
	}

	for i := 1; i < len(tokens); i += 2 {
		if i+1 >= len(tokens) {
			return 0, fmt.Errorf("incomplete expression: missing operand after operator")
		}

		operator := tokens[i]
		operand, err := resolveValue(tokens[i+1], cfg, date)
		if err != nil {
			return 0, err
		}

		switch operator {
		case "+":
			result += operand
		case "-":
			result -= operand
		default:
			return 0, fmt.Errorf("unsupported operator: %q (only + and - are supported)", operator)
		}
	}

	return result, nil
}

// tokenizeExpression splits an expression into tokens (values and operators)
func tokenizeExpression(expr string) []string {
	var tokens []string
	var current strings.Builder
	expr = strings.TrimSpace(expr)

	for _, char := range expr {
		switch char {
		case '+', '-':
			// If we have accumulated a token, add it
			if current.Len() > 0 {
				tokens = append(tokens, strings.TrimSpace(current.String()))
				current.Reset()
			}
			// Handle unary minus at the start or after an operator
			if char == '-' && (len(tokens) == 0 || tokens[len(tokens)-1] == "+" || tokens[len(tokens)-1] == "-") {
				current.WriteRune(char)
			} else {
				tokens = append(tokens, string(char))
			}
		case ' ':
			// Skip spaces, but if we have content, it's part of the current token
			if current.Len() > 0 {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the last token
	if current.Len() > 0 {
		tokens = append(tokens, strings.TrimSpace(current.String()))
	}

	return tokens
}

// resolveValue resolves a token to an integer value
// It can be a variable name (year, month, day, cfg.year, cfg.month) or a number
func resolveValue(token string, cfg Config, date time.Time) (int, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, fmt.Errorf("empty token")
	}

	// Try to parse as a number first
	if num, err := strconv.Atoi(token); err == nil {
		return num, nil
	}

	// Handle unary minus
	if strings.HasPrefix(token, "-") {
		value, err := resolveValue(token[1:], cfg, date)
		if err != nil {
			return 0, err
		}
		return -value, nil
	}

	// Handle unary plus
	if strings.HasPrefix(token, "+") {
		return resolveValue(token[1:], cfg, date)
	}

	// Resolve as a variable
	tokenLower := strings.ToLower(token)

	// Check for cfg. prefix
	if after, ok := strings.CutPrefix(tokenLower, "cfg."); ok {
		prop := after
		switch prop {
		case "year":
			return cfg.Year, nil
		case "month":
			return cfg.Month, nil
		default:
			return 0, fmt.Errorf("unknown config property: %q", prop)
		}
	}

	// Resolve date properties (year, month, day)
	switch tokenLower {
	case "year":
		return date.Year(), nil
	case "month":
		return int(date.Month()), nil
	case "day":
		return date.Day(), nil
	default:
		return 0, fmt.Errorf("unknown variable: %q (supported: year, month, day, cfg.year, cfg.month)", token)
	}
}

// parseRelativeDate parses a relative date pattern like "((3rd sunday))/10"
// Returns a specialDaysKey if successful, or an error if it's not a relative date pattern
func parseRelativeDate(s string, cfg Config) (specialDaysKey, error) {
	// Pattern: ((ordinal weekday))/month
	// Example: ((3rd sunday))/10
	pattern := regexp.MustCompile(`^\(\((.+)\)\)/(\d+)$`)
	matches := pattern.FindStringSubmatch(s)
	if len(matches) != 3 {
		return specialDaysKey{}, fmt.Errorf("not a relative date pattern")
	}

	ordinalWeekday := strings.TrimSpace(matches[1])
	monthStr := matches[2]

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return specialDaysKey{}, fmt.Errorf("invalid month in relative date: %q", monthStr)
	}
	if month < 1 || month > 12 {
		return specialDaysKey{}, fmt.Errorf("month out of range: %d (must be 1-12)", month)
	}

	// Parse ordinal and weekday
	ordinal, weekday, err := parseOrdinalWeekday(ordinalWeekday)
	if err != nil {
		return specialDaysKey{}, fmt.Errorf("invalid ordinal/weekday in relative date: %w", err)
	}

	// Calculate the actual date
	day, err := calculateOrdinalWeekdayDate(cfg, month, ordinal, weekday)
	if err != nil {
		return specialDaysKey{}, fmt.Errorf("failed to calculate date: %w", err)
	}

	return specialDaysKey{
		month: month,
		day:   day,
	}, nil
}

// parseOrdinalWeekday parses strings like "3rd sunday", "last monday", "1st friday"
func parseOrdinalWeekday(s string) (int, time.Weekday, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Check for "last"
	if after, ok := strings.CutPrefix(s, "last "); ok {
		weekdayStr := after
		weekday, err := ParseWeekday(weekdayStr)
		if err != nil {
			return 0, 0, err
		}
		return -1, weekday, nil // -1 means "last"
	}

	// Parse ordinal (1st, 2nd, 3rd, 4th)
	ordinalMap := map[string]int{
		"1st":    1,
		"2nd":    2,
		"3rd":    3,
		"4th":    4,
		"first":  1,
		"second": 2,
		"third":  3,
		"fourth": 4,
	}

	for ordinalStr, ordinal := range ordinalMap {
		if after, ok := strings.CutPrefix(s, ordinalStr+" "); ok {
			weekdayStr := after
			weekday, err := ParseWeekday(weekdayStr)
			if err != nil {
				return 0, 0, err
			}
			return ordinal, weekday, nil
		}
	}

	return 0, 0, fmt.Errorf("invalid ordinal/weekday format: %q (expected format: '1st sunday', 'last monday', etc.)", s)
}

// calculateOrdinalWeekdayDate calculates the day of month for an ordinal weekday
// ordinal: 1-4 for 1st, 2nd, 3rd, 4th, or -1 for "last"
func calculateOrdinalWeekdayDate(cfg Config, month, ordinal int, weekday time.Weekday) (int, error) {
	// Get the first day of the month
	firstDay := time.Date(cfg.Year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	firstWeekday := firstDay.Weekday()

	// Calculate days until the first occurrence of the target weekday
	daysUntilFirst := int(weekday-firstWeekday+7) % 7

	if ordinal == -1 {
		// Find the last occurrence
		// Get the last day of the month
		lastDay := firstDay.AddDate(0, 1, -1)
		lastWeekday := lastDay.Weekday()

		// Calculate days back from the last day
		daysBack := int(lastWeekday-weekday+7) % 7
		lastOccurrence := lastDay.Day() - daysBack

		if lastOccurrence < 1 {
			return 0, fmt.Errorf("no occurrence of %v in month %d", weekday, month)
		}

		return lastOccurrence, nil
	}

	// Calculate the date for the nth occurrence (1st, 2nd, 3rd, 4th)
	// First occurrence is at: 1 + daysUntilFirst
	// Nth occurrence is at: 1 + daysUntilFirst + (ordinal-1)*7
	targetDay := 1 + daysUntilFirst + (ordinal-1)*7

	// Verify the date is still in the same month
	if targetDay > 31 {
		return 0, fmt.Errorf("ordinal %d %v does not exist in month %d", ordinal, weekday, month)
	}

	// Verify it's actually the correct weekday
	testDate := time.Date(cfg.Year, time.Month(month), targetDay, 0, 0, 0, 0, time.UTC)
	if testDate.Weekday() != weekday {
		return 0, fmt.Errorf("calculated date %d is not a %v", targetDay, weekday)
	}

	// Verify the date is within the month (handles months with < 31 days)
	if int(testDate.Month()) != month {
		return 0, fmt.Errorf("ordinal %d %v does not exist in month %d", ordinal, weekday, month)
	}

	return targetDay, nil
}
