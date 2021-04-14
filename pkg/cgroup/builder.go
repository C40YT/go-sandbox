/*
	Config cgroup.
	Use Builder pattern to initialize Cgroup.
	Builder.build() returns cgroup.
*/
package cgroup

import (
	"fmt"
	"strings"
)

// Builder builds cgroup directories
type Builder struct {
	Prefix string

	CPU     bool
	CPUSet  bool
	CPUAcct bool
	Memory  bool
	Pids    bool
}

// NewBuilder return a dumb builder without any sub-cgroup
func NewBuilder(prefix string) *Builder {
	return &Builder{
		Prefix: prefix,
	}
}

// WithCPU includes cpu cgroup
func (b *Builder) WithCPU() *Builder {
	b.CPU = true
	return b
}

// WithCPUSet includes cpuset cgroup
func (b *Builder) WithCPUSet() *Builder {
	b.CPUSet = true
	return b
}

// WithCPUAcct includes cpuacct cgroup
func (b *Builder) WithCPUAcct() *Builder {
	b.CPUAcct = true
	return b
}

// WithMemory includes memory cgroup
func (b *Builder) WithMemory() *Builder {
	b.Memory = true
	return b
}

// WithPids includes pids cgroup
func (b *Builder) WithPids() *Builder {
	b.Pids = true
	return b
}

// FilterByEnv reads /proc/cgroups and filter out non-exists ones
func (b *Builder) FilterByEnv() (*Builder, error) {
	m, err := GetAllSubCgroup()
	if err != nil {
		return b, err
	}
	b.CPUSet = b.CPUSet && m["cpuset"]
	b.CPUAcct = b.CPUAcct && m["cpuacct"]
	b.Memory = b.Memory && m["memory"]
	b.Pids = b.Pids && m["pids"]
	return b, nil
}

// String prints the build properties
func (b *Builder) String() string {
	s := make([]string, 0, 3)
	for _, t := range []struct {
		name    string
		enabled bool
	}{
		{"cpuset", b.CPUSet},
		{"cpuacct", b.CPUAcct},
		{"memory", b.Memory},
		{"pids", b.Pids},
	} {
		if t.enabled {
			s = append(s, t.name)
		}
	}
	return fmt.Sprintf("cgroup builder: [%s]", strings.Join(s, ", "))
}

// Build creates new cgroup directories
func (b *Builder) Build() (cg *Cgroup, err error) {
	// sub cgroup configured must be supported in current environment
	hierarchyMap, err := GetCgroupHierarchy()
	if err != nil {
		// Could not get cgroup information in this machine
		return
	}

	var subPaths []string
	subCgroup := make(map[int]*SubCgroup) //the index is a hierarchy number

	// if failed, remove potential created directory
	defer func() {
		if err != nil {
			_ = removeAll(subPaths...)
		}
	}()

	for _, c := range []struct {
		enable    bool
		groupName string
	}{
		{b.CPU, "cpu"},
		{b.CPUSet, "cpuset"},
		{b.CPUAcct, "cpuacct"},
		{b.Memory, "memory"},
		{b.Pids, "pids"},
	} {
		if !c.enable {
			continue
		}
		h := hierarchyMap[c.groupName]
		if subCgroup[h] == nil {
			var subPath string
			if subPath, err = createTempDir(basePath, c.groupName, b.Prefix); err != nil {
				return
			}
			subPaths = append(subPaths, subPath)
			subCgroup[h] = NewSubCgroup(subPath)
		}
	}

	if b.CPUSet {
		if err = initCpuset(subCgroup[hierarchyMap["cpuset"]].path); err != nil {
			return
		}
	}

	var all []*SubCgroup
	for _, v := range subCgroup {
		all = append(all, v)
	}

	return &Cgroup{
		prefix:  b.Prefix,
		cpu:     subCgroup[hierarchyMap["cpu"]],
		cpuset:  subCgroup[hierarchyMap["cpuset"]],
		cpuacct: subCgroup[hierarchyMap["cpuacct"]],
		memory:  subCgroup[hierarchyMap["memory"]],
		pids:    subCgroup[hierarchyMap["pids"]],
		all:     all,
	}, nil
}
