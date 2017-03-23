// This file was generated by counterfeiter
package monitorfakes

import (
	"net"
	"sync"

	"github.com/cloudfoundry-incubator/switchboard/runner/monitor"
)

type FakeArpEntryRemover struct {
	RemoveEntryStub        func(ip net.IP) error
	removeEntryMutex       sync.RWMutex
	removeEntryArgsForCall []struct {
		ip net.IP
	}
	removeEntryReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeArpEntryRemover) RemoveEntry(ip net.IP) error {
	fake.removeEntryMutex.Lock()
	fake.removeEntryArgsForCall = append(fake.removeEntryArgsForCall, struct {
		ip net.IP
	}{ip})
	fake.recordInvocation("RemoveEntry", []interface{}{ip})
	fake.removeEntryMutex.Unlock()
	if fake.RemoveEntryStub != nil {
		return fake.RemoveEntryStub(ip)
	} else {
		return fake.removeEntryReturns.result1
	}
}

func (fake *FakeArpEntryRemover) RemoveEntryCallCount() int {
	fake.removeEntryMutex.RLock()
	defer fake.removeEntryMutex.RUnlock()
	return len(fake.removeEntryArgsForCall)
}

func (fake *FakeArpEntryRemover) RemoveEntryArgsForCall(i int) net.IP {
	fake.removeEntryMutex.RLock()
	defer fake.removeEntryMutex.RUnlock()
	return fake.removeEntryArgsForCall[i].ip
}

func (fake *FakeArpEntryRemover) RemoveEntryReturns(result1 error) {
	fake.RemoveEntryStub = nil
	fake.removeEntryReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeArpEntryRemover) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.removeEntryMutex.RLock()
	defer fake.removeEntryMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeArpEntryRemover) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ monitor.ArpEntryRemover = new(FakeArpEntryRemover)