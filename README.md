# What is this?

A basic Golang-based API that proxies GraphQL queries to a Uniswap v3 subgraph hosted on TheGraph [here](https://api.thegraph.com/subgraphs/name/ianlapham/uniswap-v3-alt).

# How does it work?

1. Start the app: `go run main.go` or `docker-compose up`
1. Throw requests at the endpoints:
    * `/asset/:id`: Where `:id` is an ETH contract address of some sort. (Ex: `/asset/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48`)
        * Returns:
            * `Name`: Token Name 
            * `Symbol`: Token Symbol
            * `ContractAddress`: Token Contract Address
            * `VolumeUSD`: Total USD Volume in all pools
            * `Pools`: List of all participating liqiuidity pools
    * `/block/:blocknumber`: Where `:blocknumber` is a block height integer. (Ex: `/block/13315378`)
        * Returns:
            * `BlockHeight`: Requested block height.
            * `Swaps`: List of swaps performed in the requested block.
            * `AssetsSwapped`: List of assets involved in all swaps in the requested block. 