package tmux

// MockClient is a mock implementation of Client for testing
type MockClient struct {
	InTmuxResult      bool
	CreateSessionErr  error
	AttachSessionErr  error
	SplitWindowPaneID int
	SplitWindowErr    error
	SplitPanePaneID   int
	SplitPaneErr      error
	FocusPaneErr      error
	KillPaneErr       error
	ResizePaneErr     error
	ToggleZoomErr     error
	BreakPanePaneID   int
	BreakPaneErr      error
	JoinPanePaneID    int
	JoinPaneErr       error
	CapturePaneResult string
	CapturePaneErr    error
	ListPanesResult   []PaneInfo
	ListPanesErr      error
	PaneExistsResult  bool

	// Track calls for verification
	Calls []MockCall
}

// MockCall records a method call for testing verification
type MockCall struct {
	Method string
	Args   []interface{}
}

// NewMockClient creates a new mock client with sensible defaults
func NewMockClient() *MockClient {
	return &MockClient{
		InTmuxResult:      true,
		SplitWindowPaneID: 1,
		SplitPanePaneID:   2,
		BreakPanePaneID:   100, // Different ID to simulate pane ID change
		JoinPanePaneID:    101, // Different ID to simulate pane ID change
		ListPanesResult:   []PaneInfo{},
		PaneExistsResult:  true,
	}
}

func (m *MockClient) recordCall(method string, args ...interface{}) {
	m.Calls = append(m.Calls, MockCall{Method: method, Args: args})
}

func (m *MockClient) InTmux() bool {
	m.recordCall("InTmux")
	return m.InTmuxResult
}

func (m *MockClient) CreateSession(name string) error {
	m.recordCall("CreateSession", name)
	return m.CreateSessionErr
}

func (m *MockClient) AttachSession(name string) error {
	m.recordCall("AttachSession", name)
	return m.AttachSessionErr
}

func (m *MockClient) SplitWindow(dir, command string, args ...string) (int, error) {
	m.recordCall("SplitWindow", dir, command, args)
	return m.SplitWindowPaneID, m.SplitWindowErr
}

func (m *MockClient) SplitPane(targetPaneID int, vertical bool, dir, command string, args ...string) (int, error) {
	m.recordCall("SplitPane", targetPaneID, vertical, dir, command, args)
	return m.SplitPanePaneID, m.SplitPaneErr
}

func (m *MockClient) FocusPane(paneID int) error {
	m.recordCall("FocusPane", paneID)
	return m.FocusPaneErr
}

func (m *MockClient) KillPane(paneID int) error {
	m.recordCall("KillPane", paneID)
	return m.KillPaneErr
}

func (m *MockClient) ResizePane(paneID int, width int) error {
	m.recordCall("ResizePane", paneID, width)
	return m.ResizePaneErr
}

func (m *MockClient) ToggleZoom(paneID int) error {
	m.recordCall("ToggleZoom", paneID)
	return m.ToggleZoomErr
}

func (m *MockClient) CapturePane(paneID int, lines int) (string, error) {
	m.recordCall("CapturePane", paneID, lines)
	return m.CapturePaneResult, m.CapturePaneErr
}

func (m *MockClient) ListPanes() ([]PaneInfo, error) {
	m.recordCall("ListPanes")
	return m.ListPanesResult, m.ListPanesErr
}

func (m *MockClient) PaneExists(paneID int) bool {
	m.recordCall("PaneExists", paneID)
	return m.PaneExistsResult
}

func (m *MockClient) BreakPane(paneID int) (int, error) {
	m.recordCall("BreakPane", paneID)
	return m.BreakPanePaneID, m.BreakPaneErr
}

func (m *MockClient) JoinPane(paneID int, targetPaneID int) (int, error) {
	m.recordCall("JoinPane", paneID, targetPaneID)
	return m.JoinPanePaneID, m.JoinPaneErr
}
