// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	chainhash "github.com/btcsuite/btcd/chaincfg/chainhash"
	clients "github.com/gonative-cc/relayer/bitcoinspv/clients"

	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/gonative-cc/relayer/bitcoinspv/types"

	wire "github.com/btcsuite/btcd/wire"
)

// MockBitcoinSPV is an autogenerated mock type for the BitcoinSPV type
type MockBitcoinSPV struct {
	mock.Mock
}

type MockBitcoinSPV_Expecter struct {
	mock *mock.Mock
}

func (_m *MockBitcoinSPV) EXPECT() *MockBitcoinSPV_Expecter {
	return &MockBitcoinSPV_Expecter{mock: &_m.Mock}
}

// ContainsBlock provides a mock function with given fields: ctx, blockHash
func (_m *MockBitcoinSPV) ContainsBlock(ctx context.Context, blockHash chainhash.Hash) (bool, error) {
	ret := _m.Called(ctx, blockHash)

	if len(ret) == 0 {
		panic("no return value specified for ContainsBlock")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, chainhash.Hash) (bool, error)); ok {
		return rf(ctx, blockHash)
	}
	if rf, ok := ret.Get(0).(func(context.Context, chainhash.Hash) bool); ok {
		r0 = rf(ctx, blockHash)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, chainhash.Hash) error); ok {
		r1 = rf(ctx, blockHash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockBitcoinSPV_ContainsBlock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ContainsBlock'
type MockBitcoinSPV_ContainsBlock_Call struct {
	*mock.Call
}

// ContainsBlock is a helper method to define mock.On call
//   - ctx context.Context
//   - blockHash chainhash.Hash
func (_e *MockBitcoinSPV_Expecter) ContainsBlock(ctx interface{}, blockHash interface{}) *MockBitcoinSPV_ContainsBlock_Call {
	return &MockBitcoinSPV_ContainsBlock_Call{Call: _e.mock.On("ContainsBlock", ctx, blockHash)}
}

func (_c *MockBitcoinSPV_ContainsBlock_Call) Run(run func(ctx context.Context, blockHash chainhash.Hash)) *MockBitcoinSPV_ContainsBlock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(chainhash.Hash))
	})
	return _c
}

func (_c *MockBitcoinSPV_ContainsBlock_Call) Return(_a0 bool, _a1 error) *MockBitcoinSPV_ContainsBlock_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockBitcoinSPV_ContainsBlock_Call) RunAndReturn(run func(context.Context, chainhash.Hash) (bool, error)) *MockBitcoinSPV_ContainsBlock_Call {
	_c.Call.Return(run)
	return _c
}

// GetLatestBlockInfo provides a mock function with given fields: ctx
func (_m *MockBitcoinSPV) GetLatestBlockInfo(ctx context.Context) (*clients.BlockInfo, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestBlockInfo")
	}

	var r0 *clients.BlockInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*clients.BlockInfo, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *clients.BlockInfo); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*clients.BlockInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockBitcoinSPV_GetLatestBlockInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLatestBlockInfo'
type MockBitcoinSPV_GetLatestBlockInfo_Call struct {
	*mock.Call
}

// GetLatestBlockInfo is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockBitcoinSPV_Expecter) GetLatestBlockInfo(ctx interface{}) *MockBitcoinSPV_GetLatestBlockInfo_Call {
	return &MockBitcoinSPV_GetLatestBlockInfo_Call{Call: _e.mock.On("GetLatestBlockInfo", ctx)}
}

func (_c *MockBitcoinSPV_GetLatestBlockInfo_Call) Run(run func(ctx context.Context)) *MockBitcoinSPV_GetLatestBlockInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockBitcoinSPV_GetLatestBlockInfo_Call) Return(_a0 *clients.BlockInfo, _a1 error) *MockBitcoinSPV_GetLatestBlockInfo_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockBitcoinSPV_GetLatestBlockInfo_Call) RunAndReturn(run func(context.Context) (*clients.BlockInfo, error)) *MockBitcoinSPV_GetLatestBlockInfo_Call {
	_c.Call.Return(run)
	return _c
}

