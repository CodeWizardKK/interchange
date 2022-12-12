package keeper

import (
	"context"
	"errors"

	"interchange/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
)

func (k msgServer) SendBuyOrder(goCtx context.Context, msg *types.MsgSendBuyOrder) (*types.MsgSendBuyOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//ペアがオーダーブックに存在するかどうかを確認します
	pairIndex := types.OrderBookIndex(msg.Port, msg.ChannelID, msg.AmountDenom, msg.PriceDenom)
	_, found := k.GetBuyOrderBook(ctx, pairIndex)
	//存在しなかった場合
	if !found {
		return &types.MsgSendBuyOrderResponse{}, errors.New("the pair doesn't exist")
	}
	//送信者のアドレスを取得する
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return &types.MsgSendBuyOrderResponse{}, err
	}
	//トークンがIBCトークンの場合、トークンを焼却
	//トークンがネイティブトークンの場合、トークンをロック
	if err := k.SafeBurn(
		ctx, msg.Port,
		msg.ChannelID,
		sender,
		msg.AmountDenom,
		msg.Amount,
	); err != nil {
		return &types.MsgSendBuyOrderResponse{}, err
	}
	//ターゲットチェーンで受け取ったバウチャーを保存(後で元に戻すことができるようにする)
	k.SaveVoucherDenom(ctx, msg.Port, msg.ChannelID, msg.AmountDenom)

	//パケットを構築
	var packet types.BuyOrderPacketData

	packet.AmountDenom = msg.AmountDenom
	packet.Amount = msg.Amount
	packet.PriceDenom = msg.PriceDenom
	packet.Price = msg.Price
	packet.Buyer = msg.Creator

	//IBCパケットをターゲットチェーンに送信
	err = k.TransmitBuyOrderPacket(
		ctx,
		packet,
		msg.Port,
		msg.ChannelID,
		clienttypes.ZeroHeight(),
		msg.TimeoutTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendBuyOrderResponse{}, nil
}
