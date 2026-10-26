package checkout

import "testing"

func TestCheckRefName(t *testing.T) {
	for _, bad := range []string{"", "-b", "--upload-pack=x", "has space", "a\tb", "a\nb"} {
		if err := CheckRefName(bad); err == nil {
			t.Errorf("CheckRefName(%q) accepted, want refused", bad)
		}
	}
	for _, ok := range []string{"main", "feature/x", "v1.2.0", "release-1"} {
		if err := CheckRefName(ok); err != nil {
			t.Errorf("CheckRefName(%q) = %v, want accepted", ok, err)
		}
	}
}
