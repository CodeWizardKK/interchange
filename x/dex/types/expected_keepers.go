package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	//AccAddress => ModuleAccount コインを送る
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	//ModuleAccount => AccAddress コインを送る
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	//BurnCoinsはコインを燃やし、モジュールアカウントの残高からコインを削除
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	//MintCoinsはどこからともなく新しいコインを作成し、それをモジュールアカウントに追加
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}
