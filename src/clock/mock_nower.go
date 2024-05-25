// Code generated by mockery. DO NOT EDIT.

package clock

import (
	context "context"
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// MockNower is an autogenerated mock type for the Nower type
type MockNower struct {
	mock.Mock
}

type MockNower_Expecter struct {
	mock *mock.Mock
}

func (_m *MockNower) EXPECT() *MockNower_Expecter {
	return &MockNower_Expecter{mock: &_m.Mock}
}

// Now provides a mock function with given fields:
func (_m *MockNower) Now() time.Time {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Now")
	}

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// MockNower_Now_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Now'
type MockNower_Now_Call struct {
	*mock.Call
}

// Now is a helper method to define mock.On call
func (_e *MockNower_Expecter) Now() *MockNower_Now_Call {
	return &MockNower_Now_Call{Call: _e.mock.On("Now")}
}

func (_c *MockNower_Now_Call) Run(run func()) *MockNower_Now_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockNower_Now_Call) Return(_a0 time.Time) *MockNower_Now_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockNower_Now_Call) RunAndReturn(run func() time.Time) *MockNower_Now_Call {
	_c.Call.Return(run)
	return _c
}

// WithContext provides a mock function with given fields: ctx
func (_m *MockNower) WithContext(ctx context.Context) context.Context {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for WithContext")
	}

	var r0 context.Context
	if rf, ok := ret.Get(0).(func(context.Context) context.Context); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	return r0
}

// MockNower_WithContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithContext'
type MockNower_WithContext_Call struct {
	*mock.Call
}

// WithContext is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockNower_Expecter) WithContext(ctx interface{}) *MockNower_WithContext_Call {
	return &MockNower_WithContext_Call{Call: _e.mock.On("WithContext", ctx)}
}

func (_c *MockNower_WithContext_Call) Run(run func(ctx context.Context)) *MockNower_WithContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockNower_WithContext_Call) Return(_a0 context.Context) *MockNower_WithContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockNower_WithContext_Call) RunAndReturn(run func(context.Context) context.Context) *MockNower_WithContext_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockNower creates a new instance of MockNower. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockNower(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockNower {
	mock := &MockNower{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
