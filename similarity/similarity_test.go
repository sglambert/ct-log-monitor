package similarity

import (
	"testing"
)

// Testing the replaceHomoglyphs function
func TestReplaceHomoglyphs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello World"},                          // No homoglyphs
		{"аpple", "apple"},                                      // Cyrillic 'a'
		{"gооgle", "google"},                                    // Cyrillic 'o'
		{"fаcebооk", "facebook"},                                // Multiple homoglyphs
		{"Microѕoft", "Microsoft"},                              // Greek 'sigma'
		{"gтаmail", "gtamail"},                                  // Latin 't' lookalike
		{"", ""},                                                // Empty string
		{"MixedCASE", "MixedCASE"},                              // Case sensitivity (should retain)
		{"domain.com", "domain.com"},                            // Dots in domains
		{"apple.com", "apple.com"},
		{"аррӏе.сом", "apple.com"}, 							 // Mix of different homoglyphs
	}

	for _, tt := range tests {
		t.Run("Homoglyph: "+tt.input, func(t *testing.T) {
			result := replaceHomoglyphs(tt.input)
			if result != tt.expected {
				t.Errorf("replaceHomoglyphs(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
