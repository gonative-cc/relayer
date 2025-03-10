// Code generated by mockery v2.53.1. DO NOT EDIT.

package mocks

import (
	chainhash "github.com/btcsuite/btcd/chaincfg/chainhash"
	clients "github.com/gonative-cc/relayer/bitcoinspv/clients"

	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/gonative-cc/relayer/bitcoinspv/types"

	wire "github.com/btcsuite/btcd/wire"
)

// BitcoinSPV is an autogenerated mock type for the BitcoinSPV type
type BitcoinSPV struct {
	mock.Mock
}

// ContainsBlock provides a mock function with given fields: ctx, blockHash
func (_m *BitcoinSPV) ContainsBlock(ctx context.Context, blockHash chainhash.Hash) (bool, error) {
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

// GetLatestBlockInfo provides a mock function with given fields: ctx
func (_m *BitcoinSPV) GetLatestBlockInfo(ctx context.Context) (*clients.BlockInfo, error) {
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

// InsertHeaders provides a mock function with given fields: ctx, blockHeaders
func (_m *BitcoinSPV) InsertHeaders(ctx context.Context, blockHeaders []wire.BlockHeader) error {
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

// Stop provides a mock function with no fields
func (_m *BitcoinSPV) Stop() {
	_m.Called()
}

// VerifySPV provides a mock function with given fields: ctx, spvProof
func (_m *BitcoinSPV) VerifySPV(ctx context.Context, spvProof *types.SPVProof) (int, error) {
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

// NewBitcoinSPV creates a new instance of BitcoinSPV. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBitcoinSPV(t interface {
	mock.TestingT
	Cleanup(func())
}) *BitcoinSPV {
	mock := &BitcoinSPV{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
