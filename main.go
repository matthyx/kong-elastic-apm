package main

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Kong/go-pdk"
	klog "github.com/Kong/go-pdk/log"
	"github.com/Kong/go-pdk/server"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
	"go.elastic.co/apm/v2/transport"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	traceParent    = "traceparent"
	oldTraceParent = "x-external-traceparent"
)

var tracer *apm.Tracer
var Version = "1.13.1"
var Priority = 1
var False = "false"

type Config struct {
	ApmActive                 *bool   `json:"apm_active"`
	ApmApiKey                 *string `json:"apm_api_key"`
	ApmApiRequestSize         *string `json:"apm_api_request_size"`
	ApmApiRequestTime         *string `json:"apm_api_request_time"`
	ApmBreakDownMetrics       *bool   `json:"apm_api_breakdown_metrics"`
	ApmCaptureBody            *string `json:"apm_api_capture_body"`
	ApmCaptureHeaders         *bool   `json:"apm_api_capture_headers"`
	ApmCentralConfig          *bool   `json:"apm_api_central_config"`
	ApmCloudProvider          *string `json:"apm_api_cloud_provider"`
	ApmDisableMetrics         *string `json:"apm_disable_metrics"`
	ApmEnvironment            *string `json:"apm_environment"`
	ApmGlobalLabels           *string `json:"apm_global_labels"`
	ApmLogFile                *string `json:"apm_log_file"`
	ApmLogLevel               *string `json:"apm_log_level"`
	ApmMetricsInterval        *string `json:"apm_metrics_interval"`
	ApmRecording              *bool   `json:"apm_recording"`
	ApmSanitizeFieldsNames    *string `json:"apm_sanitize_fields_names"`
	ApmSecretToken            *string `json:"apm_secret_token"`
	ApmServerTimeout          *string `json:"apm_server_timeout"`
	ApmServerUrl              *string `json:"apm_server_url"`
	ApmServerCert             *string `json:"apm_server_cert"`
	ApmServerVerifyServerCert *bool   `json:"apm_server_verify_server_cert"`
	ApmServiceName            string  `json:"apm_service_name"`
	ApmServiceVersion         string  `json:"apm_service_version"`
	ApmServiceNodeName        *string `json:"apm_service_node_name"`
	ApmTransactionIgnoreUrls  *string `json:"apm_transaction_ignore_urls"`
	ApmTransactionMaxSpans    *int    `json:"apm_transaction_max_spans"`
	ApmTransactionSampleRate  *string `json:"apm_transaction_sample_rate"`
	ApmSpanFramesMinDuration  *string `json:"apm_span_frames_min_duration"`
	ApmStackTraceLimit        *int    `json:"apm_span_stack_trace_limit"`
}

func (conf Config) Active() bool {
	if conf.ApmActive != nil {
		return *conf.ApmActive
	}
	return false
}

func New() interface{} {
	return &Config{}
}

