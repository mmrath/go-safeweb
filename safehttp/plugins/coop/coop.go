// Package coop provides Cross-Origin-Opener-Policy protection. Specification: https://html.spec.whatwg.org/#cross-origin-opener-policies
package coop

import (
	"github.com/google/go-safeweb/safehttp"
)

// Mode represents a COOP mode.
type Mode string

const (
	// SameOrigin is the strictest and safest COOP available: windows can keep a reference of windows they open only if they are same-origin.
	SameOrigin Mode = "same-origin"
	// SameOriginAllowPopups relaxes the same-origin COOP: windows on this origin that open other windows are allowed to keep a reference, but the opposite is not valid.
	SameOriginAllowPopups Mode = "same-origin-allow-popups"
	// UnsafeNone disables COOP: this is the default value in browsers.
	UnsafeNone Mode = "unsafe-none"
)

// Policy represents a Cross-Origin-Opener-Policy value.
type Policy struct {
	// Mode is the mode for the policy.
	Mode Mode
	// ReportingGroup is an optional reporting group that needs to be defined with the Reporting API.
	ReportingGroup string
	// ReportOnly makes the policy report-only if set.
	ReportOnly bool
}

func (p Policy) String() string {
	if p.ReportingGroup == "" {
		return string(p.Mode)
	}
	return string(p.Mode) + `; report-to "` + p.ReportingGroup + `"`
}

// NewInterceptor constructs an interceptor that applies the given policies.
func NewInterceptor(policies ...Policy) Interceptor {
	var ro []string
	var enf []string
	for _, p := range policies {
		if p.ReportOnly {
			ro = append(ro, p.String())
		} else {
			enf = append(enf, p.String())
		}
	}
	return Interceptor{ro: ro, enf: enf}
}

// Default returns a same-origin enforcing interceptor with the given (potentially empty) report group.
func Default(reportGroup string) Interceptor {
	return NewInterceptor(Policy{Mode: SameOrigin, ReportingGroup: reportGroup})
}

// Interceptor is the interceptor for COOP.
type Interceptor struct {
	ro  []string
	enf []string
}

func (it Interceptor) Before(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, cfg safehttp.InterceptorConfig) safehttp.Result {
	if cfg != nil {
		// We got an override, run its Before phase instead.
		return Interceptor(cfg.(Overrider)).Before(w, r, nil)
	}
	w.Header().Claim("Cross-Origin-Opener-Policy")(it.enf)
	w.Header().Claim("Cross-Origin-Opener-Policy-Report-Only")(it.ro)
	return safehttp.NotWritten()
}

func (it Interceptor) Commit(w *safehttp.ResponseWriter, r *safehttp.IncomingRequest, resp safehttp.Response, cfg safehttp.InterceptorConfig) safehttp.Result {
	return safehttp.NotWritten()
}

// Overrider is a safehttp.InterceptorConfig that allows to override COOP for a specific handler.
type Overrider Interceptor

// Override creates an Overrider with the given policies.
func Override(policies ...Policy) Overrider {
	return Overrider(NewInterceptor(policies...))
}

func (p Overrider) Match(i safehttp.Interceptor) bool {
	_, ok := i.(Interceptor)
	return ok
}
