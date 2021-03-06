// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

import types "github.com/projecteru2/core/types"

// Scheduler is an autogenerated mock type for the Scheduler type
type Scheduler struct {
	mock.Mock
}

// CommonDivision provides a mock function with given fields: nodesInfo, need, total, resourceType
func (_m *Scheduler) CommonDivision(nodesInfo []types.NodeInfo, need int, total int, resourceType types.ResourceType) ([]types.NodeInfo, error) {
	ret := _m.Called(nodesInfo, need, total, resourceType)

	var r0 []types.NodeInfo
	if rf, ok := ret.Get(0).(func([]types.NodeInfo, int, int, types.ResourceType) []types.NodeInfo); ok {
		r0 = rf(nodesInfo, need, total, resourceType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.NodeInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]types.NodeInfo, int, int, types.ResourceType) error); ok {
		r1 = rf(nodesInfo, need, total, resourceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EachDivision provides a mock function with given fields: nodesInfo, need, limit, resourceType
func (_m *Scheduler) EachDivision(nodesInfo []types.NodeInfo, need int, limit int, resourceType types.ResourceType) ([]types.NodeInfo, error) {
	ret := _m.Called(nodesInfo, need, limit, resourceType)

	var r0 []types.NodeInfo
	if rf, ok := ret.Get(0).(func([]types.NodeInfo, int, int, types.ResourceType) []types.NodeInfo); ok {
		r0 = rf(nodesInfo, need, limit, resourceType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.NodeInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]types.NodeInfo, int, int, types.ResourceType) error); ok {
		r1 = rf(nodesInfo, need, limit, resourceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FillDivision provides a mock function with given fields: nodesInfo, need, limit, resourceType
func (_m *Scheduler) FillDivision(nodesInfo []types.NodeInfo, need int, limit int, resourceType types.ResourceType) ([]types.NodeInfo, error) {
	ret := _m.Called(nodesInfo, need, limit, resourceType)

	var r0 []types.NodeInfo
	if rf, ok := ret.Get(0).(func([]types.NodeInfo, int, int, types.ResourceType) []types.NodeInfo); ok {
		r0 = rf(nodesInfo, need, limit, resourceType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.NodeInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]types.NodeInfo, int, int, types.ResourceType) error); ok {
		r1 = rf(nodesInfo, need, limit, resourceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GlobalDivision provides a mock function with given fields: nodesInfo, need, total, resourceType
func (_m *Scheduler) GlobalDivision(nodesInfo []types.NodeInfo, need int, total int, resourceType types.ResourceType) ([]types.NodeInfo, error) {
	ret := _m.Called(nodesInfo, need, total, resourceType)

	var r0 []types.NodeInfo
	if rf, ok := ret.Get(0).(func([]types.NodeInfo, int, int, types.ResourceType) []types.NodeInfo); ok {
		r0 = rf(nodesInfo, need, total, resourceType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.NodeInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]types.NodeInfo, int, int, types.ResourceType) error); ok {
		r1 = rf(nodesInfo, need, total, resourceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MaxIdleNode provides a mock function with given fields: nodes
func (_m *Scheduler) MaxIdleNode(nodes []*types.Node) (*types.Node, error) {
	ret := _m.Called(nodes)

	var r0 *types.Node
	if rf, ok := ret.Get(0).(func([]*types.Node) *types.Node); ok {
		r0 = rf(nodes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*types.Node) error); ok {
		r1 = rf(nodes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SelectCPUNodes provides a mock function with given fields: nodesInfo, quota, memory
func (_m *Scheduler) SelectCPUNodes(nodesInfo []types.NodeInfo, quota float64, memory int64) ([]types.NodeInfo, map[string][]types.ResourceMap, int, error) {
	ret := _m.Called(nodesInfo, quota, memory)

	var r0 []types.NodeInfo
	if rf, ok := ret.Get(0).(func([]types.NodeInfo, float64, int64) []types.NodeInfo); ok {
		r0 = rf(nodesInfo, quota, memory)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.NodeInfo)
		}
	}

	var r1 map[string][]types.ResourceMap
	if rf, ok := ret.Get(1).(func([]types.NodeInfo, float64, int64) map[string][]types.ResourceMap); ok {
		r1 = rf(nodesInfo, quota, memory)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(map[string][]types.ResourceMap)
		}
	}

	var r2 int
	if rf, ok := ret.Get(2).(func([]types.NodeInfo, float64, int64) int); ok {
		r2 = rf(nodesInfo, quota, memory)
	} else {
		r2 = ret.Get(2).(int)
	}

	var r3 error
	if rf, ok := ret.Get(3).(func([]types.NodeInfo, float64, int64) error); ok {
		r3 = rf(nodesInfo, quota, memory)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// SelectMemoryNodes provides a mock function with given fields: nodesInfo, quota, memory
func (_m *Scheduler) SelectMemoryNodes(nodesInfo []types.NodeInfo, quota float64, memory int64) ([]types.NodeInfo, int, error) {
	ret := _m.Called(nodesInfo, quota, memory)

	var r0 []types.NodeInfo
	if rf, ok := ret.Get(0).(func([]types.NodeInfo, float64, int64) []types.NodeInfo); ok {
		r0 = rf(nodesInfo, quota, memory)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.NodeInfo)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func([]types.NodeInfo, float64, int64) int); ok {
		r1 = rf(nodesInfo, quota, memory)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func([]types.NodeInfo, float64, int64) error); ok {
		r2 = rf(nodesInfo, quota, memory)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// SelectStorageNodes provides a mock function with given fields: nodesInfo, storage
func (_m *Scheduler) SelectStorageNodes(nodesInfo []types.NodeInfo, storage int64) ([]types.NodeInfo, int, error) {
	ret := _m.Called(nodesInfo, storage)

	var r0 []types.NodeInfo
	if rf, ok := ret.Get(0).(func([]types.NodeInfo, int64) []types.NodeInfo); ok {
		r0 = rf(nodesInfo, storage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.NodeInfo)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func([]types.NodeInfo, int64) int); ok {
		r1 = rf(nodesInfo, storage)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func([]types.NodeInfo, int64) error); ok {
		r2 = rf(nodesInfo, storage)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// SelectVolumeNodes provides a mock function with given fields: nodeInfo, vbs
func (_m *Scheduler) SelectVolumeNodes(nodeInfo []types.NodeInfo, vbs types.VolumeBindings) ([]types.NodeInfo, map[string][]types.VolumePlan, int, error) {
	ret := _m.Called(nodeInfo, vbs)

	var r0 []types.NodeInfo
	if rf, ok := ret.Get(0).(func([]types.NodeInfo, types.VolumeBindings) []types.NodeInfo); ok {
		r0 = rf(nodeInfo, vbs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.NodeInfo)
		}
	}

	var r1 map[string][]types.VolumePlan
	if rf, ok := ret.Get(1).(func([]types.NodeInfo, types.VolumeBindings) map[string][]types.VolumePlan); ok {
		r1 = rf(nodeInfo, vbs)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(map[string][]types.VolumePlan)
		}
	}

	var r2 int
	if rf, ok := ret.Get(2).(func([]types.NodeInfo, types.VolumeBindings) int); ok {
		r2 = rf(nodeInfo, vbs)
	} else {
		r2 = ret.Get(2).(int)
	}

	var r3 error
	if rf, ok := ret.Get(3).(func([]types.NodeInfo, types.VolumeBindings) error); ok {
		r3 = rf(nodeInfo, vbs)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}
