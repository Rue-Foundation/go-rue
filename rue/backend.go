// Copyright 2014 The go-rue Authors
// This file is part of the go-rue library.
//
// The go-rue library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-rue library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-rue library. If not, see <http://www.gnu.org/licenses/>.

// Package rue implements the Rue protocol.
package rue

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/Rue-Foundation/go-rue/accounts"
	"github.com/Rue-Foundation/go-rue/common"
	"github.com/Rue-Foundation/go-rue/common/hexutil"
	"github.com/Rue-Foundation/go-rue/consensus"
	"github.com/Rue-Foundation/go-rue/consensus/clique"
	"github.com/Rue-Foundation/go-rue/consensus/ruehash"
	"github.com/Rue-Foundation/go-rue/core"
	"github.com/Rue-Foundation/go-rue/core/bloombits"
	"github.com/Rue-Foundation/go-rue/core/types"
	"github.com/Rue-Foundation/go-rue/core/vm"
	"github.com/Rue-Foundation/go-rue/rue/downloader"
	"github.com/Rue-Foundation/go-rue/rue/filters"
	"github.com/Rue-Foundation/go-rue/rue/gasprice"
	"github.com/Rue-Foundation/go-rue/ruedb"
	"github.com/Rue-Foundation/go-rue/event"
	"github.com/Rue-Foundation/go-rue/internal/rueapi"
	"github.com/Rue-Foundation/go-rue/log"
	"github.com/Rue-Foundation/go-rue/miner"
	"github.com/Rue-Foundation/go-rue/node"
	"github.com/Rue-Foundation/go-rue/p2p"
	"github.com/Rue-Foundation/go-rue/params"
	"github.com/Rue-Foundation/go-rue/rlp"
	"github.com/Rue-Foundation/go-rue/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Rue implements the Rue full node service.
type Rue struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the rue
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb ruedb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *RueApiBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	ruebase common.Address

	networkId     uint64
	netRPCService *rueapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and ruebase)
}

func (s *Rue) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Rue object (including the
// initialisation of the common Rue object)
func New(ctx *node.ServiceContext, config *Config) (*Rue, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run rue.Rue in light sync mode, use les.LightRue")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	rue := &Rue{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.Ruehash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		ruebase:      config.Ruebase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising Rue protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run grue upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}

	vmConfig := vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
	rue.blockchain, err = core.NewBlockChain(chainDb, rue.chainConfig, rue.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		rue.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	rue.bloomIndexer.Start(rue.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	rue.txPool = core.NewTxPool(config.TxPool, rue.chainConfig, rue.blockchain)

	if rue.protocolManager, err = NewProtocolManager(rue.chainConfig, config.SyncMode, config.NetworkId, rue.eventMux, rue.txPool, rue.engine, rue.blockchain, chainDb); err != nil {
		return nil, err
	}
	rue.miner = miner.New(rue, rue.chainConfig, rue.EventMux(), rue.engine)
	rue.miner.SetExtra(makeExtraData(config.ExtraData))

	rue.ApiBackend = &RueApiBackend{rue, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	rue.ApiBackend.gpo = gasprice.NewOracle(rue.ApiBackend, gpoParams)

	return rue, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"grue",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (ruedb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*ruedb.LDBDatabase); ok {
		db.Meter("rue/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Rue service
func CreateConsensusEngine(ctx *node.ServiceContext, config *ruehash.Config, chainConfig *params.ChainConfig, db ruedb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowMode == ruehash.ModeFake:
		log.Warn("Ruehash used in fake mode")
		return ruehash.NewFaker()
	case config.PowMode == ruehash.ModeTest:
		log.Warn("Ruehash used in test mode")
		return ruehash.NewTester()
	case config.PowMode == ruehash.ModeShared:
		log.Warn("Ruehash used in shared mode")
		return ruehash.NewShared()
	default:
		engine := ruehash.New(ruehash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the rue package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Rue) APIs() []rpc.API {
	apis := rueapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "rue",
			Version:   "1.0",
			Service:   NewPublicRueAPI(s),
			Public:    true,
		}, {
			Namespace: "rue",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "rue",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "rue",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Rue) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Rue) Ruebase() (eb common.Address, err error) {
	s.lock.RLock()
	ruebase := s.ruebase
	s.lock.RUnlock()

	if ruebase != (common.Address{}) {
		return ruebase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			ruebase := accounts[0].Address

			s.lock.Lock()
			s.ruebase = ruebase
			s.lock.Unlock()

			log.Info("Ruebase automatically configured", "address", ruebase)
			return ruebase, nil
		}
	}
	return common.Address{}, fmt.Errorf("ruebase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *Rue) SetRuebase(ruebase common.Address) {
	self.lock.Lock()
	self.ruebase = ruebase
	self.lock.Unlock()

	self.miner.SetRuebase(ruebase)
}

func (s *Rue) StartMining(local bool) error {
	eb, err := s.Ruebase()
	if err != nil {
		log.Error("Cannot start mining without ruebase", "err", err)
		return fmt.Errorf("ruebase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Ruebase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Rue) StopMining()         { s.miner.Stop() }
func (s *Rue) IsMining() bool      { return s.miner.Mining() }
func (s *Rue) Miner() *miner.Miner { return s.miner }

func (s *Rue) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Rue) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Rue) TxPool() *core.TxPool               { return s.txPool }
func (s *Rue) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Rue) Engine() consensus.Engine           { return s.engine }
func (s *Rue) ChainDb() ruedb.Database            { return s.chainDb }
func (s *Rue) IsListening() bool                  { return true } // Always listening
func (s *Rue) RueVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Rue) NetVersion() uint64                 { return s.networkId }
func (s *Rue) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Rue) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Rue protocol implementation.
func (s *Rue) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = rueapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		maxPeers -= s.config.LightPeers
		if maxPeers < srvr.MaxPeers/2 {
			maxPeers = srvr.MaxPeers / 2
		}
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Rue protocol.
func (s *Rue) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
