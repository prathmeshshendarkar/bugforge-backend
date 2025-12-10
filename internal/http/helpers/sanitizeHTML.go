package helpers

import (
	"github.com/microcosm-cc/bluemonday"
)

func SanitizeHTML(input string) string {
    // Create a policy allowing basic formatting (bold, italic, links, lists, etc.)
    policy := bluemonday.UGCPolicy()

    // Sanitize input HTML
    safe := policy.Sanitize(input)

    return safe
}
