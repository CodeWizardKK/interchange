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

// TransmitSellOrderPacket transmits the packet over IBC with the specified source port and source channel
func (k Keeper) TransmitSellOrderPacket(
	ctx sdk.Context,
	packetData types.SellOrderPacketData,
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

// ターゲットチェーンで "sell order" パケットを受信した場合に行う処理
func (k Keeper) OnRecvSellOrderPacket(ctx sdk.Context, packet channeltypes.Packet, data types.SellOrderPacketData) (packetAck types.SellOrderPacketAck, err error) {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return packetAck, err
	}

	//指定されたdenomペアのオーダーブックが存在することを確認します。
	pairIndex := types.OrderBookIndex(packet.SourcePort, packet.SourceChannel, data.AmountDenom, data.PriceDenom)
	book, found := k.GetBuyOrderBook(ctx, pairIndex)
	//存在しなかった場合
	if !found {
		//ペアは存在しません
		return packetAck, errors.New("the pair doesn't exist")
	}

	//売り注文約定(売りオーダーブックを更新する)
	remaining, liquidated, gain, _ := book.FillSellOrder(types.Order{
		Amount: data.Amount,
		Price:  data.Price,
	})

	//残高と利益を返す
	packetAck.RemainingAmount = remaining.Amount
	packetAck.Gain = gain

	//売上を分配する前に、デノムを解決します
	//まず、受け取ったデノムが元々このチェーンからのものかどうかを確認
	finalAmountDenom, saved := k.OriginalDenom(ctx, packet.SourcePort, packet.SourceChannel, data.AmountDenom)
	if !saved {
		//このチェーンからのものではない場合、バウチャーをデノムとして使用
		VoucherDenom(packet.SourcePort, packet.SourceChannel, data.AmountDenom)
	}

	//販売したトークンを購入者に配布
	//約定試行後、チェーンAに売り注文を送信します。
	for _, liquidation := range liquidated {
		liquidation := liquidation
		addr, err := sdk.AccAddressFromBech32(liquidation.Creator)
		if err != nil {
			return packetAck, err
		}
		if err = k.SafeMint(
			ctx,
			packet.DestinationPort,
			packet.DestinationChannel,
			addr,
			finalAmountDenom,
			liquidation.Amount,
		); err != nil {
			return packetAck, err
		}
	}

	//新しい買いオーダーブックを保存する
	k.SetBuyOrderBook(ctx, book)

	return packetAck, nil
}

// IBCパケットがターゲットチェーンで処理された後、
// 確認応答がソースチェーンに返された後に行う処理
func (k Keeper) OnAcknowledgementSellOrderPacket(ctx sdk.Context, packet channeltypes.Packet, data types.SellOrderPacketData, ack channeltypes.Acknowledgement) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		//エラーが発生した場合、ネイティブトークンを元に戻す
		receiver, err := sdk.AccAddressFromBech32(data.Seller)
		if err != nil {
			return err
		}
		if err := k.SafeMint(
			ctx,
			packet.SourcePort,
			packet.SourceChannel,
			receiver,
			data.AmountDenom,
			data.Amount,
		); err != nil {
			return err
		}
		return nil
	case *channeltypes.Acknowledgement_Result:
		//パケット確認応答をデコードする
		var packetAck types.SellOrderPacketAck
		if err := types.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}

		//残りの売り注文を、売り注文帳に保管
		pairIndex := types.OrderBookIndex(packet.SourcePort, packet.SourceChannel, data.AmountDenom, data.PriceDenom)
		book, found := k.GetSellOrderBook(ctx, pairIndex)
		if !found {
			//"売り注文帳が存在する必要があります"
			panic("sell order book must exist")
		}

		//販売されたトークンを購入者に配布
		//売り手に販売された金額の価格を分配
		// 注文の残りの金額を追加する
		if packetAck.RemainingAmount > 0 {
			_, err := book.AppendOrder(data.Seller, packetAck.RemainingAmount, data.Price)
			if err != nil {
				return err
			}
			// 新しいオーダーブックを保存する
			k.SetSellOrderBook(ctx, book)
		}

		//エラーが発生した場合、焼き付けられたトークンをミント
		if packetAck.Gain > 0 {
			receiver, err := sdk.AccAddressFromBech32(data.Seller)
			if err != nil {
				return err
			}
			finalPriceDenom, saved := k.OriginalDenom(ctx, packet.SourcePort, packet.SourceChannel, data.PriceDenom)
			if !saved {
				// このチェーンからのものではない場合、バウチャーをデノムとして使用します
				finalPriceDenom = VoucherDenom(packet.DestinationPort, packet.DestinationChannel, data.PriceDenom)
			}

			if err := k.SafeMint(ctx, packet.SourcePort, packet.SourceChannel, receiver, finalPriceDenom, packetAck.Gain); err != nil {
				return err
			}
		}

		return nil
	default:
		// 相手方モジュールが正しい確認応答形式を実装していません
		return errors.New("invalid acknowledgment format")
	}
}

// OnTimeoutSellOrderPacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutSellOrderPacket(ctx sdk.Context, packet channeltypes.Packet, data types.SellOrderPacketData) error {

	// TODO: packet timeout logic

	return nil
}
