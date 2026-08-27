package fixture

import (
	"path/filepath"
	"runtime"
)

func testdataDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("fixture: cannot resolve testdata directory")
	}

	return filepath.Join(filepath.Dir(file), "..", "testdata")
}

// TestdataPath returns the filesystem path to e2e/testdata/name.
func TestdataPath(name string) string {
	return filepath.Join(testdataDir(), name)
}
