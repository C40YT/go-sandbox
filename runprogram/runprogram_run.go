package runprogram

import (
	libseccomp "github.com/seccomp/libseccomp-golang"

	"github.com/criyle/go-judger/secutil"
	"github.com/criyle/go-judger/tracee"
	"github.com/criyle/go-judger/tracer"
)

// Start starts the tracing process
func (r *RunProgram) Start() (rt tracer.TraceResult, err error) {
	// build seccomp filter
	filter, err := buildFilter(r.ShowDetails, r.SyscallAllowed, r.SyscallTraced)
	if err != nil {
		println(err)
		return
	}
	defer filter.Release()

	bpf, err := secutil.FilterToBPF(filter)
	if err != nil {
		println(err)
		return
	}

	ch := &tracee.Runner{
		Args:    r.Args,
		Env:     r.Env,
		RLimits: r.RLimits.prepareRLimit(),
		Files:   r.Files,
		WorkDir: r.WorkDir,
		BPF:     bpf,
	}

	th := &tracerHandler{
		ShowDetails: r.ShowDetails,
		Unsafe:      r.Unsafe,
		Handler:     r.Handler,
	}
	return tracer.Trace(th, ch, tracer.ResLimit(r.TraceLimit))
}

// build filter builds the libseccomp filter according to the allow, trace and show details
func buildFilter(showDetails bool, allow, trace []string) (*libseccomp.ScmpFilter, error) {
	// make filter
	var defaultAction libseccomp.ScmpAction
	// if debug, allow all syscalls and output what was blocked
	if showDetails {
		defaultAction = libseccomp.ActTrace.SetReturnCode(tracer.MsgDisallow)
	} else {
		defaultAction = libseccomp.ActKill
	}
	return secutil.BuildFilter(defaultAction, libseccomp.ActTrace.SetReturnCode(tracer.MsgHandle), allow, trace)
}