package main

import "testing"

// A fresh machine has no saved settings and no clone; State must say so
// instead of inventing values the pages would then show as saved.
func TestState_Defaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	st := NewApp().State()

	if st.Settings != (Settings{}) {
		t.Errorf("Settings = %+v, want empty", st.Settings)
	}
	if st.Cloned {
		t.Error("Cloned = true, want false before any clone")
	}
	if st.Env.DefaultKiroHome == "" {
		t.Error("Env.DefaultKiroHome is empty, want the folder the Settings page offers")
	}
}
