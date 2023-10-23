package tracing

import (
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type SpanIndex struct {
	ch.CHModel `ch:"table:spans_index,insert:spans_index_buffer,alias:s"`

	*Span

	Count           float32
	LinkCount       uint8
	EventCount      uint8
	EventErrorCount uint8
	EventLogCount   uint8

	AllKeys      []string `ch:"type:Array(LowCardinality(String))"`
	StringKeys   []string `ch:"type:Array(LowCardinality(String))"`
	StringValues []string

	TelemetrySDKName     string `ch:",lc"`
	TelemetrySDKLanguage string `ch:",lc"`
	TelemetrySDKVersion  string `ch:",lc"`
	TelemetryAutoVersion string `ch:",lc"`

	OtelLibraryName    string `ch:",lc"`
	OtelLibraryVersion string `ch:",lc"`

	DeploymentEnvironment string `ch:",lc"`

	ServiceName    string `ch:",lc"`
	ServiceVersion string `ch:",lc"`
	HostName       string `ch:",lc"`

	ClientAddress       string `ch:",lc"`
	ClientSocketAddress string `ch:",lc"`
	ClientSocketPort    int32

	URLScheme string `attr:"url.scheme" ch:",lc"`
	URLFull   string `attr:"url.full"`
	URLPath   string `attr:"url.path" ch:",lc"`

	HTTPRequestMethod      string `ch:",lc"`
	HTTPResponseStatusCode uint16
	HTTPRoute              string `ch:",lc"`

	DBSystem    string `ch:",lc"`
	DBStatement string
	DBOperation string `ch:",lc"`
	DBSqlTable  string `ch:",lc"`

	LogSeverity string `ch:",lc"`
	LogMessage  string

	ExceptionType    string `ch:",lc"`
	ExceptionMessage string
}

func initSpanIndex(index *SpanIndex, span *Span) {
	index.Span = span
	index.Count = 1

	index.TelemetrySDKName = span.Attrs.Text(attrkey.TelemetrySDKName)
	index.TelemetrySDKLanguage = span.Attrs.Text(attrkey.TelemetrySDKLanguage)
	index.TelemetrySDKVersion = span.Attrs.Text(attrkey.TelemetrySDKVersion)
	index.TelemetryAutoVersion = span.Attrs.Text(attrkey.TelemetryAutoVersion)

	index.OtelLibraryName = span.Attrs.Text(attrkey.OtelLibraryName)
	index.OtelLibraryVersion = span.Attrs.Text(attrkey.OtelLibraryVersion)

	index.DeploymentEnvironment, _ = span.Attrs[attrkey.DeploymentEnvironment].(string)

	index.ServiceName = span.Attrs.ServiceName()
	index.ServiceVersion = span.Attrs.Text(attrkey.ServiceVersion)
	index.HostName = span.Attrs.HostName()

	index.ClientAddress = span.Attrs.Text(attrkey.ClientAddress)
	index.ClientSocketAddress = span.Attrs.Text(attrkey.ClientSocketAddress)
	index.ClientSocketPort = int32(span.Attrs.Int64(attrkey.ClientSocketPort))

	index.URLScheme = span.Attrs.Text(attrkey.URLScheme)
	index.URLFull = span.Attrs.Text(attrkey.URLFull)
	index.URLPath = span.Attrs.Text(attrkey.URLPath)

	index.HTTPRequestMethod = span.Attrs.Text(attrkey.HTTPRequestMethod)
	index.HTTPResponseStatusCode = uint16(span.Attrs.Uint64(attrkey.HTTPResponseStatusCode))
	index.HTTPRoute = span.Attrs.Text(attrkey.HTTPRoute)

	index.DBSystem = span.Attrs.Text(attrkey.DBSystem)
	index.DBStatement = span.Attrs.Text(attrkey.DBStatement)
	index.DBOperation = span.Attrs.Text(attrkey.DBOperation)
	index.DBSqlTable = span.Attrs.Text(attrkey.DBSqlTable)

	index.LogSeverity = span.Attrs.Text(attrkey.LogSeverity)
	index.LogMessage = span.Attrs.Text(attrkey.LogMessage)

	index.ExceptionType = span.Attrs.Text(attrkey.ExceptionType)
	index.ExceptionMessage = span.Attrs.Text(attrkey.ExceptionMessage)

	index.AllKeys = mapKeys(span.Attrs)
	index.StringKeys, index.StringValues = attrKeysAndValues(span.Attrs)
}

func mapKeys(m AttrMap) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

var (
	indexedAttrs = []string{
		attrkey.DisplayName,

		attrkey.TelemetrySDKName,
		attrkey.TelemetrySDKLanguage,
		attrkey.TelemetrySDKVersion,
		attrkey.TelemetryAutoVersion,

		attrkey.OtelLibraryName,
		attrkey.OtelLibraryVersion,

		attrkey.DeploymentEnvironment,

		attrkey.ServiceName,
		attrkey.ServiceVersion,
		attrkey.HostName,

		attrkey.ClientAddress,
		attrkey.ClientSocketAddress,
		attrkey.ClientSocketPort,

		attrkey.URLScheme,
		attrkey.URLFull,
		attrkey.URLPath,

		attrkey.HTTPRequestMethod,
		attrkey.HTTPResponseStatusCode,
		attrkey.HTTPRoute,

		attrkey.DBSystem,
		attrkey.DBStatement,
		attrkey.DBOperation,
		attrkey.DBSqlTable,

		attrkey.LogSeverity,
		attrkey.LogMessage,

		attrkey.ExceptionType,
		attrkey.ExceptionMessage,
	}
	indexedAttrSet = listToSet(indexedAttrs)
)

func IsIndexedAttr(key string) bool {
	_, ok := indexedAttrSet[key]
	return ok
}

func attrKeysAndValues(m AttrMap) ([]string, []string) {
	keys := make([]string, 0, len(m))
	values := make([]string, 0, len(m))
	for k, v := range m {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if IsIndexedAttr(k) {
			continue
		}
		keys = append(keys, k)
		values = append(values, utf8util.TruncMedium(asString(v)))
	}
	return keys, values
}
