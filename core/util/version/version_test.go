package version

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	items := []struct {
		major, minor, bugfix, prerelease uint8
		isErr                            bool
	}{
		{1, 2, 3, 0, false},
		{2, 3, 4, 1, true},
		{2, 3, 4, 0, false},
		{8, 9, 10, 1, true},
		{8, 9, 10, 0, false},
		{8, 9, 10, 2, true},
		{8, 9, 10, 3, true},
		{8, 0, 0, 3, true},
		{8, 0, 0, 1, false},
		{8, 0, 0, 2, false},
	}

	for _, item := range items {
		options := []Option{
			WithMajor(item.major),
			WithMinor(item.minor),
			WithBugfix(item.bugfix),
		}
		if item.prerelease != 0 {
			options = append(options, WithPrerelease(item.prerelease))
		}

		version, err := New(options...)
		assert.Equal(t, item.isErr, err != nil)

		if !item.isErr {
			assert.Equal(t, item.major, version.Major())
			assert.Equal(t, item.minor, version.Minor())
			assert.Equal(t, item.bugfix, version.Bugfix())
			assert.Equal(t, item.prerelease, version.Prerelease())
		}
	}
}

func TestVersion_String(t *testing.T) {
	items := []struct {
		major, minor, bugfix, prerelease uint8
		version                          string
	}{
		{1, 2, 3, 0, "1.2.3"},
		{2, 3, 4, 0, "2.3.4"},
		{8, 9, 10, 0, "8.9.10"},
		{8, 0, 0, 1, "8.0.0.1"},
		{8, 0, 0, 2, "8.0.0.2"},
	}

	for _, item := range items {
		options := []Option{
			WithMajor(item.major),
			WithMinor(item.minor),
			WithBugfix(item.bugfix),
		}
		if item.prerelease != 0 {
			options = append(options, WithPrerelease(item.prerelease))
		}

		version, err := New(options...)
		assert.Nil(t, err)

		assert.Equal(t, item.version, version.String())
	}
}

func TestVersion_Clone(t *testing.T) {
	items := []struct {
		major, minor, bugfix, prerelease uint8
	}{
		{1, 2, 3, 0},
		{2, 3, 4, 0},
		{8, 9, 10, 0},
		{8, 0, 0, 1},
		{8, 0, 0, 2},
	}

	for _, item := range items {
		options := []Option{
			WithMajor(item.major),
			WithMinor(item.minor),
			WithBugfix(item.bugfix),
		}
		if item.prerelease != 0 {
			options = append(options, WithPrerelease(item.prerelease))
		}

		version, err := New(options...)
		assert.Nil(t, err)

		cloneVersion := version.Clone()
		assert.EqualValues(t, version, cloneVersion)
	}
}

func TestVersion_OnOrAfter(t *testing.T) {

	var items []struct {
		major, minor, bugfix, prerelease uint8
	}

	for major := uint8(0); major < 255; major++ {
		for minor := uint8(0); minor < 255; minor++ {
			for bugfix := uint8(0); bugfix < 255; bugfix++ {
				for prerelease := uint8(0); prerelease < 3; prerelease++ {
					if prerelease != 0 && bugfix == 0 && minor == 0 {
						items = append(items, struct {
							major, minor, bugfix, prerelease uint8
						}{major: major, minor: minor, bugfix: bugfix, prerelease: prerelease})
					} else {
						items = append(items, struct {
							major, minor, bugfix, prerelease uint8
						}{major: major, minor: minor, bugfix: bugfix, prerelease: 0})
					}
				}
			}
		}
	}

	versions := make([]*Version, 0)

	for i, item := range items {
		if i == 0 {
			continue
		}

		options := []Option{
			WithMajor(item.major),
			WithMinor(item.minor),
			WithBugfix(item.bugfix),
		}
		if item.prerelease != 0 {
			options = append(options, WithPrerelease(item.prerelease))
		}

		version, err := New(options...)
		assert.Nil(t, err)

		versions = append(versions, version)
	}

	size := len(versions)

	for i := 0; i < 100; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		idx := 1 + r.Intn(size-1)
		cur := versions[idx]
		preIdx := r.Intn(idx)
		pre := versions[preIdx]
		assert.Truef(t, cur.OnOrAfter(pre), "%v %v", cur, pre)
	}
}

func TestParse(t *testing.T) {
	items := []struct {
		major, minor, bugfix, prerelease uint8
		version                          string
		isErr                            bool
	}{
		{1, 2, 3, 0, "1.2.3", false},
		{2, 3, 4, 0, "2.3.4", false},
		{8, 9, 10, 0, "8.9.10", false},
		{8, 0, 0, 1, "8.0.0.1", false},
		{8, 0, 0, 2, "8.0.0.2", false},
		{8, 0, 0, 2, "8.0..2", true},
		{8, 0, 0, 2, "8..0", true},
	}

	for _, item := range items {

		if !item.isErr {
			parse, err := Parse(item.version)
			assert.Nil(t, err)
			assert.EqualValues(t, item.major, parse.Major())
			assert.EqualValues(t, item.minor, parse.Minor())
			assert.EqualValues(t, item.bugfix, parse.Bugfix())
			assert.EqualValues(t, item.prerelease, parse.Prerelease())
		} else {
			_, err := Parse(item.version)
			assert.NotNil(t, err)
		}
	}
}
