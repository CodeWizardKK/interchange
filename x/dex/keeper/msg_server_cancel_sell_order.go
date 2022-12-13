package keeper

import (
	"context"
	"errors"

	"interchange/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CancelSellOrder(goCtx context.Context, msg *types.MsgCancelSellOrder) (*types.MsgCancelSellOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//指定されたdenomペアのオーダーブックが存在することを確認
	pairIndex := types.OrderBookIndex(msg.Port, msg.Channel, msg.AmountDenom, msg.PriceDenom)
	//特定の売り注文表を取得する
	s, found := k.GetSellOrderBook(ctx, pairIndex)
	if !found {
		return &types.MsgCancelSellOrderResponse{}, nil
	}

	//注文IDを元に、特定の注文を取得
	order, err := s.Book.GetOrderFromID(msg.OrderID)
	if err != nil {
		return &types.MsgCancelSellOrderResponse{}, err
	}

	if order.Creator != msg.Creator {
		return &types.MsgCancelSellOrderResponse{}, errors.New("canceller must be creator")
	}

	//特定の注文を削除する
	if err := s.Book.RemoveOrderFromID(msg.OrderID); err != nil {
		return &types.MsgCancelSellOrderResponse{}, err
	}

	//ストアにセットする
	k.SetSellOrderBook(ctx, s)

	//出品者に残額を返金する
	seller, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return &types.MsgCancelSellOrderResponse{}, err
	}
	if err := k.SafeMint(
		ctx, msg.Port,
		msg.Channel,
		seller,
		msg.AmountDenom,
		order.Amount,
	); err != nil {
		return &types.MsgCancelSellOrderResponse{}, err
	}

	return &types.MsgCancelSellOrderResponse{}, nil
}
