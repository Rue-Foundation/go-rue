// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

import (
	"fmt"
	"math/big"

	"github.com/Rue-Foundation/go-rue/common"
)

var (
	MainnetGenesisHash = common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3") // Mainnet genesis hash to enforce below configs on
)

var (
	// MainnetChainConfig is the chain parameters to run a node on the main network.
	MainnetChainConfig = &ChainConfig{
		ChainId:        big.NewInt(1),
		FrontierBlock: big.NewInt(1),
		HorizonBlock: big.NewInt(876600),
		HopeBlock: big.NewInt(1534050),
		SettlementBlock: big.NewInt(1972350),
		ByzantiumBlock: big.NewInt(2848950),
		DunedinBlock: big.NewInt(3506400),
		BerlinBlock: big.NewInt(4383000),
		PekingBlock: big.NewInt(5259600),
		RennisanceBlock: big.NewInt(6355350),
		EdinburghBlock: big.NewInt(6574500),
		KitchenerBlock: big.NewInt(7012800),
		WaterlooBlock: big.NewInt(7670250),
		KyotoBlock: big.NewInt(8546850),
		InstanbulBlock: big.NewInt(9642600),
		NovaBlock: big.NewInt(10957500),
		SolBlock: big.NewInt(13149000),
		ChenXingBlock: big.NewInt(13368150),
		TaihakuseiBlock: big.NewInt(13806450),
		SaoHaoBlock: big.NewInt(14463900),
		JupiterBlock: big.NewInt(15350500),
		PlutoBlock: big.NewInt(16655400),
		MilkyWayBlock: big.NewInt(21915000),
		AndromedaBlock: big.NewInt(22134150),
		BodesBlock: big.NewInt(22572450),
		HoagsBlock: big.NewInt(23229900),
		MayallsBlock: big.NewInt(24106500),
		ThalesBlock: big.NewInt(26517150),
		PythagorasBlock: big.NewInt(26955450),
		ParmenidesBlock: big.NewInt(27832050),
		ZenoBlock: big.NewInt(29146950),
		SocratesBlock: big.NewInt(30900150),
		PlatoBlock: big.NewInt(33091650),
		CiceroBlock: big.NewInt(35721450),
		AquinasBlock: big.NewInt(38351250),
		DescartesBlock: big.NewInt(41200200),
		HobbesBlock: big.NewInt(43830000),
		SpinozaBlock: big.NewInt(44049150),
		LockeBlock: big.NewInt(45144900),
		NewtonBlock: big.NewInt(46021500),
		LeibnizBlock: big.NewInt(46240650),
		VoltaireBlock: big.NewInt(47117250),
		HumeBlock: big.NewInt(47774700),
		RousseauBlock: big.NewInt(49308750),
		SmithBlock: big.NewInt(50185350),
		KantBlock: big.NewInt(51061950),
		ButerinBlock: big.NewInt(51938550),
		DAOForkBlock:   big.NewInt(0),
		DAOForkSupport: false,
		EIP150Block:    big.NewInt(0),
		EIP150Hash:     common.HexToHash("0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0"),
		EIP155Block:    big.NewInt(0),
		EIP158Block:    big.NewInt(0),

		Ethash: new(EthashConfig),
	}

	// AllEthashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Ethash consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllEthashProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), new(EthashConfig), nil}

	// AllCliqueProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Clique consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCliqueProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, &CliqueConfig{Period: 0, Epoch: 30000}}

	TestChainConfig = &ChainConfig{big.NewInt(1), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), new(EthashConfig), nil}
	TestRules       = TestChainConfig.Rules(new(big.Int))
)

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainId *big.Int `json:"chainId"` // Chain id identifies the current chain and is used for replay protection

	HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)

	DAOForkBlock   *big.Int `json:"daoForkBlock,omitempty"`   // TheDAO hard-fork switch block (nil = no fork)
	DAOForkSupport bool     `json:"daoForkSupport,omitempty"` // Whether the nodes supports or opposes the DAO hard-fork

	// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	EIP150Block *big.Int    `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	EIP150Hash  common.Hash `json:"eip150Hash,omitempty"`  // EIP150 HF hash (needed for header only clients as only gas pricing changed)

	EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block

	ByzantiumBlock *big.Int `json:"byzantiumBlock,omitempty"` // Byzantium switch block (nil = no fork, 0 = already on byzantium)

	// Various consensus engines
	Ethash *EthashConfig `json:"ethash,omitempty"`
	Clique *CliqueConfig `json:"clique,omitempty"`
}

