/*
	Read cgroup information from related file.
*/

package cgroup

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Info reads the cgroup mount info from /proc/cgroups
type Info struct {
	Hierarchy  int
	NumCgroups int
	Enabled    bool
}

var (
	cacheInfo map[string]Info
	infoOnce  sync.Once
)

var ErrCgroupInfoNotInitialized = errors.New("environment info was not initialized")

func initCgroupInfo() (map[string]Info, error) {
	f, err := os.Open(procCgroupsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rt := make(map[string]Info)
	s := bufio.NewScanner(f)
	for s.Scan() {
		text := s.Text()
		if text[0] == '#' {
			continue
		}
		parts := strings.Fields(text)
		if len(parts) < 4 {
			continue
		}

		// format: subsys_name hierarchy num_cgroups enabled
		name := parts[0]
		hierarchy, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		numCgroups, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, err
		}
		enabled := parts[3] != "0"
		rt[name] = Info{
			Hierarchy:  hierarchy,
			NumCgroups: numCgroups,
			Enabled:    enabled,
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return rt, nil
}

// GetCgroupInfo() read /proc/cgroups and return the result
func getCgroupInfo() (map[string]Info, error) {
	infoOnce.Do(func() {
		cacheInfo, _ = initCgroupInfo()
	})
	if cacheInfo == nil {
		return nil, ErrCgroupInfoNotInitialized
	}
	return cacheInfo, nil
}

// GetAllSubCgroup reads /proc/cgroups and get all available sub-cgroup as set
func GetAllSubCgroup() (map[string]bool, error) {
	info, err := getCgroupInfo()
	if err != nil {
		return nil, err
	}

	rt := make(map[string]bool)
	for k, v := range info {
		if !v.Enabled {
			continue
		}
		rt[k] = true
	}
	return rt, nil
}

// GetAllSubCgroup reads /proc/cgroups and get hierarchy info for each available sub-cgroup
func GetCgroupHierarchy() (map[string]int, error) {
	info, err := getCgroupInfo()
	if err != nil {
		return nil, err
	}

	rt := make(map[string]int)
	for k, v := range info {
		if !v.Enabled {
			continue
		}
		rt[k] = v.Hierarchy
	}
	return rt, nil
}