func initTracer(conf Config, logger klog.Log) {
	if !conf.Active() {
		setEnv("ELASTIC_APM_ACTIVE", False, logger)
		_ = logger.Info("APM agent is not activated")
		return
	}
	setEnvString("ELASTIC_APM_SERVER_URL", conf.ApmServerUrl, logger)
	setEnvString("ELASTIC_APM_SERVICE_NODE_NAME", conf.ApmServiceNodeName, logger)
	setEnvString("ELASTIC_APM_ENVIRONMENT", conf.ApmEnvironment, logger)
	setEnvBool("ELASTIC_APM_RECORDING", conf.ApmRecording, logger)
	setEnvString("ELASTIC_APM_GLOBAL_LABELS", conf.ApmGlobalLabels, logger)
	setEnvString("ELASTIC_APM_TRANSACTION_IGNORE_URLS", conf.ApmTransactionIgnoreUrls, logger)
	setEnvString("ELASTIC_APM_SANITIZE_FIELD_NAMES", conf.ApmSanitizeFieldsNames, logger)
	setEnvBool("ELASTIC_APM_CAPTURE_HEADERS", conf.ApmCaptureHeaders, logger)
	setEnvString("ELASTIC_APM_CAPTURE_BODY", conf.ApmCaptureBody, logger)
	setEnvString("ELASTIC_APM_API_REQUEST_TIME", conf.ApmApiRequestTime, logger)
	setEnvString("ELASTIC_APM_API_REQUEST_SIZE", conf.ApmApiRequestSize, logger)
	setEnvInt("ELASTIC_APM_TRANSACTION_MAX_SPANS", conf.ApmTransactionMaxSpans, logger)
	setEnvString("ELASTIC_APM_SPAN_FRAMES_MIN_DURATION", conf.ApmSpanFramesMinDuration, logger)
	setEnvInt("ELASTIC_APM_STACK_TRACE_LIMIT", conf.ApmStackTraceLimit, logger)
	setEnvString("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", conf.ApmTransactionSampleRate, logger)
	setEnvString("ELASTIC_APM_METRICS_INTERVAL", conf.ApmMetricsInterval, logger)
	setEnvString("ELASTIC_APM_DISABLE_METRICS", conf.ApmDisableMetrics, logger)
	setEnvBool("ELASTIC_APM_BREAKDOWN_METRICS", conf.ApmBreakDownMetrics, logger)
	setEnvString("ELASTIC_APM_SERVER_CERT", conf.ApmServerCert, logger)
	setEnvBool("ELASTIC_APM_VERIFY_SERVER_CERT", conf.ApmServerVerifyServerCert, logger)
	setEnvString("ELASTIC_APM_LOG_FILE", conf.ApmLogFile, logger)
	setEnvString("ELASTIC_APM_LOG_LEVEL", conf.ApmLogLevel, logger)
	setEnvBool("ELASTIC_APM_CENTRAL_CONFIG", conf.ApmCentralConfig, logger)
	setEnvString("ELASTIC_APM_CLOUD_PROVIDER", conf.ApmCloudProvider, logger)
	_, err := transport.NewHTTPTransport(transport.HTTPTransportOptions{})
	if err != nil {
		_ = logger.Err("Error reinitializing APM transport: ", err.Error())
		panic(err)
	}
	tracer, err = apm.NewTracer(conf.ApmServiceName, conf.ApmServiceVersion)
	if err != nil {
		_ = logger.Err("Error creating APM tracer: ", err.Error())
		panic(err)
	}
}

func setEnvBool(env string, value *bool, logger klog.Log) {
	if value != nil {
		setEnv(env, strconv.FormatBool(*value), logger)
	}
}

func setEnvInt(env string, value *int, logger klog.Log) {
	if value != nil {
		setEnv(env, strconv.Itoa(*value), logger)
	}
}

func setEnvString(env string, value *string, logger klog.Log) {
	if value != nil {
		setEnv(env, *value, logger)
	}
}

func setEnv(env string, value string, logger klog.Log) {
	_ = logger.Info("Setting ", env, " to ", value)
	err := os.Setenv(env, value)
	if err != nil {
		_ = logger.Err("Error setting environment ", env, " : ", err.Error())
		panic(err)
	}
}

type LogMsg struct {
	Latencies struct {
		Request int `json:"request"`
		Kong    int `json:"kong"`
		Proxy   int `json:"proxy"`
	} `json:"latencies"`
	Service struct {
		Host           string `json:"host"`
		CreatedAt      int    `json:"created_at"`
		ConnectTimeout int    `json:"connect_timeout"`
		ID             string `json:"id"`
		Protocol       string `json:"protocol"`
		ReadTimeout    int    `json:"read_timeout"`
		Port           int    `json:"port"`
		Path           string `json:"path"`
		UpdatedAt      int    `json:"updated_at"`
		WriteTimeout   int    `json:"write_timeout"`
		Retries        int    `json:"retries"`
		WsID           string `json:"ws_id"`
	} `json:"service"`
	Request struct {
		Querystring struct {
		} `json:"querystring"`
		Size    int                    `json:"size"`
		URI     string                 `json:"uri"`
		URL     string                 `json:"url"`
		Headers map[string]interface{} `json:"headers"`
		Method  string                 `json:"method"`
	} `json:"request"`
	Tries []struct {
		BalancerLatency int    `json:"balancer_latency"`
		Port            int    `json:"port"`
		BalancerStart   int64  `json:"balancer_start"`
		IP              string `json:"ip"`
	} `json:"tries"`
	ClientIP    string `json:"client_ip"`
	Workspace   string `json:"workspace"`
	UpstreamURI string `json:"upstream_uri"`
	Response    struct {
		Headers map[string]interface{} `json:"headers"`
		Status  int                    `json:"status"`
		Size    int                    `json:"size"`
	} `json:"response"`
	Route struct {
		ID                      string   `json:"id"`
		Paths                   []string `json:"paths"`
		Protocols               []string `json:"protocols"`
		StripPath               bool     `json:"strip_path"`
		CreatedAt               int      `json:"created_at"`
		WsID                    string   `json:"ws_id"`
		RequestBuffering        bool     `json:"request_buffering"`
		UpdatedAt               int      `json:"updated_at"`
		PreserveHost            bool     `json:"preserve_host"`
		RegexPriority           int      `json:"regex_priority"`
		ResponseBuffering       bool     `json:"response_buffering"`
		HTTPSRedirectStatusCode int      `json:"https_redirect_status_code"`
		PathHandling            string   `json:"path_handling"`
		Service                 struct {
			ID string `json:"id"`
		} `json:"service"`
	} `json:"route"`
	StartedAt int64 `json:"started_at"`
}

