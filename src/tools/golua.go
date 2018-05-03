package tools

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	ljson "layeh.com/gopher-json"
	luar "layeh.com/gopher-luar"

	cvtypes "github.com/dappledger/AnnChain/src/types"
)

var (
	ErrRetTypeWrong = errors.New("type of returned value is neither LTable nor LTString")
)

func NewLuaState() *lua.LState {
	L := lua.NewState(lua.Options{
		CallStackSize: 120,
		RegistrySize:  120 * 20,
	})
	ljson.Preload(L) // Load module json
	return L
}

// ExecLuaWithParam will not close lua.LState, so you should close it manually afterwards.
func ExecLuaWithParam(l *lua.LState, code string, entParams cvtypes.ParamUData) (cvtypes.ParamUData, error) {
	defer l.Close()

	if err := l.DoString(code); err != nil {
		return nil, errors.Wrap(err, "[lua: dostring]")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	l.SetContext(ctx)

	err := l.CallByParam(lua.P{
		Fn:      l.GetGlobal("main"),
		NRet:    1,
		Protect: true,
	}, luar.New(l, entParams))
	if err != nil {
		return nil, errors.Wrap(err, "[lua: main]")
	}

	ret := l.Get(-1)

	// return nil in lua will cancel the event processing
	if lua.LVIsFalse(ret) {
		return nil, nil
	} else if lua.LVCanConvToString(ret) {
		return nil, errors.Wrap(errors.New(ret.String()), "[lua: main]")
	} else if tb, ok := ret.(*lua.LTable); ok {
		udata := new(cvtypes.ParamUData)
		mapper := gluamapper.NewMapper(gluamapper.Option{
			NameFunc: func(in string) string { return in },
		})
		err := mapper.Map(tb, udata)
		return *udata, err
	} else if ud, ok := ret.(*lua.LUserData); ok {
		if p, ok := ud.Value.(cvtypes.ParamUData); ok {
			return p, nil
		}

		return nil, ErrRetTypeWrong
	}

	return nil, nil
}

func LuaSyntaxCheck(code string) error {
	L := NewLuaState()
	_, err := L.LoadString(code)
	L.Close()
	return err
}

func CloseLuaState(l *lua.LState) {
	l.Close()
}
