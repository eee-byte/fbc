package ica

import (
	"fmt"

	sdk "github.com/fibonacci-chain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/fibonacci-chain/fbc/libs/cosmos-sdk/types/errors"
	controllerkeeper "github.com/fibonacci-chain/fbc/libs/ibc-go/modules/apps/27-interchain-accounts/controller/keeper"
	hostkeeper "github.com/fibonacci-chain/fbc/libs/ibc-go/modules/apps/27-interchain-accounts/host/keeper"
	tmtypes "github.com/fibonacci-chain/fbc/libs/tendermint/types"
)

func NewHandler(hostKeeper *hostkeeper.Keeper, ck *controllerkeeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		if !tmtypes.HigherThanVenus4(ctx.BlockHeight()) {
			errMsg := fmt.Sprintf("ibc ica is not supported at height %d", ctx.BlockHeight())
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}

		ctx.SetEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized  ica message type: %T", msg)
		}

	}
}