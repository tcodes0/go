// Code generated by mockery. DO NOT EDIT.

package reflectutil

import (
	reflect "reflect"

	mock "github.com/stretchr/testify/mock"
)

// MockFieldUpdater is an autogenerated mock type for the FieldUpdater type
type MockFieldUpdater struct {
	mock.Mock
}

type MockFieldUpdater_Expecter struct {
	mock *mock.Mock
}

func (_m *MockFieldUpdater) Expect() *MockFieldUpdater_Expecter {
	return &MockFieldUpdater_Expecter{mock: &_m.Mock}
}

// UpdateField provides a mock function with given fields: field, base
func (_m *MockFieldUpdater) UpdateField(field *reflect.StructField, base reflect.Value) error {
	ret := _m.Called(field, base)

	if len(ret) == 0 {
		panic("no return value specified for UpdateField")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*reflect.StructField, reflect.Value) error); ok {
		r0 = rf(field, base)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockFieldUpdater_UpdateField_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateField'
type MockFieldUpdater_UpdateField_Call struct {
	*mock.Call
}

// UpdateField is a helper method to define mock.On call
//   - field *reflect.StructField
//   - base reflect.Value
func (_e *MockFieldUpdater_Expecter) UpdateField(field interface{}, base interface{}) *MockFieldUpdater_UpdateField_Call {
	return &MockFieldUpdater_UpdateField_Call{Call: _e.mock.On("UpdateField", field, base)}
}

func (_c *MockFieldUpdater_UpdateField_Call) Run(run func(field *reflect.StructField, base reflect.Value)) *MockFieldUpdater_UpdateField_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*reflect.StructField), args[1].(reflect.Value))
	})
	return _c
}

func (_c *MockFieldUpdater_UpdateField_Call) Return(_a0 error) *MockFieldUpdater_UpdateField_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFieldUpdater_UpdateField_Call) RunAndReturn(run func(*reflect.StructField, reflect.Value) error) *MockFieldUpdater_UpdateField_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockFieldUpdater creates a new instance of MockFieldUpdater. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockFieldUpdater(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockFieldUpdater {
	mock := &MockFieldUpdater{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
