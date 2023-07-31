package subspace_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	abci "github.com/fibonacci-chain/fbc/libs/tendermint/abci/types"
	"github.com/fibonacci-chain/fbc/libs/tendermint/libs/log"
	dbm "github.com/fibonacci-chain/fbc/libs/tm-db"
	"github.com/stretchr/testify/suite"

	"github.com/fibonacci-chain/fbc/libs/cosmos-sdk/codec"
	"github.com/fibonacci-chain/fbc/libs/cosmos-sdk/store"
	sdk "github.com/fibonacci-chain/fbc/libs/cosmos-sdk/types"
	"github.com/fibonacci-chain/fbc/libs/cosmos-sdk/x/params/subspace"
)

type SubspaceTestSuite struct {
	suite.Suite

	cdc *codec.Codec
	ctx sdk.Context
	ss  subspace.Subspace
}

func (suite *SubspaceTestSuite) SetupTest() {
	cdc := codec.New()
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkey, sdk.StoreTypeTransient, db)
	suite.NoError(ms.LoadLatestVersion())

	ss := subspace.NewSubspace(cdc, key, tkey, "testsubspace")

	suite.cdc = cdc
	suite.ctx = sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	suite.ss = ss.WithKeyTable(paramKeyTable())
}

func (suite *SubspaceTestSuite) TestKeyTable() {
	suite.Require().True(suite.ss.HasKeyTable())
	suite.Require().Panics(func() {
		suite.ss.WithKeyTable(paramKeyTable())
	})
	suite.Require().NotPanics(func() {
		ss := subspace.NewSubspace(codec.New(), key, tkey, "testsubspace2")
		ss = ss.WithKeyTable(paramKeyTable())
	})
}

func (suite *SubspaceTestSuite) TestGetSet() {
	var v time.Duration
	t := time.Hour * 48

	suite.Require().Panics(func() {
		suite.ss.Get(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Require().NotEqual(t, v)
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.Require().NotPanics(func() {
		suite.ss.Get(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Require().Equal(t, v)
}

func (suite *SubspaceTestSuite) TestGetIfExists() {
	var v time.Duration

	suite.Require().NotPanics(func() {
		suite.ss.GetIfExists(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Require().Equal(time.Duration(0), v)
}

func (suite *SubspaceTestSuite) TestGetRaw() {
	t := time.Hour * 48

	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.Require().NotPanics(func() {
		res := suite.ss.GetRaw(suite.ctx, keyUnbondingTime)
		suite.Require().Equal("2231373238303030303030303030303022", fmt.Sprintf("%X", res))
	})
}

func (suite *SubspaceTestSuite) TestHas() {
	t := time.Hour * 48

	suite.Require().False(suite.ss.Has(suite.ctx, keyUnbondingTime))
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.Require().True(suite.ss.Has(suite.ctx, keyUnbondingTime))
}

func (suite *SubspaceTestSuite) TestModified() {
	t := time.Hour * 48

	suite.Require().False(suite.ss.Modified(suite.ctx, keyUnbondingTime))
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.Require().True(suite.ss.Modified(suite.ctx, keyUnbondingTime))
}

func (suite *SubspaceTestSuite) TestUpdate() {
	suite.Require().Panics(func() {
		suite.ss.Update(suite.ctx, []byte("invalid_key"), nil)
	})

	t := time.Hour * 48
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})

	bad := time.Minute * 5

	bz, err := suite.cdc.MarshalJSON(bad)
	suite.Require().NoError(err)
	suite.Require().Error(suite.ss.Update(suite.ctx, keyUnbondingTime, bz))

	good := time.Hour * 360
	bz, err = suite.cdc.MarshalJSON(good)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.ss.Update(suite.ctx, keyUnbondingTime, bz))

	var v time.Duration

	suite.Require().NotPanics(func() {
		suite.ss.Get(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Require().Equal(good, v)
}

func (suite *SubspaceTestSuite) TestGetParamSet() {
	a := params{
		UnbondingTime: time.Hour * 48,
		MaxValidators: 100,
		BondDenom:     "stake",
	}
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, a.UnbondingTime)
		suite.ss.Set(suite.ctx, keyMaxValidators, a.MaxValidators)
		suite.ss.Set(suite.ctx, keyBondDenom, a.BondDenom)
	})

	b := params{}
	suite.Require().NotPanics(func() {
		suite.ss.GetParamSet(suite.ctx, &b)
	})
	suite.Require().Equal(a.UnbondingTime, b.UnbondingTime)
	suite.Require().Equal(a.MaxValidators, b.MaxValidators)
	suite.Require().Equal(a.BondDenom, b.BondDenom)
}

func (suite *SubspaceTestSuite) TestSetParamSet() {
	testCases := []struct {
		name string
		ps   subspace.ParamSet
	}{
		{"invalid unbonding time", &params{time.Hour * 1, 100, "stake"}},
		{"invalid bond denom", &params{time.Hour * 48, 100, ""}},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.Require().Panics(func() {
				suite.ss.SetParamSet(suite.ctx, tc.ps)
			})
		})
	}

	a := params{
		UnbondingTime: time.Hour * 48,
		MaxValidators: 100,
		BondDenom:     "stake",
	}
	suite.Require().NotPanics(func() {
		suite.ss.SetParamSet(suite.ctx, &a)
	})

	b := params{}
	suite.Require().NotPanics(func() {
		suite.ss.GetParamSet(suite.ctx, &b)
	})
	suite.Require().Equal(a.UnbondingTime, b.UnbondingTime)
	suite.Require().Equal(a.MaxValidators, b.MaxValidators)
	suite.Require().Equal(a.BondDenom, b.BondDenom)
}

func (suite *SubspaceTestSuite) TestSetParamSetForInitGenesis() {
	testCases := []struct {
		name       string
		ps         subspace.ParamSet
		ignoreList [][]byte
	}{
		{"ignore all", &params{time.Hour * 48, 100, "stake"}, [][]byte{keyUnbondingTime, keyMaxValidators, keyBondDenom}},
		{"ignore two", &params{time.Hour * 48, 100, "stake"}, [][]byte{keyUnbondingTime, keyMaxValidators}},
		{"ignore one", &params{time.Hour * 48, 100, "stake"}, [][]byte{keyUnbondingTime}},
		{"ignore nil", &params{time.Hour * 48, 100, "stake"}, [][]byte{}},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.ss.SetParamSetForInitGenesis(suite.ctx, tc.ps, tc.ignoreList)
			for _, pair := range tc.ps.ParamSetPairs() {
				beIgnore := false
				for _, ignore := range tc.ignoreList {
					if bytes.Equal(ignore, pair.Key) {
						beIgnore = true
						break
					}
				}

				if beIgnore {
					suite.Require().False(suite.ss.Has(suite.ctx, pair.Key))
				} else {
					suite.Require().True(suite.ss.Has(suite.ctx, pair.Key))
				}
				suite.Require().NotPanics(func() {
					suite.ss.GetParamSetForInitGenesis(suite.ctx, tc.ps, tc.ignoreList)
				})
			}
		})
	}
}

func (suite *SubspaceTestSuite) TestName() {
	suite.Require().Equal("testsubspace", suite.ss.Name())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SubspaceTestSuite))
}