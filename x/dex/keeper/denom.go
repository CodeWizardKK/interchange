package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"

	"interchange/x/dex/types"
)

// denomのバウチャーを保存する(後で元に戻すことができるようにする)
func (k Keeper) SaveVoucherDenom(ctx sdk.Context, port string, channel string, denom string) {
	voucher := VoucherDenom(port, channel, denom)

	// denomTraceを取得
	_, saved := k.GetDenomTrace(ctx, voucher)
	//存在しない(保存されていない)場合のみ、保存
	if !saved {
		k.SetDenomTrace(ctx, types.DenomTrace{
			Index:   voucher,
			Port:    port,
			Channel: channel,
			Origin:  denom,
		})
	}
}

// ポートIDとチャネルIDからdenomのバウチャーを返す
func VoucherDenom(port string, channel string, denom string) string {
	//sourcePrefixを取得
	sourcePrefix := ibctransfertypes.GetDenomPrefix(port, channel)

	// prefixedDenomを取得(sourcePrefixの末尾に「/」が含まれている)
	prefixedDenom := sourcePrefix + denom

	// construct the denomination trace from the full raw denomination
	denomTrace := ibctransfertypes.ParseDenomTrace(prefixedDenom)
	voucher := denomTrace.IBCDenom()
	return voucher[:16]
}

// バウチャーの元のdenomを返す
func (k Keeper) OriginalDenom(ctx sdk.Context, port string, channel string, voucher string) (string, bool) {
	//denomTraceを取得
	trace, exist := k.GetDenomTrace(ctx, voucher)
	//存在した場合
	if exist {
		if trace.Port == port && trace.Channel == channel {
			//オリジン(denom)とtrueを返す
			return trace.Origin, true
		}
	}
	//存在しない(指定されたポートIDとチャネルIDがバウチャーのオリジンでない)場合は、""とfalseを返す
	return "", false
}