// InsertHeaders provides a mock function with given fields: ctx, blockHeaders
func (_m *MockBitcoinSPV) InsertHeaders(ctx context.Context, blockHeaders []wire.BlockHeader) error {
	ret := _m.Called(ctx, blockHeaders)

	if len(ret) == 0 {
		panic("no return value specified for InsertHeaders")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []wire.BlockHeader) error); ok {
		r0 = rf(ctx, blockHeaders)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockBitcoinSPV_InsertHeaders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertHeaders'
type MockBitcoinSPV_InsertHeaders_Call struct {
	*mock.Call
}

// InsertHeaders is a helper method to define mock.On call
//   - ctx context.Context
//   - blockHeaders []wire.BlockHeader
func (_e *MockBitcoinSPV_Expecter) InsertHeaders(ctx interface{}, blockHeaders interface{}) *MockBitcoinSPV_InsertHeaders_Call {
	return &MockBitcoinSPV_InsertHeaders_Call{Call: _e.mock.On("InsertHeaders", ctx, blockHeaders)}
}

func (_c *MockBitcoinSPV_InsertHeaders_Call) Run(run func(ctx context.Context, blockHeaders []wire.BlockHeader)) *MockBitcoinSPV_InsertHeaders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]wire.BlockHeader))
	})
	return _c
}

func (_c *MockBitcoinSPV_InsertHeaders_Call) Return(_a0 error) *MockBitcoinSPV_InsertHeaders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockBitcoinSPV_InsertHeaders_Call) RunAndReturn(run func(context.Context, []wire.BlockHeader) error) *MockBitcoinSPV_InsertHeaders_Call {
	_c.Call.Return(run)
	return _c
}

// Stop provides a mock function with no fields
func (_m *MockBitcoinSPV) Stop() {
	_m.Called()
}

// MockBitcoinSPV_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type MockBitcoinSPV_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *MockBitcoinSPV_Expecter) Stop() *MockBitcoinSPV_Stop_Call {
	return &MockBitcoinSPV_Stop_Call{Call: _e.mock.On("Stop")}
}

func (_c *MockBitcoinSPV_Stop_Call) Run(run func()) *MockBitcoinSPV_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockBitcoinSPV_Stop_Call) Return() *MockBitcoinSPV_Stop_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockBitcoinSPV_Stop_Call) RunAndReturn(run func()) *MockBitcoinSPV_Stop_Call {
	_c.Run(run)
	return _c
}

// VerifySPV provides a mock function with given fields: ctx, spvProof
func (_m *MockBitcoinSPV) VerifySPV(ctx context.Context, spvProof *types.SPVProof) (int, error) {
	ret := _m.Called(ctx, spvProof)

	if len(ret) == 0 {
		panic("no return value specified for VerifySPV")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.SPVProof) (int, error)); ok {
		return rf(ctx, spvProof)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.SPVProof) int); ok {
		r0 = rf(ctx, spvProof)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *types.SPVProof) error); ok {
		r1 = rf(ctx, spvProof)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockBitcoinSPV_VerifySPV_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'VerifySPV'
type MockBitcoinSPV_VerifySPV_Call struct {
	*mock.Call
}

// VerifySPV is a helper method to define mock.On call
//   - ctx context.Context
//   - spvProof *types.SPVProof
func (_e *MockBitcoinSPV_Expecter) VerifySPV(ctx interface{}, spvProof interface{}) *MockBitcoinSPV_VerifySPV_Call {
	return &MockBitcoinSPV_VerifySPV_Call{Call: _e.mock.On("VerifySPV", ctx, spvProof)}
}

func (_c *MockBitcoinSPV_VerifySPV_Call) Run(run func(ctx context.Context, spvProof *types.SPVProof)) *MockBitcoinSPV_VerifySPV_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*types.SPVProof))
	})
	return _c
}

func (_c *MockBitcoinSPV_VerifySPV_Call) Return(_a0 int, _a1 error) *MockBitcoinSPV_VerifySPV_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockBitcoinSPV_VerifySPV_Call) RunAndReturn(run func(context.Context, *types.SPVProof) (int, error)) *MockBitcoinSPV_VerifySPV_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockBitcoinSPV creates a new instance of MockBitcoinSPV. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockBitcoinSPV(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBitcoinSPV {
	mock := &MockBitcoinSPV{}
	mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
