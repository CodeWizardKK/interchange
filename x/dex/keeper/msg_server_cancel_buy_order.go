package keeper

import (
	"context"
	"errors"

	"interchange/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CancelBuyOrder(goCtx context.Context, msg *types.MsgCancelBuyOrder) (*types.MsgCancelBuyOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//指定されたdenomペアのオーダーブックが存在することを確認
	pairIndex := types.OrderBookIndex(msg.Port, msg.Channel, msg.AmountDenom, msg.PriceDenom)
	//特定の買い注文表を取得する
	b, found := k.GetBuyOrderBook(ctx, pairIndex)
	if !found {
		return &types.MsgCancelBuyOrderResponse{}, errors.New("the pair doesn't exist")
	}

	//注文IDを元に、特定の注文を取得
	order, err := b.Book.GetOrderFromID(msg.OrderID)
	if err != nil {
		return &types.MsgCancelBuyOrderResponse{}, err
	}

	if msg.Creator != order.Creator {
		return &types.MsgCancelBuyOrderResponse{}, errors.New("canceller must be creator")
	}

	//特定の注文を削除する
	if err := b.Book.RemoveOrderFromID(msg.OrderID); err != nil {
		return &types.MsgCancelBuyOrderResponse{}, err
	}

	//ストアにセットする
	k.SetBuyOrderBook(ctx, b)

	//購入者に残額を返金する
	buyer, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return &types.MsgCancelBuyOrderResponse{}, err
	}
	if err := k.SafeMint(
		ctx, msg.Port,
		msg.Channel,
		buyer,
		msg.AmountDenom,
		order.Amount*order.Price,
	); err != nil {
		return &types.MsgCancelBuyOrderResponse{}, err
	}

	return &types.MsgCancelBuyOrderResponse{}, nil
}
