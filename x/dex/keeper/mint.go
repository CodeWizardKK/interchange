package keeper

import (
	"fmt"
	"strings"

	//	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"

	"interchange/x/dex/types"
)

// トークンがIBCバウチャートークンかのチェック
func isIBCToken(denom string) bool {
	//バウチャートークンは、"ibc/B5CB286...A7B21307F"のようなdenomを持つ
	return strings.HasPrefix(denom, "ibc/")
}

// トークンがIBCバウチャーである場合はトークンを燃焼、トークンがチェーンにネイティブである場合はトークンをロックする
func (k Keeper) SafeBurn(ctx sdk.Context, port string, channel string, sender sdk.AccAddress, denom string, amount int32) error {
	if isIBCToken(denom) {
		//トークンの燃焼
		if err := k.BurnTokens(ctx, sender, sdk.NewCoin(denom, sdk.NewInt(int64(amount)))); err != nil {
			return err
		}
	} else {
		// トークンをロック
		if err := k.LockTokens(ctx, port, channel, sender, sdk.NewCoin(denom, sdk.NewInt(int64(amount)))); err != nil {
			return err
		}
	}

	return nil
}

// トークンの燃焼(IBCバウチャートークン)
func (k Keeper) BurnTokens(ctx sdk.Context, sender sdk.AccAddress, tokens sdk.Coin) error {
	// コインをモジュールアカウントに転送
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(tokens)); err != nil {
		return err
	}
	//コインを燃焼
	if err := k.bankKeeper.BurnCoins(
		ctx, types.ModuleName, sdk.NewCoins(tokens),
	); err != nil {
		// 上記の手順でモジュール アカウントが取得され、書き込むのに十分な残高があるため、この問題は発生しません。
		panic(fmt.Sprintf("cannot burn coins after a successful send to a module account: %v", err))
		//"モジュールアカウントへの送信が成功した後、コインを燃やすことができません"
	}

	return nil
}

// トークンをロック(IBCバウチャートークンではない、もう一方のネイティブトークン)
func (k Keeper) LockTokens(ctx sdk.Context, sourcePort string, sourceChannel string, sender sdk.AccAddress, tokens sdk.Coin) error {
	// ネイティブトークンのエスクローアドレスを作成する(トークンをロックする為)
	escrowAddress := ibctransfertypes.GetEscrowAddress(sourcePort, sourceChannel)

	//ネイティブトークンをエスクローアドレスに送信(ロック)
	if err := k.bankKeeper.SendCoins(
		ctx, sender, escrowAddress, sdk.NewCoins(tokens),
	); err != nil {
		//残高不足だと失敗する
		return err
	}

	return nil
}

// トークンがIBCバウチャートークン(ibc/....) である場合、MintTokens(トークンを受信者のアカウントに送信する)
// それ以外の場合は、UnlockTokens(ネイティブトークンのロックを解除)
func (k Keeper) SafeMint(ctx sdk.Context, port string, channel string, receiver sdk.AccAddress, denom string, amount int32) error {
	//IBCバウチャートークンの場合
	if isIBCToken(denom) {
		//トークンを受信者のアカウントに送信する
		if err := k.MintTokens(ctx, receiver, sdk.NewCoin(denom, sdk.NewInt(int64(amount)))); err != nil {
			return err
		}
	} else {
		//ネイティブトークンのロックを解除
		if err := k.UnlockTokens(
			ctx,
			port,
			channel,
			receiver,
			sdk.NewCoin(denom, sdk.NewInt(int64(amount))),
		); err != nil {
			return err
		}
	}

	return nil
}

// トークンを受信者のアカウントに送信する
func (k Keeper) MintTokens(ctx sdk.Context, receiver sdk.AccAddress, tokens sdk.Coin) error {
	//転送元が同じチェーンの場合、新しいトークンを作成
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(tokens)); err != nil {
		return err
	}

	//トークンを送る
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, sdk.NewCoins(tokens)); err != nil {
		panic(fmt.Sprintf("unable to send coins from module to account despite previously minting coins to module account: %v", err))
	}

	return nil
}

// ネイティブブロックチェーンに送り返された後にトークンをロック解除する
func (k Keeper) UnlockTokens(ctx sdk.Context, sourcePort string, sourceChannel string, receiver sdk.AccAddress, tokens sdk.Coin) error {
	// ネイティブトークンのエスクローアドレスを作成する(トークンをロック解除する為)
	escrowAddress := ibctransfertypes.GetEscrowAddress(sourcePort, sourceChannel)

	//エスクローアドレスから受信者へネイティブトークンが送り返される(ロック解除)
	if err := k.bankKeeper.SendCoins(
		ctx, escrowAddress, receiver, sdk.NewCoins(tokens),
	); err != nil {
		//残高不足だと失敗する
		return err
	}

	return nil
}
