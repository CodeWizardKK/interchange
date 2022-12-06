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

// TransmitCreatePairPacketは、指定されたソースポートとソースチャネルを使用してIBC経由でパケットを送信します。
func (k Keeper) TransmitCreatePairPacket(
	ctx sdk.Context,
	packetData types.CreatePairPacketData,
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

// パケット受信を処理
// IBCパケットがターゲットチェーンで受信されると、send-create-pairコマンドは買い注文書を作成する
func (k Keeper) OnRecvCreatePairPacket(ctx sdk.Context, packet channeltypes.Packet, data types.CreatePairPacketData) (packetAck types.CreatePairPacketAck, err error) {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return packetAck, err
	}

	//IBCパケットがターゲットチェーンで受信されると、
	//モジュールは買い注文書が既に存在するかどうかを確認する必要がある。
	pairIndex := types.OrderBookIndex(packet.SourcePort, packet.SourceChannel, data.SourceDenom, data.TargetDenom)
	_, found := k.GetBuyOrderBook(ctx, pairIndex)
	//買い注文書(ペア)が既に存在している場合
	if found {
		return packetAck, errors.New("the pair already exist")
	}
	//買い注文書が存在しなかった場合、指定されたdenomsの買い注文書を作成
	book := types.NewBuyOrderBook(data.SourceDenom, data.TargetDenom)
	//OrderBookIndexの割り当て
	book.Index = pairIndex
	//買い注文ストアに保存
	k.SetBuyOrderBook(ctx, book)

	return packetAck, nil
}

// パケットの成功または失敗に応答します
// ソースチェーンでIBC確認が受信されると、send-create-pairコマンドは売り注文書を作成
func (k Keeper) OnAcknowledgementCreatePairPacket(ctx sdk.Context, packet channeltypes.Packet, data types.CreatePairPacketData, ack channeltypes.Acknowledgement) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		return nil
	case *channeltypes.Acknowledgement_Result:
		// Decode the packet acknowledgment
		var packetAck types.CreatePairPacketAck
		if err := types.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}

		//売り注文書を作成
		pairIndex := types.OrderBookIndex(packet.SourcePort, packet.SourceChannel, data.SourceDenom, data.TargetDenom)
		book := types.NewSellOrderBook(data.SourceDenom, data.TargetDenom)
		book.Index = pairIndex
		k.SetSellOrderBook(ctx, book)

		return nil
	default:
		// 相手方モジュールが正しい確認応答形式を実装していない場合
		return errors.New("invalid acknowledgment format")
	}
}

// OnTimeoutCreatePairPacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutCreatePairPacket(ctx sdk.Context, packet channeltypes.Packet, data types.CreatePairPacketData) error {

	// TODO: packet timeout logic

	return nil
}
