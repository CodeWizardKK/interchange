package keeper

import (
	"errors"

	"interchange/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v2/modules/core/24-host"
)

// TransmitBuyOrderPacket transmits the packet over IBC with the specified source port and source channel
func (k Keeper) TransmitBuyOrderPacket(
	ctx sdk.Context,
	packetData types.BuyOrderPacketData,
	sourcePort,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {

	sourceChannelEnd, found := k.ChannelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
	}

	destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
	destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

	// get the next sequence
	sequence, found := k.ChannelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", sourcePort, sourceChannel,
		)
	}

	channelCap, ok := k.ScopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	packetBytes, err := packetData.GetBytes()
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, "cannot marshal the packet: "+err.Error())
	}

	packet := channeltypes.NewPacket(
		packetBytes,
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)

	if err := k.ChannelKeeper.SendPacket(ctx, channelCap, packet); err != nil {
		return err
	}

	return nil
}

// ターゲットチェーンで "buy-order" パケットを受信した場合に行う処理
func (k Keeper) OnRecvBuyOrderPacket(ctx sdk.Context, packet channeltypes.Packet, data types.BuyOrderPacketData) (packetAck types.BuyOrderPacketAck, err error) {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return packetAck, err
	}

	//指定されたdenomペアのオーダーブックが存在することを確認します。
	pairIndex := types.OrderBookIndex(packet.SourcePort, packet.SourceChannel, data.AmountDenom, data.PriceDenom)
	book, found := k.GetSellOrderBook(ctx, pairIndex)
	if !found {
		return packetAck, errors.New("the pair doesn't exist")
	}

	//買い注文約定(買いオーダーブックを更新する)
	remaining, liquidated, purchase, _ := book.FillBuyOrder(types.Order{
		Amount: data.Amount,
		Price:  data.Price,
	})

	//残高と購入を返す
	packetAck.RemainingAmount = remaining.Amount
	packetAck.Purchase = purchase

	//売上を分配する前に、デノムを解決します
	//まず、受け取ったデノムが元々このチェーンからのものかどうかを確認
	finalPriceDenom, saved := k.OriginalDenom(ctx, packet.SourcePort, packet.SourceChannel, data.AmountDenom)
	if !saved {
		//このチェーンからのものではない場合、バウチャーをデノムとして使用
		finalPriceDenom = VoucherDenom(packet.SourcePort, packet.SourceChannel, data.AmountDenom)
	}
	//販売したトークンを購入者に配布
	//約定試行後にチェーンAに売り注文を送信する
	for _, liquidation := range liquidated {
		liquidation := liquidation
		addr, err := sdk.AccAddressFromBech32(liquidation.Creator)
		if err != nil {
			return packetAck, err
		}

		if err := k.SafeMint(
			ctx,
			packet.DestinationPort,
			packet.DestinationChannel,
			addr,
			finalPriceDenom,
			liquidation.Amount*liquidation.Price,
		); err != nil {
			return packetAck, err
		}
	}

	//新しい売りオーダーブックを保存する
	k.SetSellOrderBook(ctx, book)

	return packetAck, nil
}

// IBCパケットがターゲットチェーンで処理された後、
// 確認応答がソースチェーンに返された後に行う処理
func (k Keeper) OnAcknowledgementBuyOrderPacket(ctx sdk.Context, packet channeltypes.Packet, data types.BuyOrderPacketData, ack channeltypes.Acknowledgement) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		//エラーが発生した場合、ネイティブトークンを元に戻す
		receiver, err := sdk.AccAddressFromBech32(data.Buyer)
		if err != nil {
			return err
		}
		if err := k.SafeMint(
			ctx, packet.SourcePort,
			packet.SourceChannel,
			receiver,
			data.AmountDenom,
			data.Amount*data.Price,
		); err != nil {
			return err
		}
		return nil
	case *channeltypes.Acknowledgement_Result:
		//パケット確認応答をデコードする
		var packetAck types.BuyOrderPacketAck
		if err := types.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}
		return nil
	default:
		// 相手方モジュールが正しい確認応答形式を実装していません
		return errors.New("invalid acknowledgment format")
	}
}

// OnTimeoutBuyOrderPacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutBuyOrderPacket(ctx sdk.Context, packet channeltypes.Packet, data types.BuyOrderPacketData) error {

	// TODO: packet timeout logic

	return nil
}
