package theme

import "testing"

func TestGetDefault(t *testing.T) {
	th := Get("default")
	if th == nil {
		t.Fatal("default theme is nil")
	}
	if th.Name != "default" {
		t.Errorf("expected name 'default', got %q", th.Name)
	}
	if len(th.Colors) == 0 {
		t.Error("expected colors, got empty")
	}
}

func TestGetAllThemes(t *testing.T) {
	for _, name := range Names() {
		th := Get(name)
		if th == nil {
			t.Errorf("theme %q is nil", name)
			continue
		}
		if th.Name != name {
			t.Errorf("expected name %q, got %q", name, th.Name)
		}
		if len(th.Colors) < 4 {
			t.Errorf("theme %q has too few colors: %d", name, len(th.Colors))
		}
	}
}

func TestGetFallback(t *testing.T) {
	th := Get("nonexistent")
	if th == nil {
		t.Fatal("fallback theme is nil")
	}
	if th.Name != "default" {
		t.Errorf("expected fallback to 'default', got %q", th.Name)
	}
}

func TestNext(t *testing.T) {
	names := Names()
	first := Get(names[0])
	next := Next(names[0])
	if next.Name == first.Name {
		t.Error("Next should return a different theme")
	}
	if next.Name != names[1] {
		t.Errorf("expected %q after %q, got %q", names[1], names[0], next.Name)
	}
}

func TestNextWraps(t *testing.T) {
	names := Names()
	last := names[len(names)-1]
	next := Next(last)
	if next.Name != names[0] {
		t.Errorf("expected wrap to %q, got %q", names[0], next.Name)
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) < 5 {
		t.Errorf("expected at least 5 themes, got %d", len(names))
	}
}
