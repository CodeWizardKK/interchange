package keeper

import (
	"context"
	"errors"

	"interchange/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
)

func (k msgServer) SendSellOrder(goCtx context.Context, msg *types.MsgSendSellOrder) (*types.MsgSendSellOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//指定されたdenomペアのオーダーブックが存在することを確認します。
	pairIndex := types.OrderBookIndex(msg.Port, msg.ChannelID, msg.AmountDenom, msg.PriceDenom)
	_, found := k.GetSellOrderBook(ctx, pairIndex)
	//存在しなかった場合
	if !found {
		return &types.MsgSendSellOrderResponse{}, errors.New("the pair doesn't exist")
	}

	//送信者のアドレスを取得する
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return &types.MsgSendSellOrderResponse{}, err
	}

	//SafeBurnを使用して、新しいネイティブトークンが作成されないようにする
	if err := k.SafeBurn(
		ctx, msg.Port,
		msg.ChannelID,
		sender,
		msg.AmountDenom,
		msg.Amount,
	); err != nil {
		return &types.MsgSendSellOrderResponse{}, err
	}

	//ターゲットチェーンで受け取ったバウチャーを保存(後で元に戻すことができるようにする)
	k.SaveVoucherDenom(ctx, msg.Port, msg.ChannelID, msg.AmountDenom)

	//パケットを構築
	var packet types.SellOrderPacketData
	packet.AmountDenom = msg.AmountDenom
	packet.Amount = msg.Amount
	packet.PriceDenom = msg.PriceDenom
	packet.Price = msg.Price
	packet.Seller = msg.Creator

	//IBCパケットをターゲットチェーンに送信
	err = k.TransmitSellOrderPacket(
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

	return &types.MsgSendSellOrderResponse{}, nil
}