func toString(i interface{}) string {
	return fmt.Sprintf("%v", i)
}

func translateHeaders(in map[string]interface{}) map[string][]string {
	out := make(map[string][]string)
	for k, v := range in {
		switch x := v.(type) {
		case []interface{}:
			for _, s := range x {
				out[k] = append(out[k], toString(s))
			}
		default:
			out[k] = []string{toString(v)}
		}
	}
	return out
}

func (conf Config) Access(kong *pdk.PDK) {
	if !conf.Active() {
		return
	}
	var err error
	traceparent, _ := kong.Request.GetHeader(traceParent)
	var traceId apm.TraceID
	if traceparent != "" {
		err = kong.ServiceRequest.SetHeader(oldTraceParent, traceparent)
		if err != nil {
			_ = kong.Log.Err(fmt.Sprintf("Error setting %s header: ", oldTraceParent), err.Error())
			return
		}
		bytes, err := hex.DecodeString(traceparent[3:35])
		if err != nil {
			_ = kong.Log.Err("Error decoding trace id: ", err.Error())
			return
		}
		copy(traceId[:], bytes)
	} else {
		_, err = cryptorand.Read(traceId[:])
		if err != nil {
			_ = kong.Log.Err("Error generating new trace id: ", err.Error())
			return
		}
	}
	var spanId apm.SpanID
	_, err = cryptorand.Read(spanId[:])
	if err != nil {
		_ = kong.Log.Err("Error generating new span id: ", err.Error())
		return
	}
	err = kong.ServiceRequest.SetHeader(traceParent, fmt.Sprintf("00-%s-%s-01", traceId, spanId))
	if err != nil {
		_ = kong.Log.Err(fmt.Sprintf("Error setting %s header: ", traceParent), err.Error())
		return
	}
}

