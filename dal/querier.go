// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package dal

import (
	"context"
)

type Querier interface {
	GetBitcoinTx(ctx context.Context, arg *GetBitcoinTxParams) (*BitcoinTx, error)
	GetBitcoinTxsToBroadcast(ctx context.Context, status BitcoinTxStatus) ([]*IkaSignRequest, error)
	GetBroadcastedBitcoinTxsInfo(ctx context.Context) ([]*GetBroadcastedBitcoinTxsInfoRow, error)
	GetIkaSignRequestByID(ctx context.Context, id uint64) (*IkaSignRequest, error)
	GetIkaSignRequestWithStatus(ctx context.Context, id uint64) (*GetIkaSignRequestWithStatusRow, error)
	GetIkaTx(ctx context.Context, arg *GetIkaTxParams) (*IkaTx, error)
	GetPendingIkaSignRequests(ctx context.Context) ([]*IkaSignRequest, error)
	InsertBtcTx(ctx context.Context, arg *InsertBtcTxParams) error
	InsertIkaSignRequest(ctx context.Context, arg *InsertIkaSignRequestParams) error
	InsertIkaTx(ctx context.Context, arg *InsertIkaTxParams) error
	UpdateBitcoinTxToConfirmed(ctx context.Context, arg *UpdateBitcoinTxToConfirmedParams) error
	UpdateIkaSignRequestFinalSig(ctx context.Context, arg *UpdateIkaSignRequestFinalSigParams) error
}

var _ Querier = (*Queries)(nil)
