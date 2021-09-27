package main

import (
	"fmt"
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/bridge"
	klog "github.com/Kong/go-pdk/log"
	"github.com/Kong/go-pdk/server"
	"github.com/google/uuid"
	"go.elastic.co/apm"
	"go.elastic.co/apm/transport"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const transactionID = "transactionID"

var spans = make(map[string]*apm.Span)
var transactions = make(map[string]*apm.Transaction)
var tracer *apm.Tracer
var Version = "1.13.1"
var Priority = 1
var False = "false"

type Config struct {
	ApmActive                 bool   `json:apm_active`
	ApmApiKey                 string `json:apm_api_key`
	ApmApiRequestSize         string `json:apm_api_request_size`
	ApmApiRequestTime         string `json:apm_api_request_time`
	ApmBreakDownMetrics       bool   `json:apm_api_breakdown_metrics`
	ApmCaptureBody            string `json:apm_api_capture_body`
	ApmCaptureHeaders         bool   `json:apm_api_capture_headers`
	ApmCentralConfig          bool   `json:apm_api_central_config`
	ApmCloudProvider          string `json:apm_api_cloud_provider`
	ApmDisableMetrics         string `json:apm_disable_metrics`
	ApmEnvironment            string `json:apm_environment`
	ApmGlobalLabels           string `json:apm_global_labels`
	ApmLogFile                string `json:apm_log_file`
	ApmLogLevel               string `json:apm_log_level`
	ApmMetricsInterval        string `json:apm_metrics_interval`
	ApmRecording              bool   `json:apm_recording`
	ApmSanitizeFieldsNames    string `json:apm_sanitize_fields_names`
	ApmSecretToken            string `json:apm_secret_token`
	ApmServerTimeout          string `json:apm_server_timeout`
	ApmServerUrl              string `json:apm_server_url`
	ApmServerCert             string `json:apm_server_cert`
	ApmServerVerifyServerCert bool   `json:apm_server_verify_server_cert`
	ApmServiceName            string `json:apm_service_name`
	ApmServiceVersion         string `json:apm_service_version`
	ApmServiceNodeName        string `json:apm_service_node_name`
	ApmTransactionIgnoreUrls  string `json:apm_transaction_ignore_urls`
	ApmTransactionMaxSpans    int    `json:apm_transaction_max_spans`
	ApmTransactionSampleRate  string `json:apm_transaction_sample_rate`
	ApmSpanFramesMinDuration  string `json:apm_span_frames_min_duration`
	ApmStackTraceLimit        int    `json:apm_span_stack_trace_limit`
}

func New() interface{} {
	return &Config{}
}

func initTracer(conf Config, logger klog.Log) {
	if conf.ApmActive == false {
		setEnv("ELASTIC_APM_ACTIVE", False, logger)
		logger.Info("APM agent is not activated")
		return
	}
	setEnv("ELASTIC_APM_SERVER_URL", conf.ApmServerUrl, logger)
	_, err := transport.InitDefault()
	if err != nil {
		logger.Err("Error reinitializing APM transport: ", err.Error())
		panic(err)
	}
	tracer, err = apm.NewTracer(conf.ApmServiceName, conf.ApmServiceVersion)
	if err != nil {
		logger.Err("Error creating APM tracer: ", err.Error())
		panic(err)
	}
	setEnv("ELASTIC_APM_SERVICE_NODE_NAME", conf.ApmServiceNodeName, logger)
	setEnv("ELASTIC_APM_ENVIRONMENT", conf.ApmEnvironment, logger)
	if conf.ApmRecording == false {
		setEnv("ELASTIC_APM_RECORDING", False, logger)
	}
	setEnv("ELASTIC_APM_GLOBAL_LABELS", conf.ApmGlobalLabels, logger)
	setEnv("ELASTIC_APM_TRANSACTION_IGNORE_URLS", conf.ApmTransactionIgnoreUrls, logger)
	setEnv("ELASTIC_APM_SANITIZE_FIELD_NAMES", conf.ApmSanitizeFieldsNames, logger)
	if conf.ApmCaptureHeaders == false {
		setEnv("ELASTIC_APM_CAPTURE_HEADERS", False, logger)
	}
	setEnv("ELASTIC_APM_CAPTURE_BODY", conf.ApmCaptureBody, logger)
	setEnv("ELASTIC_APM_API_REQUEST_TIME", conf.ApmApiRequestTime, logger)
	setEnv("ELASTIC_APM_API_REQUEST_SIZE", conf.ApmApiRequestSize, logger)
	setEnv("ELASTIC_APM_TRANSACTION_MAX_SPANS", strconv.Itoa(conf.ApmTransactionMaxSpans), logger)
	setEnv("ELASTIC_APM_SPAN_FRAMES_MIN_DURATION", conf.ApmSpanFramesMinDuration, logger)
	setEnv("ELASTIC_APM_STACK_TRACE_LIMIT", strconv.Itoa(conf.ApmStackTraceLimit), logger)
	setEnv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", conf.ApmTransactionSampleRate, logger)
	setEnv("ELASTIC_APM_METRICS_INTERVAL", conf.ApmMetricsInterval, logger)
	setEnv("ELASTIC_APM_DISABLE_METRICS", conf.ApmDisableMetrics, logger)
	if conf.ApmBreakDownMetrics == false {
		setEnv("ELASTIC_APM_BREAKDOWN_METRICS", False, logger)
	}
	setEnv("ELASTIC_APM_SERVER_CERT", conf.ApmServerCert, logger)
	if conf.ApmServerVerifyServerCert == false {
		setEnv("ELASTIC_APM_VERIFY_SERVER_CERT", "false", logger)
	}
	setEnv("ELASTIC_APM_LOG_FILE", conf.ApmLogFile, logger)
	setEnv("ELASTIC_APM_LOG_LEVEL", conf.ApmLogLevel, logger)
	if conf.ApmCentralConfig == false {
		setEnv("ELASTIC_APM_CENTRAL_CONFIG", "false", logger)
	}
	setEnv("ELASTIC_APM_CLOUD_PROVIDER", conf.ApmCloudProvider, logger)

}

func setEnv(env string, value string, logger klog.Log) {
	if value != "" {
		logger.Info("Setting ", env, " to ", value)
		err := os.Setenv(env, value)
		if err != nil {
			logger.Err("Error setting environment ", env, " : ", err.Error())
			panic(err)
		}
	}
}

func askMap(b bridge.PdkBridge, method string, logger klog.Log) map[string][]string {
	m, err := b.AskMap(method)
	if err != nil {
		logger.Err("Cannot read map ", method, ": ", err.Error())
		return nil
	}
	logger.Debug(method, " ", m)
	return m
}

func askInt(b bridge.PdkBridge, method string, logger klog.Log) int {
	i, err := b.AskInt(method)
	if err != nil {
		logger.Err("Cannot read int ", method, ": ", err.Error())
		return 0
	}
	logger.Debug(method, " ", i)
	return i
}

func askString(b bridge.PdkBridge, method string, logger klog.Log) string {
	s, err := b.AskString(method)
	if err != nil {
		logger.Err("Cannot read string ", method, ": ", err.Error())
		return ""
	}
	logger.Debug(method, " ", s)
	return s
}

func (conf Config) Access(kong *pdk.PDK) {
	if conf.ApmActive == false {
		return
	}
	// (eventually) initialize tracer
	if tracer == nil {
		initTracer(conf, kong.Log)
	}
	// create and record transaction
	txID := uuid.New().String()
	method := askString(kong.Request.PdkBridge, "kong.request.get_method", kong.Log)
	transactions[txID] = tracer.StartTransaction(fmt.Sprintf("%s %s:%d%s",
		method,
		askString(kong.Request.PdkBridge, "kong.request.get_forwarded_host", kong.Log),
		askInt(kong.Request.PdkBridge, "kong.request.get_forwarded_port", kong.Log),
		askString(kong.Request.PdkBridge, "kong.request.get_forwarded_path", kong.Log),
	), "request")
	kong.Log.Info("Started transaction: ", txID)
	err := kong.Ctx.SetShared(transactionID, txID)
	if err != nil {
		kong.Log.Err("Error saving transactionID in shared context: ", err.Error())
		return
	}
	// only continue if this transaction is sampled
	if !transactions[txID].Sampled() {
		return
	}
	// enrich transaction
	svc, err := kong.Router.GetService()
	if err != nil {
		kong.Log.Err("Error getting service from router: ", err.Error())
		return
	}
	// create span
	spans[txID] = transactions[txID].StartSpan(fmt.Sprintf("%s %s:%d%s",
		method, // assume same as transaction
		svc.Host,
		svc.Port,
		svc.Path,
	), "external", nil)
	kong.Log.Info("Started span: ", txID)
	// enrich span
	spans[txID].Action = method
	spans[txID].Context.SetDestinationAddress(svc.Host, svc.Port)
	spans[txID].Subtype = svc.Protocol
}

func (conf Config) Response(kong *pdk.PDK) {
	if conf.ApmActive == false {
		return
	}
	// retrieve transaction ID
	txID, err := kong.Ctx.GetSharedString(transactionID)
	if err != nil {
		kong.Log.Err("Error getting transactionID from shared context: ", err.Error())
		return
	}
	status := askInt(kong.ServiceResponse.PdkBridge, "kong.service.response.get_status", kong.Log)
	// enrich and close span
	if _, ok := spans[txID]; ok {
		// get service
		svc, err := kong.Router.GetService()
		if err != nil {
			kong.Log.Err("Error getting service from router: ", err.Error())
			return
		}
		// create a fake request to enrich span
		s := strings.SplitN(spans[txID].Name, " ", 2)
		fakeRequest, _ := http.NewRequest(s[0], s[1], nil)
		fakeRequest.Header = askMap(kong.ServiceResponse.PdkBridge, "kong.service.response.get_headers", kong.Log)
		fakeRequest.Host = svc.Host
		fakeRequest.Method = askString(kong.Request.PdkBridge, "kong.request.get_method", kong.Log)
		fakeRequest.RequestURI = fmt.Sprintf("%s://%s:%d/%s",
			svc.Protocol,
			svc.Host,
			svc.Port,
			svc.Path,
		)
		spans[txID].Context.SetHTTPRequest(fakeRequest)
		service := apm.DestinationServiceSpanContext{Resource: fmt.Sprintf("%s:%d",
			svc.Host,
			svc.Port,
		)}
		spans[txID].Context.SetDestinationService(service)
		// If, when the transaction ends, its Outcome field has not
		// been explicitly set, it will be set based on the status code:
		// "success" if statusCode < 400, and "failure" otherwise.
		spans[txID].Context.SetHTTPStatusCode(status)
		kong.Log.Info("Ending span: ", txID)
		spans[txID].End()
		delete(spans, txID)
	}
	// enrich and close transaction
	if _, ok := transactions[txID]; ok {
		// create a fake request to enrich transaction
		s := strings.SplitN(transactions[txID].Name, " ", 2)
		fakeRequest, _ := http.NewRequest(s[0], s[1], nil)
		fakeRequest.Header = askMap(kong.Request.PdkBridge, "kong.request.get_headers", kong.Log)
		fakeRequest.Host = askString(kong.Request.PdkBridge, "kong.request.get_forwarded_host", kong.Log)
		fakeRequest.Method = askString(kong.Request.PdkBridge, "kong.request.get_method", kong.Log)
		fakeRequest.RequestURI = fmt.Sprintf("%s://%s:%d/%s",
			askString(kong.Request.PdkBridge, "kong.request.get_scheme", kong.Log),
			askString(kong.Request.PdkBridge, "kong.request.get_forwarded_host", kong.Log),
			askInt(kong.Request.PdkBridge, "kong.request.get_forwarded_port", kong.Log),
			askString(kong.Request.PdkBridge, "kong.request.get_forwarded_path", kong.Log),
		)
		transactions[txID].Context.SetHTTPRequest(fakeRequest)
		transactions[txID].Context.SetHTTPResponseHeaders(askMap(kong.Response.PdkBridge, "kong.response.get_headers", kong.Log))
		transactions[txID].Result = fmt.Sprintf("HTTP %d", status)
		kong.Log.Info("Ending transaction: ", txID)
		transactions[txID].End()
		delete(transactions, txID)
	}
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
