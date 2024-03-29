// Code generated by mockery v2.33.3. DO NOT EDIT.

package mocks

import (
	context "context"

	order "github.com/karta0898098/mome/pkg/order"
	mock "github.com/stretchr/testify/mock"
)

// MockRepository is an autogenerated mock type for the Repository type
type MockRepository struct {
	mock.Mock
}

type MockRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRepository) EXPECT() *MockRepository_Expecter {
	return &MockRepository_Expecter{mock: &_m.Mock}
}

// FindOrder provides a mock function with given fields: ctx, id
func (_m *MockRepository) FindOrder(ctx context.Context, id string) (*order.Order, error) {
	ret := _m.Called(ctx, id)

	var r0 *order.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*order.Order, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *order.Order); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*order.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRepository_FindOrder_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindOrder'
type MockRepository_FindOrder_Call struct {
	*mock.Call
}

// FindOrder is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *MockRepository_Expecter) FindOrder(ctx interface{}, id interface{}) *MockRepository_FindOrder_Call {
	return &MockRepository_FindOrder_Call{Call: _e.mock.On("FindOrder", ctx, id)}
}

func (_c *MockRepository_FindOrder_Call) Run(run func(ctx context.Context, id string)) *MockRepository_FindOrder_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockRepository_FindOrder_Call) Return(_a0 *order.Order, err error) *MockRepository_FindOrder_Call {
	_c.Call.Return(_a0, err)
	return _c
}

func (_c *MockRepository_FindOrder_Call) RunAndReturn(run func(context.Context, string) (*order.Order, error)) *MockRepository_FindOrder_Call {
	_c.Call.Return(run)
	return _c
}

// SaveOrder provides a mock function with given fields: ctx, _a1
func (_m *MockRepository) SaveOrder(ctx context.Context, _a1 *order.Order) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *order.Order) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockRepository_SaveOrder_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SaveOrder'
type MockRepository_SaveOrder_Call struct {
	*mock.Call
}

// SaveOrder is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 *order.Order
func (_e *MockRepository_Expecter) SaveOrder(ctx interface{}, _a1 interface{}) *MockRepository_SaveOrder_Call {
	return &MockRepository_SaveOrder_Call{Call: _e.mock.On("SaveOrder", ctx, _a1)}
}

func (_c *MockRepository_SaveOrder_Call) Run(run func(ctx context.Context, _a1 *order.Order)) *MockRepository_SaveOrder_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*order.Order))
	})
	return _c
}

func (_c *MockRepository_SaveOrder_Call) Return(err error) *MockRepository_SaveOrder_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockRepository_SaveOrder_Call) RunAndReturn(run func(context.Context, *order.Order) error) *MockRepository_SaveOrder_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRepository creates a new instance of MockRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRepository {
	mock := &MockRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
