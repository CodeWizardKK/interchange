package types

func NewBuyOrderBook(AmountDenom string, PriceDenom string) BuyOrderBook {
	book := NewOrderBook()
	return BuyOrderBook{
		AmountDenom: AmountDenom,
		PriceDenom:  PriceDenom,
		Book:        &book,
	}
}

func (b *BuyOrderBook) AppendOrder(creator string, amount int32, price int32) (int32, error) {
	return b.Book.appendOrder(creator, amount, price, Decreasing)
}

// オーダーブックで買い注文を約定しようとし、すべての副作用を返します。
func (b *BuyOrderBook) FillSellOrder(order Order) (
	remainingSellOrder Order,
	liquidated []Order,
	gain int32,
	filled bool,
) {
	var liquidatedList []Order
	totalGain := int32(0)
	remainingSellOrder = order

	// 一致している限り清算する
	for {
		var match bool
		var liquidation Order
		remainingSellOrder, liquidation, gain, match, filled = b.LiquidateFromSellOrder(
			remainingSellOrder,
		)
		if !match {
			break
		}

		// 利益を更新する
		totalGain += gain

		// 清算リスト
		liquidatedList = append(liquidatedList, liquidation)

		if filled {
			break
		}
	}

	return remainingSellOrder, liquidatedList, totalGain, filled
}

// 買い注文から最初の売り注文を清算
// 一致するものが見つからない場合、もしくは、一致する場合はfalseを返す
func (b *BuyOrderBook) LiquidateFromSellOrder(order Order) (
	remainingSellOrder Order,
	liquidatedBuyOrder Order,
	gain int32,
	match bool,
	filled bool,
) {
	remainingSellOrder = order

	// 注文がない場合は一致しない
	orderCount := len(b.Book.Orders)
	if orderCount == 0 {
		return order, liquidatedBuyOrder, gain, false, false
	}

	// Check if match
	highestBid := b.Book.Orders[orderCount-1]
	if order.Price > highestBid.Price {
		return order, liquidatedBuyOrder, gain, false, false
	}

	liquidatedBuyOrder = *highestBid

	// 売り注文が完全に約定できるかどうかを確認する
	if highestBid.Amount >= order.Amount {
		remainingSellOrder.Amount = 0
		liquidatedBuyOrder.Amount = order.Amount
		gain = order.Amount * highestBid.Price

		// それが完全に清算された場合、最高入札額を削除します
		highestBid.Amount -= order.Amount
		if highestBid.Amount == 0 {
			b.Book.Orders = b.Book.Orders[:orderCount-1]
		} else {
			b.Book.Orders[orderCount-1] = highestBid
		}

		return remainingSellOrder, liquidatedBuyOrder, gain, true, true
	}

	// 完全に満たされていない
	gain = highestBid.Amount * highestBid.Price
	b.Book.Orders = b.Book.Orders[:orderCount-1]
	remainingSellOrder.Amount -= highestBid.Amount

	return remainingSellOrder, liquidatedBuyOrder, gain, true, false
}
