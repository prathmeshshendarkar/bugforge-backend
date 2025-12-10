package helpers

import "regexp"

// Extracts @username mentions from raw comment body
func ExtractMentions(body string) []string {
    re := regexp.MustCompile(`@([a-zA-Z0-9._-]+)`)
    matches := re.FindAllStringSubmatch(body, -1)

    out := []string{}
    for _, m := range matches {
        out = append(out, m[1]) // the username portion
    }
    return out
}
