package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Szotasz/connectors-cli/internal/api"
)

func testManifest() *api.Manifest {
	return &api.Manifest{
		Connectors: []api.ConnectorInfo{
			{ID: "nav", Name: "NAV Online Invoice"},
			{ID: "google_workspace", Name: "Google Workspace"},
		},
		Tools: []api.ToolEntry{
			{Connector: "nav", Command: "query_taxpayer", Description: "Look up a taxpayer.",
				Args: []api.Arg{{Name: "taxNumber", Type: "string", Required: true}}},
			{Connector: "google_workspace", Command: "list_accounts", Description: "List accounts."},
		},
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func TestSkillIsSplitIntoIndexAndReferences(t *testing.T) {
	dir := t.TempDir()
	if err := updateSkillAt(dir, testManifest()); err != nil {
		t.Fatal(err)
	}
	skill := read(t, filepath.Join(dir, "SKILL.md"))
	if strings.Contains(skill, "### connectors") {
		t.Error("SKILL.md must not carry per-command sections; they belong in references/")
	}
	for _, want := range []string{
		"- [NAV Online Invoice](references/nav.md) -- `connectors nav` (1): query_taxpayer",
		"- [Google Workspace](references/google-workspace.md) -- `connectors google_workspace` (1): list_accounts",
		LocalNotesStart, LocalNotesEnd,
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("SKILL.md missing %q", want)
		}
	}
	nav := read(t, filepath.Join(dir, "references", "nav.md"))
	if !strings.HasPrefix(nav, GeneratedMarker) {
		t.Error("a generated reference must start with the marker")
	}
	if !strings.Contains(nav, "### connectors nav query_taxpayer") || !strings.Contains(nav, "--taxNumber  string (required)") {
		t.Errorf("nav.md lacks the command section:\n%s", nav)
	}
}

func TestLocalNotesSurviveASync(t *testing.T) {
	dir := t.TempDir()
	if err := updateSkillAt(dir, testManifest()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "SKILL.md")
	notes := "> Use the absolute path /Users/x/.local/bin/connectors.\n"
	edited := strings.Replace(read(t, path), LocalNotesStart+"\n", LocalNotesStart+"\n"+notes, 1)
	if err := os.WriteFile(path, []byte(edited), 0644); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ { // twice: the notes must not double up or drift
		if err := updateSkillAt(dir, testManifest()); err != nil {
			t.Fatal(err)
		}
	}
	skill := read(t, path)
	if strings.Count(skill, notes) != 1 {
		t.Errorf("local notes not kept exactly once:\n%s", skill)
	}
	if _, err := os.Stat(filepath.Join(dir, unmarkedBackupName)); !os.IsNotExist(err) {
		t.Error("a marked SKILL.md must not be backed up")
	}
}

func TestStaleGeneratedReferenceIsRemovedButForeignFilesStay(t *testing.T) {
	dir := t.TempDir()
	if err := updateSkillAt(dir, testManifest()); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(dir, "references", "my-notes.md")
	if err := os.WriteFile(foreign, []byte("# hand written\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := testManifest()
	m.Connectors = m.Connectors[:1] // google_workspace dropped from the manifest
	if err := updateSkillAt(dir, m); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "references", "google-workspace.md")); !os.IsNotExist(err) {
		t.Error("the dropped connector's generated reference must be removed")
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Errorf("a file without the generated marker must never be removed: %v", err)
	}
	if strings.Contains(read(t, filepath.Join(dir, "SKILL.md")), "google-workspace.md") {
		t.Error("the index must not link a removed reference")
	}
}

func TestUnmarkedSkillIsBackedUpBeforeReplace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	old := "---\nname: connectors-hu\n---\n\nhand edited, no markers\n"
	if err := os.WriteFile(path, []byte(old), 0644); err != nil {
		t.Fatal(err)
	}
	if err := updateSkillAt(dir, testManifest()); err != nil {
		t.Fatal(err)
	}
	if got := read(t, filepath.Join(dir, unmarkedBackupName)); got != old {
		t.Errorf("backup = %q, want the previous SKILL.md", got)
	}
	if strings.Contains(read(t, path), "hand edited") {
		t.Error("unmarked text is not carried over; it lives in the backup")
	}
}

func TestEveryReferenceIsLinkedAndNoTempFileLeft(t *testing.T) {
	dir := t.TempDir()
	if err := updateSkillAt(dir, testManifest()); err != nil {
		t.Fatal(err)
	}
	skill := read(t, filepath.Join(dir, "SKILL.md"))
	entries, err := os.ReadDir(filepath.Join(dir, "references"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
		if !strings.Contains(skill, "(references/"+e.Name()+")") {
			t.Errorf("references/%s is not linked from SKILL.md", e.Name())
		}
	}
}
