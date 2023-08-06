package repodata

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Embed test data https://pkg.go.dev/embed
//
//go:embed testdata/*
var testdata embed.FS

func writeTestdataToTmpfile(t *testing.T, testdataPath string) string {
	tmpdir := t.TempDir()
	data, err := testdata.ReadFile("testdata/" + testdataPath)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	tmpfile := filepath.Join(tmpdir, filepath.Base(testdataPath))
	if err := os.WriteFile(tmpfile, data, 0644); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	return tmpfile
}

func loadTestdataRepodata(t *testing.T, testdataPath string) *Repodata {
	data, err := testdata.ReadFile("testdata/" + testdataPath)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	var repodata Repodata
	if err := json.Unmarshal(data, &repodata); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	return &repodata
}

func TestPackageIsAllowed(t *testing.T) {
	allowedPackages := NewSet(&[]string{"foo"})

	assert.True(t, packageIsAllowed("foo", allowedPackages))
	assert.False(t, packageIsAllowed("bar", allowedPackages))
	assert.True(t, packageIsAllowed("bar", nil))
}

func TestFilenameIsValid(t *testing.T) {
	repodataLinux64 := Repodata{
		RepodataVersion: 1,
		Info:            RepodataInfo{Subdir: "linux-64"},
	}
	repodataWinArm64 := Repodata{
		RepodataVersion: 1,
		Info:            RepodataInfo{Subdir: "win-arm64"},
	}
	record := RepodataRecord{
		Subdir:  "linux-64",
		Name:    "foo",
		Version: "1.0",
		Build:   "py_0",
	}

	testCases := []struct {
		filename  string
		repodata  *Repodata
		extension string
		valid     bool
	}{
		{"foo-1.0-py_0.conda", &repodataLinux64, ".conda", true},
		{"foo-1.0-py_0.tar.bz2", &repodataLinux64, ".tar.bz2", true},
		{"foo-1.0-py_x.conda", &repodataLinux64, ".conda", false},
		{"foo-1.0-py_0.conda", &repodataWinArm64, ".conda", false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v,%v,%v", tc.filename, tc.repodata, tc.extension), func(t *testing.T) {
			assert.Equal(t,
				tc.valid,
				FilenameIsValid(tc.filename, tc.extension, tc.repodata, &record))
		},
		)
	}
}

func TestLoadRepodata(t *testing.T) {
	repodata_json := writeTestdataToTmpfile(t, filepath.Join("linux-64", "repodata.json"))
	repodata, err := LoadRepodata(repodata_json)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	assert.Equal(t, 2, len(repodata.PackagesConda))
	assert.Equal(t, repodata.PackagesConda["d-2023.1.1-0.conda"], RepodataRecord{
		Subdir:      "linux-64",
		Name:        "d",
		Version:     "2023.1.1",
		Build:       "0",
		BuildNumber: 0,
		Depends:     []string{"c"},
		Extra:       map[string]interface{}{},
	})
	assert.Equal(t, repodata.PackagesConda["e-12.34.56-78.conda"], RepodataRecord{
		Subdir:      "linux-64",
		Name:        "e",
		Version:     "12.34.56",
		Build:       "78",
		BuildNumber: 78,
		Depends:     []string{},
		Extra:       map[string]interface{}{},
	})
}

func TestParseRepodata(t *testing.T) {
	testCases := []struct {
		subDir                   string
		expectedFileNames        *[]string
		expectedPackageNames     *[]string
		expectedLenPackages      int
		expectedLenPackagesConda int
	}{
		{
			"noarch",
			&[]string{
				"testdata/noarch/a-0.1.0-0.tar.bz2",
				"testdata/noarch/a-0.2.0-abc_0.tar.bz2",
				"testdata/noarch/b-1-10.tar.bz2",
				"testdata/noarch/c-1.2.3-aaa_0.conda",
			},
			&[]string{"a", "b", "c"}, 3, 1,
		},
		{
			"linux-64",
			&[]string{
				"testdata/linux-64/d-2023.1.1-0.conda",
				"testdata/linux-64/e-12.34.56-78.conda",
			},
			&[]string{"d", "e"}, 0, 2,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.subDir), func(t *testing.T) {
			repodata_json := writeTestdataToTmpfile(t, filepath.Join(tc.subDir, "repodata.json"))
			filtered, fileNames, packageNames, err := ParseRepodata("testdata", repodata_json, nil)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			assert.ElementsMatch(t, *tc.expectedFileNames, *fileNames.Items())
			assert.ElementsMatch(t, *tc.expectedPackageNames, *packageNames.Items())

			assert.Equal(t, tc.expectedLenPackages, len(filtered.Packages))
			assert.Equal(t, tc.expectedLenPackagesConda, len(filtered.PackagesConda))
		},
		)
	}
}

