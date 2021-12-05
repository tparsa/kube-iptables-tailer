package metrics

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestCase struct {
	srcPodName string
	srcNamespace string
	dstPodName string
	dstNamespace string
}

// Test if Metrics can process packetDropsCount with its namespace, other side's service name, and traffic direction
func TestMetricsProcessPacketDrops(t *testing.T) {
	// key: TestCase for packet drop; value: number of count it happens
	testCaseMap := make(map[TestCase]int)
	// construct test case with namespace "test-namespace-i" and count i
	// for trafficDirection, set it "SEND" if i is even, "RECEIVE" if i is odd
	for i := 1; i <= 5; i++ {
		testCase := TestCase{
			srcPodName: fmt.Sprintf("%v", i),
			srcNamespace: "test-namespace",
			dstPodName: fmt.Sprintf("%v", i),
			dstNamespace: "other-side-service-name",
		}
		testCaseMap[testCase] = i
	}
	// simulate the process of Metrics updating packetDropsCount
	// trafficDirection is simulated as sending when namespace has odd number and receiving when it has even number
	for testCase := range testCaseMap {
		for i := 0; i < testCaseMap[testCase]; i++ {
			GetInstance().ProcessPacketDrop(testCase.srcNamespace, testCase.srcPodName, testCase.dstNamespace, testCase.dstPodName)
		}
	}
	// check the actual metrics raw data with expected string
	metricsResult := requestContentBody(GetInstance().GetHandler())
	for testCase, count := range testCaseMap {
		expected := getPacketDropsCountMetricsString(testCase, count)
		if !strings.Contains(metricsResult, expected) {
			t.Fatalf("Expected %s, but couldn't find it from result %s", expected, metricsResult)
		}

	}
}

// Helper function to get string showing in metrics of given test case and its count
func getPacketDropsCountMetricsString(testCase TestCase, count int) string {
	// tags must be in alphabetical order
	return fmt.Sprintf("packet_drops_count{dstNamesapce=\"%s\",dstPod=\"%s\",srcNamespace=\"%s\",srcPod=\"%s\"} %v", testCase.dstNamespace, testCase.dstPodName, testCase.srcNamespace, testCase.srcPodName, count)
}

// Helper function to request content body from the handler.
func requestContentBody(handler http.Handler) string {
	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w.Body.String()
}
