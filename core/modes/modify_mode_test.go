package modes_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/SpectoLabs/hoverfly/core/models"
	"github.com/SpectoLabs/hoverfly/core/modes"
	. "github.com/onsi/gomega"
)

type hoverflyModifyStub struct{}

func (this hoverflyModifyStub) DoRequest(request *http.Request) (*http.Response, *time.Duration, error) {
	response := &http.Response{}
	if request.Host == "error.com" {
		return nil, nil, errors.New("Could not reach error.com")
	}

	request.Host = "modified.com"

	response.StatusCode = 200
	response.Body = io.NopCloser(bytes.NewBufferString("test"))

	duration := 1 * time.Second
	return response, &duration, nil
}

func (this hoverflyModifyStub) ApplyMiddleware(pair models.RequestResponsePair) (models.RequestResponsePair, error) {
	if pair.Request.Path == "/middleware-error" {
		return pair, errors.New("middleware-error")
	}
	pair.Response.Body = "modified by test middleware"
	return pair, nil
}

func Test_ModifyMode_WhenGivenARequestItWillModifyTheRequestAndExecuteIt(t *testing.T) {
	RegisterTestingT(t)

	hoverflyStub := &hoverflyModifyStub{}

	unit := &modes.ModifyMode{
		Hoverfly: hoverflyStub,
	}

	requestDetails := models.RequestDetails{
		Scheme:      "http",
		Destination: "positive-match.com",
	}

	request, err := http.NewRequest("GET", "http://positive-match.com", nil)
	Expect(err).To(BeNil())

	result, err := unit.Process(request, requestDetails)
	Expect(err).To(BeNil())

	Expect(result.Response.StatusCode).To(Equal(200))
	Expect(result.Response.Request.Host).To(Equal("modified.com"))

	responseBody, err := io.ReadAll(result.Response.Body)
	Expect(err).To(BeNil())

	Expect(string(responseBody)).To(Equal("modified by test middleware"))
}

func Test_ModifyMode_WhenGivenABadRequestItWillError(t *testing.T) {
	RegisterTestingT(t)

	hoverflyStub := &hoverflyModifyStub{}

	unit := &modes.ModifyMode{
		Hoverfly: hoverflyStub,
	}

	requestDetails := models.RequestDetails{
		Scheme:      "http",
		Destination: "error.com",
	}

	request, err := http.NewRequest("GET", "http://error.com", nil)
	Expect(err).To(BeNil())

	result, err := unit.Process(request, requestDetails)
	Expect(err).ToNot(BeNil())

	Expect(result.Response.StatusCode).To(Equal(http.StatusBadGateway))

	responseBody, err := io.ReadAll(result.Response.Body)
	Expect(err).To(BeNil())

	Expect(string(responseBody)).To(ContainSubstring("There was an error when forwarding the request to the intended destination"))
	Expect(string(responseBody)).To(ContainSubstring("Could not reach error.com"))
}

func Test_ModifyMode_WillErrorWhenMiddlewareFails(t *testing.T) {
	RegisterTestingT(t)

	hoverflyStub := &hoverflyModifyStub{}

	unit := &modes.ModifyMode{
		Hoverfly: hoverflyStub,
	}

	requestDetails := models.RequestDetails{
		Scheme:      "http",
		Destination: "test.com",
		Path:        "/middleware-error",
	}

	request, err := http.NewRequest("GET", "http://test.com/middleware-error", nil)
	Expect(err).To(BeNil())

	result, err := unit.Process(request, requestDetails)
	Expect(err).ToNot(BeNil())

	Expect(result.Response.StatusCode).To(Equal(http.StatusBadGateway))

	responseBody, err := io.ReadAll(result.Response.Body)
	Expect(err).To(BeNil())

	Expect(string(responseBody)).To(ContainSubstring("There was an error when executing middleware"))
	Expect(string(responseBody)).To(ContainSubstring("middleware-error"))
}
