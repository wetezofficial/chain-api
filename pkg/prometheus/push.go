package prometheus

import (
	"bytes"
	"fmt"
	"net/http"
)

const pushgatewayBase = "http://localhost:9091"

func PushMetrics(networkName string, metrics []ErrorNumMetric) {
	body := ""
	for _, metric := range metrics {
		body += metric.String() + "\n"
	}
	err := pushBody(networkName, body)
	if err != nil {
		fmt.Printf("failed to push metrics to pushgateway: %v", err)
	}
}

type ErrorNumMetric struct {
	NodeName string
	ErrorNum int
}

func (m *ErrorNumMetric) String() string {
	return fmt.Sprintf(`error_num{node="%s"} %d`, m.NodeName, m.ErrorNum)
}

// pushBody sends a Prometheus exposition-format body to the Pushgateway in one HTTP request.
// The body may contain multiple metrics, one per line (e.g. "metric_name{label=\"val\"} 123").
func pushBody(networkName string, body string) error {
	url := fmt.Sprintf("%s/metrics/job/rpc_nde/instance/rpc_nde_error/network/%s", pushgatewayBase, networkName)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("pushgateway returned nil response")
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pushgateway returned status %d", resp.StatusCode)
	}
	return nil
}
