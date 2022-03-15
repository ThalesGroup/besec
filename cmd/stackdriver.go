// Adapted from https://github.com/andyfusniak/stackdriver-gae-logrus-plugin v0.1.3, MIT license
// Initialize with:
//	   log.SetFormatter(NewStackdriverFormatter(projectID))
// Add the XCloudTraceContext middleware, then use like this in a handler:
//     contextLogger := log.WithContext(r.Context())
//     contextLogger.Debug("...")

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type xctcKeyType string

const xctcKey xctcKeyType = "xctc"

// getXCTC returns the XCloudTraceContent value from the context.
func getXCTC(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	xctc, ok := ctx.Value(xctcKey).(string)
	if !ok {
		return ""
	}
	return xctc
}

// XCloudTraceContext middleware extracts the X-Cloud-Trace-Context
// from the request header and injects it into the context. The value
// is read by the logrus GAEStandardFormatter to thread log entries
// by request.
func XCloudTraceContext(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), xctcKey, r.Header.Get("X-Cloud-Trace-Context"))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// The severity of the event described in a log entry.
// See https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
const (
	lsDEFAULT = "DEFAULT"
	lsDEBUG   = "DEBUG"
	lsINFO    = "INFO"
	//lsNOTICE    = "NOTICE"
	lsWARNING  = "WARNING"
	lsERROR    = "ERROR"
	lsCRITICAL = "CRITICAL"
	//lsALERT     = "ALERT"
	lsEMERGENCY = "EMERGENCY"
)

// StackdriverFormatter implements threaded Stackdriver formatting for logrus.
type StackdriverFormatter struct {
	projectID string
}

type entry struct {
	// Trace string
	// Optional. Resource name of the trace associated with the log
	// entry, if any. If it contains a relative resource name, the name
	// is assumed to be relative to //tracing.googleapis.com. Example:
	// projects/my-projectid/traces/06796866738c859f2f19b7cfb3214824
	Trace string `json:"logging.googleapis.com/trace,omitempty"`

	// Span
	// Optional. The span ID within the trace associated with the log
	// entry.
	//
	// For Trace spans, this is the same format that the Trace API
	// v2 uses: a 16-character hexadecimal encoding of an 8-byte
	// array, such as "000000000000004a"
	SpanID string `json:"logging.googleapis.com/spanId,omitempty"`

	Data     logrus.Fields `json:"data"`
	Message  string        `json:"message,omitempty"`
	Severity string        `json:"severity,omitempty"`
}

// NewStackdriverFormatter returns a new Stackdriver Formatter.
func NewStackdriverFormatter(projectID string, options ...StackdriverOption) *StackdriverFormatter {
	fmtr := StackdriverFormatter{projectID: projectID}
	for _, option := range options {
		option(&fmtr)
	}
	return &fmtr
}

// "X-Cloud-Trace-Context: TRACE_ID/SPAN_ID;o=TRACE_TRUE"
//
// `TRACE_ID` is a 32-character hexadecimal value representing a 128-bit
// number. It should be unique between your requests, unless you
// intentionally want to bundle the requests together. You can use UUIDs.
//
// `SPAN_ID` is the decimal representation of the (unsigned) span ID. It
// should be 0 for the first span in your trace. For subsequent requests,
// set SPAN_ID to the span ID of the parent request. See the description
// of TraceSpan (REST, RPC) for more information about nested traces.
//
// `TRACE_TRUE` must be 1 to trace this request. Specify 0 to not trace the
// request.
func parseXCloudTraceContext(t string) (traceID, spanID string) {
	if t == "" {
		return "", ""
	}

	// 32 characters plus 1 (forward slash) plus 1 (at least one decimal
	// representing the span).
	if len(t) < 34 {
		return "", ""
	}

	// The first character after the TRACE_ID should be a forward slash.
	if t[32] != '/' {
		return "", ""
	}

	// handle "TRACE_ID/SPAN_ID" missing the ";o=1" part.
	last := strings.LastIndex(t, ";")
	if last == -1 {
		return t[0:32], t[33:]
	}
	return t[0:32], t[33:last]
}

// StackdriverOption lets you configure the Formatter.
type StackdriverOption func(*StackdriverFormatter)

// Format formats a logrus entry in Stackdriver format.
func (f *StackdriverFormatter) Format(e *logrus.Entry) ([]byte, error) {
	var levToSev = map[logrus.Level]string{
		logrus.TraceLevel: lsDEFAULT,
		logrus.DebugLevel: lsDEBUG,
		logrus.InfoLevel:  lsINFO,
		logrus.WarnLevel:  lsWARNING,
		logrus.ErrorLevel: lsERROR,
		logrus.FatalLevel: lsCRITICAL,
		logrus.PanicLevel: lsEMERGENCY,
	}

	ee := entry{
		Severity: levToSev[e.Level],
		Message:  e.Message,
		Data:     e.Data,
	}

	xctc := getXCTC(e.Context)
	if xctc != "" {
		traceID, spanID := parseXCloudTraceContext(xctc)
		if traceID != "" && spanID != "" {
			ee.Trace = fmt.Sprintf("projects/%s/traces/%s", f.projectID, traceID)
			ee.SpanID = spanID
		}
	}

	b, err := json.Marshal(ee)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