func TestFilterRepodataByAllowed(t *testing.T) {
	repodata := &Repodata{
		RepodataVersion: 1,
		Info:            RepodataInfo{Subdir: "linux-64"},
		Packages: map[string]RepodataRecord{
			"a-1-0.tar.bz2": RepodataRecord{
				Subdir:  "linux-64",
				Name:    "a",
				Version: "1",
				Build:   "0",
			},
			"b-2-0.tar.bz2": RepodataRecord{
				Subdir:  "linux-64",
				Name:    "b",
				Version: "2",
				Build:   "0",
			},
		},
		PackagesConda: map[string]RepodataRecord{
			"c-3-0.conda": RepodataRecord{
				Subdir:  "linux-64",
				Name:    "c",
				Version: "3",
				Build:   "0",
			},
			"d-4-0.conda": RepodataRecord{
				Subdir:  "linux-64",
				Name:    "d",
				Version: "4",
				Build:   "0",
			},
		},
	}

	assert.Equal(t, 2, len(repodata.Packages))
	assert.Equal(t, 2, len(repodata.PackagesConda))
	testCases := []struct {
		allowed           *Set
		expectedFilenames *[]string
	}{
		{
			NewSet(&[]string{"a", "d"}),
			&[]string{
				"a-1-0.tar.bz2",
				"d-4-0.conda",
			},
		},
		{
			NewSet(&[]string{}),
			&[]string{},
		},
		{
			nil,
			&[]string{
				"a-1-0.tar.bz2",
				"b-2-0.tar.bz2",
				"c-3-0.conda",
				"d-4-0.conda",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.allowed), func(t *testing.T) {
			filtered := FilterRepodataByAllowed(repodata, tc.allowed)
			assert.Equal(t, len(*tc.expectedFilenames), len(filtered.Packages)+len(filtered.PackagesConda))
		})
	}
}

func TestParseListFromFile(t *testing.T) {
	allowedFile := writeTestdataToTmpfile(t, "allowed_packages.txt")
	allowed := ParseListFromFile(allowedFile)
	assert.ElementsMatch(t, []string{"bar", "baz", "foo"}, *allowed.Items())
}

func TestParseDependencyName(t *testing.T) {
	testCases := []string{
		"python >=3.10,<3.11",
		"python>=3.9,<3.10",
		"python 3.10",
		"python",
		"python=3.8",
		"python<3.11",
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			assert.Equal(t, "python", parseDependencyName(tc))
		})
	}
}

func TestUpdateDependencyMap(t *testing.T) {
	dependencyMap := make(map[string]Set)

	{
		repodataNoarch := loadTestdataRepodata(t, "noarch/repodata.json")
		UpdateDependencyMap(&dependencyMap, repodataNoarch)

		assert.Equal(t, 3, len(dependencyMap))

		aDeps := dependencyMap["a"]
		assert.ElementsMatch(t, []string{"b", "c", "d"}, *aDeps.Items())

		bDeps := dependencyMap["b"]
		assert.Equal(t, []string{"c"}, *bDeps.Items())

		cDeps := dependencyMap["c"]
		assert.Equal(t, []string{}, *cDeps.Items())
	}

	{
		repodataLinux := loadTestdataRepodata(t, "linux-64/repodata.json")
		UpdateDependencyMap(&dependencyMap, repodataLinux)

		assert.Equal(t, 5, len(dependencyMap))

		aDeps := dependencyMap["a"]
		assert.ElementsMatch(t, []string{"b", "c", "d"}, *aDeps.Items())

		bDeps := dependencyMap["b"]
		assert.Equal(t, []string{"c"}, *bDeps.Items())

		cDeps := dependencyMap["c"]
		assert.Equal(t, []string{}, *cDeps.Items())

		dDeps := dependencyMap["d"]
		assert.Equal(t, []string{"c"}, *dDeps.Items())

		eDeps := dependencyMap["e"]
		assert.Equal(t, []string{}, *eDeps.Items())
	}
}

func TestGetChannelPackageDependencies(t *testing.T) {
	dependencyMap := map[string]Set{
		"a": *NewSet(&[]string{"b", "c", "d"}),
		"b": *NewSet(&[]string{"c"}),
		"c": *NewSet(nil),
		"d": *NewSet(&[]string{"c"}),
		"e": *NewSet(nil),
	}

	testCases := []struct {
		allowedPackages *[]string
		expectedNames   *[]string
	}{
		{&[]string{"a"}, &[]string{"a", "b", "c", "d"}},
		{&[]string{"b"}, &[]string{"b", "c"}},
		{&[]string{"c"}, &[]string{"c"}},
		{&[]string{"d"}, &[]string{"c", "d"}},
		{&[]string{"e"}, &[]string{"e"}},

		{&[]string{"b", "e"}, &[]string{"b", "c", "e"}},
		{&[]string{"b", "e", "a"}, &[]string{"a", "b", "c", "d", "e"}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.allowedPackages), func(t *testing.T) {
			names := GetChannelPackageDependencies(dependencyMap, NewSet(tc.allowedPackages))
			assert.ElementsMatch(t, *tc.expectedNames, *names.Items())
		})
	}
}
