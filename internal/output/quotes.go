package output

import (
	"regexp"
	"strings"
)

// Quote attribution patterns for various email clients
var quoteAttributionPatterns = []*regexp.Regexp{
	// Gmail: "On Mon, Jan 1, 2025 at 10:00 AM Name <email@example.com> wrote:"
	regexp.MustCompile(`(?im)^On .+wrote:\s*$`),

	// Apple Mail: "On 8 Dec 2025, at 14:30, other@example.com wrote:"
	regexp.MustCompile(`(?im)^On .+, at .+wrote:\s*$`),

	// Norwegian with weekday: "man. 1. des. 2025 kl. 14:30 skrev Name <email>:"
	// Also handles abbreviated weekdays: man., tir., ons., tor., fre., lør., søn.
	regexp.MustCompile(`(?im)^[a-zæøå]{2,4}\.\s+\d+\..+kl\..+skrev`),

	// Norwegian without weekday: "8. des. 2025 kl. 14:30 skrev Name <email>:"
	regexp.MustCompile(`(?im)^\d+\..+kl\..+skrev`),

	// Generic "wrote:" pattern
	regexp.MustCompile(`(?im)^.+wrote:\s*$`),

	// Outlook separator line (underscores)
	regexp.MustCompile(`(?m)^_{10,}\s*$`),

	// Outlook "From:" header block start
	regexp.MustCompile(`(?im)^From:\s+.+$`),

	// "Original Message" markers
	regexp.MustCompile(`(?im)^-+\s*Original Message\s*-+\s*$`),

	// "Forwarded message" markers
	regexp.MustCompile(`(?im)^-+\s*Forwarded message\s*-+\s*$`),
}

// StripQuotedContent removes quoted reply content from an email body,
// preserving only the new content written by the sender.
func StripQuotedContent(body string) string {
	lines := strings.Split(body, "\n")
	var result []string
	inQuotedBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this line starts a quoted block
		if !inQuotedBlock && startsQuotedBlock(trimmed, lines, i) {
			inQuotedBlock = true
			continue
		}

		// Skip lines that are part of quoted content
		if inQuotedBlock {
			continue
		}

		// Skip lines that start with ">" (inline quotes)
		if strings.HasPrefix(trimmed, ">") {
			continue
		}

		result = append(result, line)
	}

	// Trim trailing empty lines
	for len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "" {
		result = result[:len(result)-1]
	}

	output := strings.Join(result, "\n")
	output = strings.TrimSpace(output)

	// If everything was quoted, return a placeholder
	if output == "" {
		return "[No new content - forwarded/quoted message only]"
	}

	return output
}

// startsQuotedBlock checks if the current line marks the beginning of a quoted block.
func startsQuotedBlock(line string, lines []string, currentIdx int) bool {
	// Check against known attribution patterns
	for _, pattern := range quoteAttributionPatterns {
		if pattern.MatchString(line) {
			// For "From:" pattern, also check if followed by typical Outlook headers
			if strings.HasPrefix(strings.ToLower(line), "from:") {
				return isOutlookHeaderBlock(lines, currentIdx)
			}
			return true
		}
	}

	return false
}

// isOutlookHeaderBlock checks if we're at the start of an Outlook-style quoted block.
// Outlook quotes typically have: From:, Sent:, To:, Subject: headers in sequence.
func isOutlookHeaderBlock(lines []string, startIdx int) bool {
	// Look for typical Outlook header sequence within next few lines
	headerCount := 0
	outlookHeaders := []string{"from:", "sent:", "to:", "subject:", "date:"}

	for i := startIdx; i < len(lines) && i < startIdx+6; i++ {
		lower := strings.ToLower(strings.TrimSpace(lines[i]))
		for _, header := range outlookHeaders {
			if strings.HasPrefix(lower, header) {
				headerCount++
				break
			}
		}
	}

	// If we found at least 3 Outlook-style headers, it's likely a quoted block
	return headerCount >= 3
}
