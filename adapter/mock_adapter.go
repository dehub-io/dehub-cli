package adapter

import (
	"errors"
	"io"
)

// MockError mock 错误类型
type MockError string

func (e MockError) Error() string {
	return string(e)
}

// MockAdapter 用于测试的 mock 适配器
type MockAdapter struct {
	LoginError          error
	LogoutError         error
	AuthStatus          *AuthStatus
	AuthStatusError     error
	Namespaces          []*Namespace
	ListNamespacesError error
	CreateNamespaceResult *Namespace
	CreateNamespaceError  error
	GetNamespaceResult    *Namespace
	GetNamespaceError     error
	Packages             []*Package
	ListPackagesError    error
	GetPackageResult     *Package
	GetPackageError      error
	PublishError         error
	InstallError         error
	SearchResults        []*Package
	SearchError          error
}

func NewMockAdapter() *MockAdapter {
	return &MockAdapter{
		AuthStatus: &AuthStatus{LoggedIn: false, ServerType: "mock"},
	}
}

func (m *MockAdapter) Login() error {
	if m.LoginError != nil {
		return m.LoginError
	}
	m.AuthStatus = &AuthStatus{LoggedIn: true, Username: "test-user", ServerType: "mock"}
	return nil
}

func (m *MockAdapter) Logout() error {
	if m.LogoutError != nil {
		return m.LogoutError
	}
	m.AuthStatus = &AuthStatus{LoggedIn: false, ServerType: "mock"}
	return nil
}

func (m *MockAdapter) GetAuthStatus() (*AuthStatus, error) {
	if m.AuthStatusError != nil {
		return nil, m.AuthStatusError
	}
	return m.AuthStatus, nil
}

func (m *MockAdapter) CreateNamespace(name, description string) (*Namespace, error) {
	if m.CreateNamespaceError != nil {
		return nil, m.CreateNamespaceError
	}
	if m.CreateNamespaceResult != nil {
		return m.CreateNamespaceResult, nil
	}
	return &Namespace{Name: name, Status: "pending", Description: description}, nil
}

func (m *MockAdapter) ListNamespaces() ([]*Namespace, error) {
	if m.ListNamespacesError != nil {
		return nil, m.ListNamespacesError
	}
	if m.Namespaces != nil {
		return m.Namespaces, nil
	}
	return []*Namespace{}, nil
}

func (m *MockAdapter) GetNamespace(name string) (*Namespace, error) {
	if m.GetNamespaceError != nil {
		return nil, m.GetNamespaceError
	}
	if m.GetNamespaceResult != nil {
		return m.GetNamespaceResult, nil
	}
	return nil, errors.New("namespace not found")
}

func (m *MockAdapter) Publish(pkgPath string, progress io.Writer) error {
	return m.PublishError
}

func (m *MockAdapter) ListPackages(namespace string) ([]*Package, error) {
	if m.ListPackagesError != nil {
		return nil, m.ListPackagesError
	}
	if m.Packages != nil {
		return m.Packages, nil
	}
	return []*Package{}, nil
}

func (m *MockAdapter) GetPackage(name string) (*Package, error) {
	if m.GetPackageError != nil {
		return nil, m.GetPackageError
	}
	if m.GetPackageResult != nil {
		return m.GetPackageResult, nil
	}
	return nil, errors.New("package not found")
}

func (m *MockAdapter) Install(dependencies map[string]string, targetDir string) error {
	return m.InstallError
}

func (m *MockAdapter) Search(query string) ([]*Package, error) {
	if m.SearchError != nil {
		return nil, m.SearchError
	}
	if m.SearchResults != nil {
		return m.SearchResults, nil
	}
	return []*Package{}, nil
}
