# mome

mome matches incoming buy/sell orders to create trade.

## Currently supported

- order types
    - market order — execute an order as fast as possible, cross the spread.
    - limit order - execute an order with a limit on bid/ask prices.
- order params
    - STOP - stop order, set a stop price which will activate the order once the market price crosses it
    - FOK - filled or killed, immediately match an order in full (without partial fills) or cancel it.
    - AON - all or nothing, don't allow partial fills
    - IOC - immediate or cancel, immediately fill what's possible, cancel the rest

## How order matching?

Market orders are always given priority above all other orders, then sorted according to time of arrival.

- orders are FIFO
    - bids - price (descending), time (ascending)
    - asks - price (ascending), time (ascending)
    - market price is set at the last trade price

## TODO

* [ ] raft recovery implement
* [ ] trade info sends to MessageQueue

## Architecture

```mermaid
---
title: Architecture
---
flowchart TB
    gateway[Gateway] --> traderServers
    subgraph traderServers[trader servers]
        direction TB
        ts1(trader server)
        ts2(trader server)
    end

    subgraph matchingEngines[matching engines]
        me1(me leader)
        me2(me follower)
        me3(me follower)
    end

    subgraph matchingEngines2[matching engines]
        me4(me leader)
        me5(me follower)
        me6(me follower)
    end

    subgraph matchingEngines3[matching engines]
        me7(me leader)
        me8(me follower)
        me9(me follower)
    end

    traderServers -- grpc unary --> matchingEngines
    traderServers -- grpc unary --> matchingEngines2
    traderServers -- grpc unary --> matchingEngines3

```

### Gateway

The gateway receives all frontend requests and aggregates the necessary business logic.

### Trader server

The trader server determines the routing for the 'me-cluster,' manages its state, conducts asset verification, and
calculates transaction fees.

### Matching engine cluster

Splitting the engine into different groups based on stock codes, each 'me cluster' handles a specific trading volume.
Designing the 'me cluster' to use Raft for maintaining data consistency within the cluster.

