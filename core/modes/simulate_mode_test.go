package modes_test

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/SpectoLabs/hoverfly/core/errors"
	"github.com/SpectoLabs/hoverfly/core/models"
	"github.com/SpectoLabs/hoverfly/core/modes"
	. "github.com/onsi/gomega"
)

type hoverflySimulateStub struct{}

func (this hoverflySimulateStub) GetResponse(requestDetails models.RequestDetails) (*models.ResponseDetails, *errors.HoverflyError) {
	if requestDetails.Destination == "positive-match.com" {
		return &models.ResponseDetails{
			Status: 200,
		}, nil
	} else {
		return nil, &errors.HoverflyError{
			Message: "matching-error",
		}
	}
}

func (this hoverflySimulateStub) ApplyMiddleware(pair models.RequestResponsePair) (models.RequestResponsePair, error) {
	if pair.Request.Path == "middleware-error" {
		return pair, fmt.Errorf("middleware-error")
	}
	return pair, nil
}

func Test_SimulateMode_WhenGivenAMatchingRequestItReturnsTheCorrectResponse(t *testing.T) {
	RegisterTestingT(t)

	unit := &modes.SimulateMode{
		Hoverfly: hoverflySimulateStub{},
	}

	request := models.RequestDetails{
		Destination: "positive-match.com",
	}

	result, err := unit.Process(nil, request)
	Expect(err).To(BeNil())

	Expect(result.Response.StatusCode).To(Equal(200))
}

func Test_SimulateMode_WhenGivenANonMatchingRequestItReturnsAnError(t *testing.T) {
	RegisterTestingT(t)

	unit := &modes.SimulateMode{
		Hoverfly: hoverflySimulateStub{},
	}

	request := models.RequestDetails{
		Destination: "negative-match.com",
	}

	result, err := unit.Process(&http.Request{}, request)
	Expect(err).ToNot(BeNil())

	Expect(result.Response.StatusCode).To(Equal(http.StatusBadGateway))

	responseBody, err := io.ReadAll(result.Response.Body)
	Expect(err).To(BeNil())

	Expect(string(responseBody)).To(ContainSubstring("There was an error when matching"))
	Expect(string(responseBody)).To(ContainSubstring("matching-error"))
}

func Test_SimulateMode_WhenGivenAMatchingRequesAndMiddlewareFaislItReturnsAnError(t *testing.T) {
	RegisterTestingT(t)

	unit := &modes.SimulateMode{
		Hoverfly: hoverflySimulateStub{},
	}

	request := models.RequestDetails{
		Destination: "positive-match.com",
		Path:        "middleware-error",
	}

	result, err := unit.Process(&http.Request{}, request)
	Expect(err).ToNot(BeNil())

	Expect(result.Response.StatusCode).To(Equal(http.StatusBadGateway))

	responseBody, err := io.ReadAll(result.Response.Body)
	Expect(err).To(BeNil())

	Expect(string(responseBody)).To(ContainSubstring("There was an error when executing middleware"))
	Expect(string(responseBody)).To(ContainSubstring("middleware-error"))
}
