package game_version

import (
	"fmt"
	"strconv"
	"strings"
)

type GameVersion struct {
	Major    int
	Minor    int
	Build    int
	Revision int
}

func Parse(version string) (GameVersion, error) {
	var v GameVersion
	parts := strings.Split(version, ".")
	if len(parts) != 4 {
		return v, fmt.Errorf("invalid version format: %s", version)
	}

	nums := make([]int, 4)
	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return v, fmt.Errorf("invalid number in version: %s", part)
		}
		nums[i] = num
	}

	return GameVersion{
		Major:    nums[0],
		Minor:    nums[1],
		Build:    nums[2],
		Revision: nums[3],
	}, nil
}

func (v GameVersion) IsNewerThan(other GameVersion) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	if v.Build != other.Build {
		return v.Build > other.Build
	}
	return v.Revision > other.Revision
}

func GetLatest(versions []string) (GameVersion, error) {
	if len(versions) == 0 {
		return GameVersion{}, fmt.Errorf("no versions provided")
	}

	var latest GameVersion

	for i, ver := range versions {
		parsed, err := Parse(ver)
		if err != nil {
			return GameVersion{}, err
		}

		if i == 0 || parsed.IsNewerThan(latest) {
			latest = parsed
		}
	}

	return latest, nil
}