func (conf Config) Log(kong *pdk.PDK) {
	if !conf.Active() {
		return
	}
	// (eventually) initialize tracer
	if tracer == nil {
		initTracer(conf, kong.Log)
	}
	// get and parse log message
	s, err := kong.Log.Serialize()
	if err != nil {
		_ = kong.Log.Err("Error getting log message: ", err.Error())
		return
	}
	var msg LogMsg
	if err := json.Unmarshal([]byte(s), &msg); err != nil {
		_ = kong.Log.Err("Error unmarshalling log message: ", err.Error())
		return
	}
	transactionOptions := apm.TransactionOptions{
		Start: time.Unix(0, msg.StartedAt*int64(time.Millisecond)),
	}
	spanOptions := apm.SpanOptions{
		Start: time.Unix(0, msg.StartedAt*int64(time.Millisecond)+int64(msg.Latencies.Kong*int(time.Millisecond)/2)),
	}
	// check if there is an existing trace parent
	if traceParentHeader, ok := msg.Request.Headers[traceParent]; ok {
		_ = kong.Log.Debug("Found traceParent: ", traceParentHeader)
		traceContext, err := apmhttp.ParseTraceparentHeader(toString(traceParentHeader))
		if err != nil {
			_ = kong.Log.Err("Error parsing traceParent: ", err.Error())
			return
		} else {
			spanOptions.SpanID = traceContext.Span
			if oldTraceParentHeader, ok := msg.Request.Headers[oldTraceParent]; ok {
				_ = kong.Log.Debug("Found oldTraceParent: ", oldTraceParentHeader)
				transactionOptions.TraceContext, err = apmhttp.ParseTraceparentHeader(toString(oldTraceParentHeader))
				if err != nil {
					_ = kong.Log.Err("Error parsing oldTraceParent: ", err.Error())
					return
				}
			} else {
				transactionOptions.TraceContext.Trace = traceContext.Trace
				transactionOptions.TraceContext.Span = apm.SpanID{}
				transactionOptions.TraceContext.Options = apm.TraceOptions(0).WithRecorded(true)
			}
		}
	} else {
		_ = kong.Log.Debug("No traceParent found, skipping message.")
		return
	}
	// create transaction
	transaction := tracer.StartTransactionOptions(fmt.Sprintf("%s %s",
		msg.Request.Method,
		msg.Request.URL,
	), "request", transactionOptions)
	transaction.Duration = time.Duration(msg.Latencies.Request) * time.Millisecond
	defer transaction.End()
	_ = kong.Log.Debug(fmt.Sprintf("Started transaction: %+v", transaction.TraceContext()))
	// only continue if this transaction is sampled
	if !transaction.Sampled() {
		_ = kong.Log.Err("Transaction not sampled, skipping")
		return
	}
	// create span
	serviceURL := fmt.Sprintf("%s:%d%s",
		msg.Service.Host,
		msg.Service.Port,
		msg.Service.Path,
	)
	span := transaction.StartSpanOptions(fmt.Sprintf("%s %s",
		msg.Request.Method, // assume same as transaction
		serviceURL,
	), "external", spanOptions)
	span.Duration = time.Duration(msg.Latencies.Request-msg.Latencies.Kong) * time.Millisecond
	defer span.End()
	_ = kong.Log.Debug(fmt.Sprintf("Started span: %+v", span.TraceContext()))
	// enrich transaction
	// create a fake request to record info
	fakeTransactionRequest, _ := http.NewRequest(msg.Request.Method, msg.Request.URL, nil)
	fakeTransactionRequest.Header = translateHeaders(msg.Request.Headers)
	u, err := url.Parse(msg.Request.URL)
	if err == nil {
		fakeTransactionRequest.Host = u.Host
	}
	fakeTransactionRequest.Method = msg.Request.Method
	fakeTransactionRequest.RequestURI = msg.Request.URL
	transaction.Context.SetHTTPRequest(fakeTransactionRequest)
	transaction.Context.SetHTTPResponseHeaders(translateHeaders(msg.Request.Headers))
	transaction.Result = strconv.Itoa(msg.Response.Status)
	_ = kong.Log.Debug(fmt.Sprintf("Finished with transaction: %+v", transaction.TraceContext()))
	// enrich span
	span.Action = msg.Request.Method
	span.Context.SetDestinationAddress(msg.Service.Host, msg.Service.Port)
	span.Subtype = msg.Service.Protocol
	// create a fake request to enrich span
	fakeSpanRequest, _ := http.NewRequest(msg.Request.Method, serviceURL, nil)
	fakeSpanRequest.Header = translateHeaders(msg.Response.Headers)
	fakeSpanRequest.Host = msg.Service.Host
	fakeSpanRequest.Method = msg.Request.Method
	fakeSpanRequest.RequestURI = fmt.Sprintf("%s://%s",
		msg.Service.Protocol,
		serviceURL,
	)
	span.Context.SetHTTPRequest(fakeSpanRequest)
	service := apm.DestinationServiceSpanContext{Resource: fmt.Sprintf("%s:%d",
		msg.Service.Host,
		msg.Service.Port,
	)}
	span.Context.SetDestinationService(service)
	// If, when the transaction ends, its Outcome field has not
	// been explicitly set, it will be set based on the status code:
	// "success" if statusCode < 400, and "failure" otherwise.
	span.Context.SetHTTPStatusCode(msg.Response.Status)
	_ = kong.Log.Debug(fmt.Sprintf("Finished with span: %+v", span.TraceContext()))
}

func main() {
	err := server.StartServer(New, Version, Priority)
	if err != nil {
		log.Printf("Error starting embedded plugin server: %s", err.Error())
		panic(err)
	}
	if tracer != nil {
		tracer.Flush(nil)
		tracer.Close()
	}
}
