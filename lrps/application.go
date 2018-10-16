package lrps

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var (
	stateKey            = []byte("stateKey")
	lrpsMemberPrefixKey = []byte("lrpsMemberKey:")
)

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

func loadState(db dbm.DB) State {
	stateBytes := db.Get(stateKey)
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

func prefixKey(key []byte) []byte {
	return append(lrpsMemberPrefixKey, key...)
}

type LRPSApplication struct {
	state State
}

func NewLRPSApplication() *LRPSApplication {
	state := loadState(dbm.NewMemDB())
	return &LRPSApplication{state: state}
}

func (app *LRPSApplication) Info(req types.RequestInfo) types.ResponseInfo {
	return types.ResponseInfo{}
}

func (app *LRPSApplication) SetOption(req types.RequestSetOption) types.ResponseSetOption {
	return types.ResponseSetOption{}
}

func (app *LRPSApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	var key, value []byte
	parts := bytes.Split(tx, []byte("="))
	if len(parts) == 2 {
		key, value = parts[0], parts[1]
	} else {
		key, value = tx, tx
	}
	app.state.db.Set(prefixKey(key), value)
	app.state.Size += 1

	tags := []cmn.KVPair{
		{Key: []byte("app.key"), Value: key},
	}
	return types.ResponseDeliverTx{Code: code.CodeTypeOK, Tags: tags}

}

func (app *LRPSApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
}

func (app *LRPSApplication) Commit() types.ResponseCommit {
	// Using a memdb - just return the big endian size of the db
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height += 1
	saveState(app.state)
	return types.ResponseCommit{Data: appHash}

}

func (app *LRPSApplication) Query(req types.RequestQuery) (res types.ResponseQuery) {
	if req.Prove {
		value := app.state.db.Get(prefixKey(req.Data))
		res.Index = -1 // TODO make Proof return index
		res.Key = req.Data
		res.Value = value
		if value != nil {
			res.Log = "exists"
		} else {
			res.Log = "does not exist"
		}
		return
	} else {
		value := app.state.db.Get(prefixKey(req.Data))
		res.Value = value
		if value != nil {
			res.Log = "exists"
		} else {
			res.Log = "does not exist"
		}
		return
	}
}

func (app *LRPSApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	return types.ResponseInitChain{}
}

func (app *LRPSApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	return types.ResponseBeginBlock{}
}

func (app *LRPSApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{}
}
