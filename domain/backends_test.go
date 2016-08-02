package domain_test

import (
	"github.com/cloudfoundry-incubator/switchboard/config"
	"github.com/cloudfoundry-incubator/switchboard/domain"
	"github.com/cloudfoundry-incubator/switchboard/models/modelsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/cloudfoundry-incubator/switchboard/models"
)

var _ = Describe("Backends", func() {
	var (
		backends        *domain.Backends
		backendProvider func(name string, host string, port uint, statusPort uint, statusEndpoint string, logger lager.Logger) models.Backend
	)

	var backendChanToSlice = func(c <-chan models.Backend) []models.Backend {
		var result []models.Backend
		for b := range c {
			result = append(result, b)
		}
		return result
	}
	BeforeEach(func() {
		backendProvider = domain.BackendProvider
	})

	JustBeforeEach(func() {
		logger := lagertest.NewTestLogger("Backends test")

		backendConfigs := []config.Backend{
			{"localhost", 50000, 60000, "status", "backend-0"},
			{"localhost", 50001, 60001, "status", "backend-1"},
			{"localhost", 50002, 60002, "status", "backend-2"},
		}

		backends = domain.NewBackends(backendConfigs, logger)
	})

	AfterEach(func() {
		domain.BackendProvider = backendProvider
	})

	Describe("Concurrent operations", func() {
		It("do not result in a race", func() {
			readySetGo := make(chan interface{})

			doneChans := []chan interface{}{
				make(chan interface{}),
				make(chan interface{}),
				make(chan interface{}),
				make(chan interface{}),
				make(chan interface{}),
			}

			backend := backendChanToSlice(backends.All())[0]

			go func() {
				<-readySetGo
				backends.All()
				close(doneChans[0])
			}()

			go func() {
				<-readySetGo
				backends.Active()
				close(doneChans[1])
			}()

			go func() {
				<-readySetGo
				backends.SetHealthy(backend)
				close(doneChans[2])
			}()

			go func() {
				<-readySetGo
				backends.SetUnhealthy(backend)
				close(doneChans[3])
			}()

			go func() {
				<-readySetGo
				backends.Healthy()
				close(doneChans[4])
			}()

			close(readySetGo)

			for _, done := range doneChans {
				<-done
			}
		})

	})

	Describe("All", func() {
		It("allows iterating over all the backends", func() {
			backendsSeen := []string{}
			for backend := range backends.All() {
				backendsSeen = append(backendsSeen, backend.HealthcheckUrl())
			}

			Expect(backendsSeen).To(ContainElement("http://localhost:60000/status"))
			Expect(backendsSeen).To(ContainElement("http://localhost:60001/status"))
			Expect(backendsSeen).To(ContainElement("http://localhost:60002/status"))
		})
	})

	Describe("Healthy", func() {
		It("allows iterating over only the healthy backends", func() {
			healthy := backendChanToSlice(backends.Healthy())
			numHealthy := 3
			Expect(len(healthy)).To(Equal(numHealthy))

			for _, b := range healthy {
				backends.SetUnhealthy(b)
				numHealthy--
				healthy = backendChanToSlice(backends.Healthy())
				Expect(len(healthy)).To(Equal(numHealthy))
			}
		})
	})

	Describe("SetHealthy", func() {
		var unhealthy models.Backend

		JustBeforeEach(func() {
			unhealthy = backendChanToSlice(backends.Healthy())[0]
			backends.SetUnhealthy(unhealthy)
		})

		It("sets the backend to be healthy", func() {
			Expect(len(backendChanToSlice(backends.Healthy()))).To(Equal(2))
			backends.SetHealthy(unhealthy)
			Expect(len(backendChanToSlice(backends.Healthy()))).To(Equal(3))
		})

		Context("when all backends are unhealthy and there is no active backend", func() {
			JustBeforeEach(func() {
				for backend := range backends.Healthy() {
					backends.SetUnhealthy(backend)
				}
			})

			It("sets the newly healthy backend as the new active backend", func() {
				Expect(backends.Active()).To(BeNil())
				backend := backendChanToSlice(backends.All())[0]
				backends.SetHealthy(backend)
				Expect(backends.Active()).To(Equal(backend))
			})
		})
	})

	Describe("SetUnhealthy", func() {
		It("sets the backend to be unhealthy", func() {
			backend := backendChanToSlice(backends.Healthy())[0]
			Expect(len(backendChanToSlice(backends.Healthy()))).To(Equal(3))
			backends.SetUnhealthy(backend)
			Expect(len(backendChanToSlice(backends.Healthy()))).To(Equal(2))
		})

		Context("when this is active", func() {
			BeforeEach(func() {
				domain.BackendProvider = func(string, string, uint, uint, string, lager.Logger) models.Backend {
					return new(modelsfakes.FakeBackend)
				}
			})

			It("severs all open connections", func() {
				backend := backends.Active()
				backends.SetUnhealthy(backend)
				Expect(backend.(*modelsfakes.FakeBackend).SeverConnectionsCallCount()).To(Equal(1))
			})

			It("sets another healthy backend as the new active backend", func() {
				numHealthy := len(backendChanToSlice(backends.Healthy()))
				for _ = range backends.Healthy() {
					previousActive := backends.Active()
					backends.SetUnhealthy(previousActive)
					nextActive := backends.Active()
					Expect(nextActive).ToNot(Equal(previousActive))

					numHealthy--
					if numHealthy > 0 { // more healthy backends
						Expect(backends.Active()).ToNot(BeNil())
					} else { // no more healthy backends -> no active backend
						Expect(backends.Active()).To(BeNil())
					}
				}
			})
		})
	})
})
