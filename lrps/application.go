package lrps

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

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

func memberPrefixKey(key []byte) []byte {
	return append(lrpsMemberPrefixKey, key...)
}

type MemberState struct {
	Cards map[string]int64 `json:"hands"`
	Stars int64            `json:"stars"`
}

func NewMemberStateFromBytes(b []byte) *MemberState {
	var state MemberState
	if len(b) != 0 {
		err := json.Unmarshal(b, &state)
		if err != nil {
			panic(err)
		}
	}
	return &state
}

func (s *MemberState) Leave() bool {
	if s.Stars == 0 {
		return true
	} else if s.Stars <= 2 {
		for _, n := range s.Cards {
			if n != 0 {
				return false
			}
		}
		return true
	}
	return false
}

func (s *MemberState) Bytes() []byte {
	res, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return res
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

var (
	initTxPrefix = []byte("init:")
	playTxPrefix = []byte("play:")
)

func (app *LRPSApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	if bytes.HasPrefix(tx, initTxPrefix) {
		members := bytes.Split(tx[len(initTxPrefix):], []byte(","))

		for _, m := range members {
			app.state.db.Set(memberPrefixKey(m), (&MemberState{
				Cards: map[string]int64{
					"rock":    4,
					"paper":   4,
					"scissor": 4,
				},
				Stars: 3,
			}).Bytes())
		}
		app.state.Size = int64(len(members))
	} else if bytes.HasPrefix(tx, playTxPrefix) {
		prefixIDTx := bytes.SplitN(tx[len(playTxPrefix):], []byte(":"), 2)
		id := prefixIDTx[0]

		nameHandPairBytes := bytes.SplitN(prefixIDTx[1], []byte(","), 2)

		names := make([][]byte, 2)
		hands := make([]Hand, 2)
		states := make([]*MemberState, 2)
		for i, nameHandBytes := range nameHandPairBytes {
			name, hand, err := parseNameHand(nameHandBytes)
			if err != nil {
				return types.ResponseDeliverTx{Code: CodeFormatError}
			}

			stateBytes := app.state.db.Get(memberPrefixKey(name))
			if stateBytes == nil {
				return types.ResponseDeliverTx{Code: CodeMemberNotFoundError}
			}
			state := NewMemberStateFromBytes(stateBytes)

			if state.Cards[hand.String()] == 0 {
				return types.ResponseDeliverTx{Code: CodeNoCardError}
			}

			names[i] = name
			hands[i] = hand
			states[i] = state
		}

		for i, s := range states {
			s.Cards[hands[i].String()]--
		}
		result := CompareHands(hands[0], hands[1])
		var winner []byte
		if result > 0 {
			states[0].Stars++
			states[1].Stars--
			winner = names[0]
		} else if result < 0 {
			states[0].Stars--
			states[1].Stars++
			winner = names[1]
		}
		for i, s := range states {
			if s.Leave() {
				app.state.db.Delete(memberPrefixKey(names[i]))
			} else {
				app.state.db.Set(memberPrefixKey(names[i]), states[i].Bytes())
			}
		}
		tags := []cmn.KVPair{
			{Key: []byte("app.id"), Value: id},
			{Key: []byte("app.winner"), Value: winner},
		}
		return types.ResponseDeliverTx{Code: CodeTypeOK, Tags: tags}
	}
	return types.ResponseDeliverTx{Code: CodeFormatError}
}

func parseNameHand(nameHandBytes []byte) ([]byte, Hand, error) {
	nameHand := bytes.SplitN(nameHandBytes, []byte("="), 2)
	name, handBytes := nameHand[0], nameHand[1]

	hand, err := ParseHand(string(handBytes))
	return name, hand, err
}

func (app *LRPSApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	return types.ResponseCheckTx{Code: CodeTypeOK, GasWanted: 1}
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
		value := app.state.db.Get(memberPrefixKey(req.Data))
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
		value := app.state.db.Get(memberPrefixKey(req.Data))
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
