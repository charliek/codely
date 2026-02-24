package shed

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockClient(t *testing.T) {
	m := NewMockClient()

	assert.True(t, m.Available())
	assert.Len(t, m.Calls, 1)
	assert.Equal(t, "Available", m.Calls[0].Method)
}

func TestMockClientListSheds(t *testing.T) {
	m := NewMockClient()
	m.ListShedsResult = []Shed{
		{Name: "test-shed", Server: "mini-desktop", Status: "running"},
		{Name: "other-shed", Server: "cloud-vps", Status: "stopped"},
	}

	sheds, err := m.ListSheds()

	assert.NoError(t, err)
	assert.Len(t, sheds, 2)
	assert.Equal(t, "test-shed", sheds[0].Name)
	assert.Equal(t, "running", sheds[0].Status)
}

func TestMockClientCreateShed(t *testing.T) {
	m := NewMockClient()

	err := m.CreateShed("new-shed", CreateOpts{
		Repo:   "user/repo",
		Server: "mini-desktop",
	})

	assert.NoError(t, err)
	assert.Len(t, m.Calls, 1)
	assert.Equal(t, "CreateShed", m.Calls[0].Method)
	assert.Equal(t, "new-shed", m.Calls[0].Args[0])
}

func TestMockClientStartStop(t *testing.T) {
	m := NewMockClient()

	err := m.StartShed("test-shed")
	assert.NoError(t, err)

	err = m.StopShed("test-shed")
	assert.NoError(t, err)

	assert.Len(t, m.Calls, 2)
	assert.Equal(t, "StartShed", m.Calls[0].Method)
	assert.Equal(t, "StopShed", m.Calls[1].Method)
}

func TestMockClientDeleteShed(t *testing.T) {
	m := NewMockClient()

	err := m.DeleteShed("test-shed", true)

	assert.NoError(t, err)
	assert.Len(t, m.Calls, 1)
	assert.Equal(t, "DeleteShed", m.Calls[0].Method)
	assert.Equal(t, "test-shed", m.Calls[0].Args[0])
	assert.Equal(t, true, m.Calls[0].Args[1])
}

func TestMockClientExecCommand(t *testing.T) {
	m := NewMockClient()

	cmd := m.ExecCommand("test-shed", "claude", "--help")

	assert.NotNil(t, cmd)
	assert.Len(t, m.Calls, 1)
	assert.Equal(t, "ExecCommand", m.Calls[0].Method)
}

func TestMockClientConsole(t *testing.T) {
	m := NewMockClient()

	cmd := m.Console("test-shed")

	assert.NotNil(t, cmd)
	assert.Len(t, m.Calls, 1)
	assert.Equal(t, "Console", m.Calls[0].Method)
}

func TestShedStruct(t *testing.T) {
	now := time.Now()
	s := Shed{
		Name:      "test-shed",
		Server:    "mini-desktop",
		Status:    "running",
		CreatedAt: now,
		Repo:      "user/repo",
		Backend:   "docker",
	}

	assert.Equal(t, "test-shed", s.Name)
	assert.Equal(t, "mini-desktop", s.Server)
	assert.Equal(t, "running", s.Status)
	assert.Equal(t, now, s.CreatedAt)
	assert.Equal(t, "user/repo", s.Repo)
	assert.Equal(t, "docker", s.Backend)
}

func TestShedStructJSON(t *testing.T) {
	input := `{"name":"test","server":"mini","status":"running","created_at":"2025-01-01T00:00:00Z","backend":"firecracker"}`

	var s Shed
	require.NoError(t, json.Unmarshal([]byte(input), &s))

	assert.Equal(t, "test", s.Name)
	assert.Equal(t, "mini", s.Server)
	assert.Equal(t, "running", s.Status)
	assert.Equal(t, "firecracker", s.Backend)
}

func TestShedStructJSON_OmitsEmptyBackend(t *testing.T) {
	s := Shed{Name: "test", Server: "mini", Status: "running"}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	assert.NotContains(t, string(data), "backend")
}

func TestParseExecError_JSONError(t *testing.T) {
	// Use Output() so stderr is captured into ExitError.Stderr
	cmd := exec.Command("sh", "-c", `echo '{"error":"shed not found"}' >&2; exit 1`)
	_, err := cmd.Output()
	require.Error(t, err)

	result := parseExecError("start", err)
	assert.Contains(t, result.Error(), "shed start: shed not found")
}

func TestParseExecError_RawStderr(t *testing.T) {
	cmd := exec.Command("sh", "-c", `echo 'something went wrong' >&2; exit 1`)
	// Use Output() to capture stderr into ExitError
	_, err := cmd.Output()
	require.Error(t, err)

	result := parseExecError("stop", err)
	assert.Contains(t, result.Error(), "something went wrong")
}

func TestParseExecError_EmptyStderr(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 1")
	_, err := cmd.Output()
	require.Error(t, err)

	result := parseExecError("delete", err)
	assert.Contains(t, result.Error(), "shed delete failed")
}

func TestParseExecError_NonExitError(t *testing.T) {
	err := fmt.Errorf("not an exec error")
	result := parseExecError("create", err)
	assert.Contains(t, result.Error(), "shed create failed")
}
