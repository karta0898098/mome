# mome
mome matches incoming buy/sell orders to create trade.


## Currently supported
- order types
  - market order — execute an order as fast as possible, cross the spread.
  - limit order - execute an order with a limit on bid/ask prices.
- order method
  - FOK - filled or killed, immediately match an order in full (without partial fills) or cancel it.

## How order matching?

Market orders are always given priority above all other orders, then sorted according to time of arrival.

- orders are FIFO
  - bids - price (descending), time (ascending)
  - asks - price (ascending), time (ascending)
  - market price is set at the last trade price