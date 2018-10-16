# Limited Rock-paper-scissors

Study of [tendermint](https://www.tendermint.com/).

## Development & Test

```sh
# Reset node
rm -rf ~/.tendermint/data && tendermint unsafe_reset_priv_validator
# Run
go run cmd/limited-rps/main.go
# Invite members to boat
curl -s 'localhost:26657/broadcast_tx_commit?tx="init:foo,bar"'
# Play game
curl -s 'localhost:26657/broadcast_tx_commit?tx="play:1:foo=rock,bar=paper"'
# Check status
curl -s 'localhost:26657/abci_query?data="bar"' | jq -r .result.response.value | base64 -d
# Play more games
curl -s 'localhost:26657/broadcast_tx_commit?tx="play:2:foo=rock,bar=paper"'
curl -s 'localhost:26657/broadcast_tx_commit?tx="play:3:foo=rock,bar=paper"'
# He has gone to the underground labor
curl -s 'localhost:26657/abci_query?data="foo"'
```