// EthashConfig is the consensus engine configs for proof-of-work based sealing.
type EthashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *EthashConfig) String() string {
	return "ethash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var engine interface{}
	switch {
	case c.Ethash != nil:
		engine = c.Ethash
	case c.Clique != nil:
		engine = c.Clique
	default:
		engine = "unknown"
	}
	return fmt.Sprintf("{ChainID: %v Frontier: %v Horizon: %v Hope: %v Settlement: %v Byzantium: %v Dunedin: %v Berlin: %v Peking: %v Rennisance: %v Edinburgh: %v Kitchener: %v Waterloo: %v Kyoto: %v Instanbul: %v Nova: %v Sol: %v ChenXing: %v Taihakusei: %v SaoHao: %v Jupiter: %v Pluto: %v MilkyWay: %v Andromeda: %v Bodes: %v Hoags: %v Mayalls: %v Thales: %v Pythagoras: %v Parmenides: %v Zeno: %v Socrates: %v Plato: %v Cicero: %v Aquinas: %v Descartes: %v Hobbes: %v Spinoza: %v Locke: %v Newton: %v Voltaire: %v Hume: %v Rousseau: %v Smith: %v Kant: %v Buterin: %v DAO: %v DAOSupport: %v EIP150: %v EIP155: %v EIP158: %v Engine: %v}",
		c.ChainId,
		c.FrontierBlock,	   
		c.HorizonBlock,
		c.HopeBlock,
		c.SettlementBlock,
		c.ByzantiumBlock,
		c.DunedinBlock,
		c.BerlinBlock,
		c.PekingBlock,
		c.RennisanceBlock,
		c.EdinburghBlock,
		c.KitchenerBlock,
		c.WaterlooBlock,
		c.KyotoBlock,
		c.InstanbulBlock,
		c.NovaBlock,
		c.SolBlock,
		c.ChenXingBlock,
		c.TaihakuseiBlock,
		c.SaoHaoBlock,
		c.JupiterBlock,
		c.PlutoBlock,
		c.MilkyWayBlock,
		c.AndromedaBlock,
		c.BodesBlock,
		c.HoagsBlock,
		c.MayallsBlock,
		c.ThalesBlock,
		c.PythagorasBlock,
		c.ParmenidesBlock,
		c.ZenoBlock,
		c.SocratesBlock,
		c.PlatoBlock,
		c.CiceroBlock,
		c.AquinasBlock,
		c.DescartesBlock,
		c.SpinozaBlock,
		c.LockeBlock,
		c.NewtonBlock,
		c.LeibnizBlock,
		c.VoltaireBlock,
		c.HumeBlock,
		c.RousseauBlock,
		c.SmithBlock,
		c.KantBlock,
		c.ButerinBlock,
		c.DAOForkBlock,
		c.DAOForkSupport,
		c.EIP150Block,
		c.EIP155Block,
		c.EIP158Block,
		engine,
	)
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsFrontier(num *big.Int) bool {
	return isForked(c.FrontierBlock, num)
}
func (c *ChainConfig) IsHorizon(num *big.Int) bool {
	return isForked(c.HorizonBlock, num)
}
func (c *ChainConfig) IsHope(num *big.Int) bool {
	return isForked(c.HopeBlock, num)
}
func (c *ChainConfig) IsSettlement(num *big.Int) bool {
	return isForked(c.SettlementBlock, num)
}
func (c *ChainConfig) IsByzantium(num *big.Int) bool {
	return isForked(c.ByzantiumBlock, num)
}
func (c *ChainConfig) IsDunedin(num *big.Int) bool {
	return isForked(c.DunedinBlock, num)
}
func (c *ChainConfig) IsBerlin(num *big.Int) bool {
	return isForked(c.BerlinBlock, num)
}
func (c *ChainConfig) IsPeking(num *big.Int) bool {
	return isForked(c.PekingBlock, num)
}
func (c *ChainConfig) IsRennisance(num *big.Int) bool {
	return isForked(c.RennisanceBlock, num)
}
func (c *ChainConfig) IsEdinburgh(num *big.Int) bool {
	return isForked(c.EdinburghBlock, num)
}
func (c *ChainConfig) IsKitchener(num *big.Int) bool {
	return isForked(c.KitchenerBlock, num)
}
func (c *ChainConfig) IsWaterloo(num *big.Int) bool {
	return isForked(c.WaterlooBlock, num)
}
func (c *ChainConfig) IsKyoto(num *big.Int) bool {
	return isForked(c.KyotoBlock, num)
}
func (c *ChainConfig) IsInstanbul(num *big.Int) bool {
	return isForked(c.InstanbulBlock, num)
}
func (c *ChainConfig) IsNova(num *big.Int) bool {
	return isForked(c.NovaBlock, num)
}
func (c *ChainConfig) IsSol(num *big.Int) bool {
	return isForked(c.SolBlock, num)
}
func (c *ChainConfig) IsChenXing(num *big.Int) bool {
	return isForked(c.ChenXingBlock, num)
}
func (c *ChainConfig) IsTaihakusei(num *big.Int) bool {
	return isForked(c.TaihakuseiBlock, num)
}
func (c *ChainConfig) IsSaoHao(num *big.Int) bool {
	return isForked(c.FrontierBlock, num)
}
func (c *ChainConfig) IsJupiter(num *big.Int) bool {
	return isForked(c.JupiterBlock, num)
}
func (c *ChainConfig) IsPluto(num *big.Int) bool {
	return isForked(c.PlutoBlock, num)
}
func (c *ChainConfig) IsMilkyWay(num *big.Int) bool {
	return isForked(c.MilkyWayBlock, num)
}
func (c *ChainConfig) IsAndromeda(num *big.Int) bool {
	return isForked(c.AndromedaBlock, num)
}
func (c *ChainConfig) IsFrontier(num *big.Int) bool {
	return isForked(c.FrontierBlock, num)
}
func (c *ChainConfig) IsHoags(num *big.Int) bool {
	return isForked(c.HoagsBlock, num)
}
func (c *ChainConfig) IsMayalls(num *big.Int) bool {
	return isForked(c.MayallsBlock, num)
}
func (c *ChainConfig) IsThales(num *big.Int) bool {
	return isForked(c.ThalesBlock, num)
}
func (c *ChainConfig) IsPythagoras(num *big.Int) bool {
	return isForked(c.PythagorasBlock, num)
}
func (c *ChainConfig) IsParmenides(num *big.Int) bool {
	return isForked(c.ParmenidesBlock, num)
}
func (c *ChainConfig) IsZeno(num *big.Int) bool {
	return isForked(c.ZenoBlock, num)
}
func (c *ChainConfig) IsSocrates(num *big.Int) bool {
	return isForked(c.SocratesBlock, num)
}
func (c *ChainConfig) IsPlato(num *big.Int) bool {
	return isForked(c.PlatoBlock, num)
}
func (c *ChainConfig) IsCicero(num *big.Int) bool {
	return isForked(c.CiceroBlock, num)
}
func (c *ChainConfig) IsAquinas(num *big.Int) bool {
	return isForked(c.AquinasBlock, num)
}
func (c *ChainConfig) IsDescartes(num *big.Int) bool {
	return isForked(c.DescartesBlock, num)
}
func (c *ChainConfig) IsHobbes(num *big.Int) bool {
	return isForked(c.HobbesBlock, num)
}
func (c *ChainConfig) IsSpinoza(num *big.Int) bool {
	return isForked(c.SpinozaBlock, num)
}
func (c *ChainConfig) IsLocke(num *big.Int) bool {
	return isForked(c.LockeBlock, num)
}
func (c *ChainConfig) IsNewton(num *big.Int) bool {
	return isForked(c.NewtonBlock, num)
}
func (c *ChainConfig) IsLeibniz(num *big.Int) bool {
	return isForked(c.LeibnizBlock, num)
}
func (c *ChainConfig) IsVoltaire(num *big.Int) bool {
	return isForked(c.VoltaireBlock, num)
}
func (c *ChainConfig) IsHume(num *big.Int) bool {
	return isForked(c.HumeBlock, num)
}
func (c *ChainConfig) IsRousseau(num *big.Int) bool {
	return isForked(c.RousseauBlock, num)
}
func (c *ChainConfig) IsSmith(num *big.Int) bool {
	return isForked(c.SmithBlock, num)
}
func (c *ChainConfig) IsKant(num *big.Int) bool {
	return isForked(c.KantBlock, num)
}
func (c *ChainConfig) IsButerin(num *big.Int) bool {
	return isForked(c.ButerinBlock, num)
}

// IsDAO returns whether num is either equal to the DAO fork block or greater.
func (c *ChainConfig) IsDAOFork(num *big.Int) bool {
	return isForked(c.DAOForkBlock, num)
}

func (c *ChainConfig) IsEIP150(num *big.Int) bool {
	return isForked(c.EIP150Block, num)
}

func (c *ChainConfig) IsEIP155(num *big.Int) bool {
	return isForked(c.EIP155Block, num)
}

func (c *ChainConfig) IsEIP158(num *big.Int) bool {
	return isForked(c.EIP158Block, num)
}


// GasTable returns the gas table corresponding to the current phase (homestead or homestead reprice).
//
// The returned GasTable's fields shouldn't, under any circumstances, be changed.
func (c *ChainConfig) GasTable(num *big.Int) GasTable {
	if num == nil {
		return GasTableHomestead
	}
	switch {
	case c.IsEIP158(num):
		return GasTableEIP158
	case c.IsEIP150(num):
		return GasTableEIP150
	default:
		return GasTableHomestead
	}
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64) *ConfigCompatError {
	bhead := new(big.Int).SetUint64(height)

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead.SetUint64(err.RewindTo)
	}
	return lasterr
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, head *big.Int) *ConfigCompatError {
	if isForkIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, head) {
		return newCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	}
	if isForkIncompatible(c.DAOForkBlock, newcfg.DAOForkBlock, head) {
		return newCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if c.IsDAOFork(head) && c.DAOForkSupport != newcfg.DAOForkSupport {
		return newCompatError("DAO fork support flag", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if isForkIncompatible(c.EIP150Block, newcfg.EIP150Block, head) {
		return newCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	}
	if isForkIncompatible(c.EIP155Block, newcfg.EIP155Block, head) {
		return newCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	}
	if isForkIncompatible(c.EIP158Block, newcfg.EIP158Block, head) {
		return newCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	}
	if c.IsEIP158(head) && !configNumEqual(c.ChainId, newcfg.ChainId) {
		return newCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	}
	if isForkIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, head) {
		return newCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	}
	return nil
}

// isForkIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkIncompatible(s1, s2, head *big.Int) bool {
	return (isForked(s1, head) || isForked(s2, head)) && !configNumEqual(s1, s2)
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configNumEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntatic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainId                                   *big.Int
	IsHomestead, IsEIP150, IsEIP155, IsEIP158 bool
	IsByzantium                               bool
}

func (c *ChainConfig) Rules(num *big.Int) Rules {
	chainId := c.ChainId
	if chainId == nil {
		chainId = new(big.Int)
	}
	return Rules{ChainId: new(big.Int).Set(chainId), IsHomestead: c.IsHomestead(num), IsEIP150: c.IsEIP150(num), IsEIP155: c.IsEIP155(num), IsEIP158: c.IsEIP158(num), IsByzantium: c.IsByzantium(num)}
}
