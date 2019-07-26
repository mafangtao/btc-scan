package main

import (
	"flag"
	"github.com/judwhite/go-svc/svc"
	"github.com/liyue201/btc-scan/btc"
	"github.com/liyue201/btc-scan/config"
	"github.com/liyue201/btc-scan/logic"
	"github.com/liyue201/btc-scan/rest"
	st "github.com/liyue201/btc-scan/storage"
	"github.com/liyue201/btc-scan/storage/leveldb"
	"github.com/liyue201/btc-scan/uitils"
	"github.com/liyue201/go-logger"
	"syscall"
)

type program struct {
	utils.WaitGroupWrapper
	btcClient  *btc.Client
	worker     *btc.Worker
	httpServer *rest.RestSever
	storage    st.Storage
}

func (p *program) Init(env svc.Environment) error {
	btcCli, err := btc.NewClient(&config.Cfg.Btcnode)
	if err != nil {
		logger.Errorf("Init btc.NewClient %s", err)
		return err
	}
	p.btcClient = btcCli

	storage, err := leveldb.NewLevelDbStorage(config.Cfg.Leveldb.Path)
	if err != nil {
		logger.Errorf("Init leveldb.NewLevelDbStorage %s", err)
		return err
	}
	p.storage = storage
	txmgr, err := logic.NewTxMgr(storage, config.Cfg.Btcnode.Mempool)
	if err != nil {
		logger.Errorf("Initlogic.NewTxMgr  %s", err)
		return err
	}
	p.worker = btc.NewWorker(btcCli, txmgr)

	p.httpServer = rest.NewHttpServer(config.Cfg.Rest.Port, txmgr, btcCli)

	logger.Info("program inited")
	return nil
}

func (p *program) Start() error {
	p.Wrap(func() {
		p.btcClient.Run()
	})
	p.Wrap(func() {
		p.worker.Run()
	})
	p.Wrap(func() {
		p.httpServer.Run()
	})
	logger.Info("program start")
	return nil
}

func (p *program) Stop() error {
	p.httpServer.Stop()
	p.worker.Stop()
	p.btcClient.Stop()
	p.Wait()
	p.storage.Close()
	logger.Info("program stopped")
	return nil
}

func main() {
	cfg := flag.String("C", "configs/scan.yml", "configuration file")
	flag.Parse()
	err := config.Init(*cfg)
	if err != nil {
		logger.Error("init config error:", err)
		return
	}
	service := &program{}
	if err := svc.Run(service, syscall.SIGINT, syscall.SIGTERM); err != nil {
		logger.Errorf("main: %v", err)
	}
}
