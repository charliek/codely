package shed

import "os/exec"

// MockClient is a mock implementation of Client for testing
type MockClient struct {
	AvailableResult   bool
	ListShedsResult   []Shed
	ListShedsErr      error
	ListServersResult []Server
	ListServersErr    error
	CreateShedErr     error
	StartShedErr      error
	StopShedErr       error
	DeleteShedErr     error

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
		AvailableResult: true,
		ListShedsResult: []Shed{},
	}
}

func (m *MockClient) recordCall(method string, args ...interface{}) {
	m.Calls = append(m.Calls, MockCall{Method: method, Args: args})
}

func (m *MockClient) Available() bool {
	m.recordCall("Available")
	return m.AvailableResult
}

func (m *MockClient) ListSheds() ([]Shed, error) {
	m.recordCall("ListSheds")
	return m.ListShedsResult, m.ListShedsErr
}

func (m *MockClient) ListServers() ([]Server, error) {
	m.recordCall("ListServers")
	return m.ListServersResult, m.ListServersErr
}

func (m *MockClient) CreateShed(name string, opts CreateOpts) error {
	m.recordCall("CreateShed", name, opts)
	return m.CreateShedErr
}

func (m *MockClient) StartShed(name string) error {
	m.recordCall("StartShed", name)
	return m.StartShedErr
}

func (m *MockClient) StopShed(name string) error {
	m.recordCall("StopShed", name)
	return m.StopShedErr
}

func (m *MockClient) DeleteShed(name string, force bool) error {
	m.recordCall("DeleteShed", name, force)
	return m.DeleteShedErr
}

func (m *MockClient) CreateShedStreaming(name string, opts CreateOpts) (string, <-chan string, <-chan error) {
	m.recordCall("CreateShedStreaming", name, opts)

	cmdLine := "shed create " + name + " --json"
	outputCh := make(chan string)
	doneCh := make(chan error, 1)

	go func() {
		close(outputCh)
		doneCh <- m.CreateShedErr
		close(doneCh)
	}()

	return cmdLine, outputCh, doneCh
}

func (m *MockClient) ExecCommand(shedName, command string, args ...string) *exec.Cmd {
	m.recordCall("ExecCommand", shedName, command, args)
	// Return a dummy command that will work
	return exec.Command("echo", "mock")
}

func (m *MockClient) Console(shedName string) *exec.Cmd {
	m.recordCall("Console", shedName)
	return exec.Command("echo", "mock")
}
