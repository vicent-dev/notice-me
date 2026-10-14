package notification

import "testing"

func TestNewNotificationMapsFields(t *testing.T) {
	n := NewNotification("body", "client", "group", true, "origin")

	if n.Body != "body" {
		t.Errorf("Body = %q, want %q", n.Body, "body")
	}
	if n.ClientId != "client" {
		t.Errorf("ClientId = %q, want %q", n.ClientId, "client")
	}
	if n.ClientGroupId != "group" {
		t.Errorf("ClientGroupId = %q, want %q", n.ClientGroupId, "group")
	}
	if !n.Instant {
		t.Errorf("Instant = false, want true")
	}
	if n.OriginClientId != "origin" {
		t.Errorf("OriginClientId = %q, want %q", n.OriginClientId, "origin")
	}
	if n.NotifiedAt != nil {
		t.Errorf("NotifiedAt = %v, want nil", n.NotifiedAt)
	}
}

func TestNewNotificationGeneratesUUID(t *testing.T) {
	a := NewNotification("body", "client", "group", false, "")
	b := NewNotification("body", "client", "group", false, "")

	if a.ID.String() == "" {
		t.Fatal("generated ID is empty")
	}
	if a.ID == b.ID {
		t.Fatalf("two notifications share the same ID: %s", a.ID)
	}
}

func TestFormatHTMLReturnsBody(t *testing.T) {
	n := NewNotification("<b>hello</b>", "client", "group", false, "")

	if got := n.FormatHTML(); got != "<b>hello</b>" {
		t.Errorf("FormatHTML() = %q, want %q", got, "<b>hello</b>")
	}
}
