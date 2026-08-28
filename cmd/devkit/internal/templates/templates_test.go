package templates

import "testing"

// The registry must not contain duplicates or empty fields.
func TestRegistry(t *testing.T) {
	seen := map[string]bool{}
	for _, tmpl := range All {
		if tmpl.Name == "" || tmpl.Subdir == "" || tmpl.Lang == "" {
			t.Errorf("incomplete template entry: %+v", tmpl)
		}
		if seen[tmpl.Name] {
			t.Errorf("duplicate template name %q", tmpl.Name)
		}
		seen[tmpl.Name] = true
		if _, err := Find(tmpl.Name); err != nil {
			t.Errorf("Find(%q): %v", tmpl.Name, err)
		}
	}
	if _, err := Find("does-not-exist"); err == nil {
		t.Error("Find with unknown name should fail")
	}
}
