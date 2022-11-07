package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/sirupsen/logrus"
)

type Transaction struct {
}

func (e *Transaction) Name() string {
	return "evm_transaction"
}

func (e *Transaction) Model() interface{} {
	return new(evmmodel.Transaction)
}

func (e *Transaction) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewFullNodeRPCV1 failed")
	}
	defer closer()

	tipSetCid, err := tipSet.Key().Cid()
	if err != nil {
		return errors.Wrap(err, "tipSetCid failed")
	}

	hash, err := api.EthHashFromCid(tipSetCid)
	if err != nil {
		return errors.Wrap(err, "rpc EthHashFromCid failed")
	}
	ethBlock, err := node.EthGetBlockByHash(ctx, hash, true)
	if err != nil {
		return errors.Wrap(err, "rpc EthGetBlockByHash failed")
	}

	transactions := ethBlock.Transactions
	if len(transactions) == 0 {
		logrus.Debugf("can not find any transaction")
		return nil
	}

	var evmTransaction []interface{}
	for _, transaction := range transactions {
		tm := transaction.(map[string]interface{})

		et := evmmodel.Transaction{
			Height:               int64(tipSet.Height()),
			Version:              version,
			Hash:                 tm["hash"].(string),
			BlockHash:            tm["blockHash"].(string),
			From:                 tm["from"].(string),
			Value:                utils.ParseHexToBigInt(tm["value"].(string)).String(),
			MaxFeePerGas:         utils.ParseHexToBigInt(tm["maxFeePerGas"].(string)).String(),
			MaxPriorityFeePerGas: utils.ParseHexToBigInt(tm["maxPriorityFeePerGas"].(string)).String(),
		}

		if _, ok := tm["to"]; ok {
			v, ok := tm["to"].(string)
			if ok {
				et.To = v
			}
		}
		if _, ok := tm["gasLimit"]; ok {
			et.GasLimit, err = utils.ParseHexToUint64(tm["gasLimit"].(string))
			if err != nil {
				return errors.Wrap(err, "ParseHexToUint64 failed")
			}
		}

		et.ChainID, err = utils.ParseHexToUint64(tm["chainId"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.Nonce, err = utils.ParseHexToUint64(tm["nonce"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.BlockNumber, err = utils.ParseHexToUint64(tm["blockNumber"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.TransactionIndex, err = utils.ParseHexToUint64(tm["transacionIndex"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.Type, err = utils.ParseHexToUint64(tm["type"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.Gas, err = utils.ParseHexToUint64(tm["gas"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}

		et.V, err = utils.ParseStrToHex(tm["v"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseStrToHex failed")
		}
		et.R, err = utils.ParseStrToHex(tm["r"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseStrToHex failed")
		}
		et.S, err = utils.ParseStrToHex(tm["s"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseStrToHex failed")
		}
		et.Input, err = utils.ParseStrToHex(tm["input"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseStrToHex failed")
		}

		evmTransaction = append(evmTransaction, et)
	}

	if len(evmTransaction) > 0 {
		if err := storage.WriteMany(ctx, evmTransaction...); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d transaction", len(evmTransaction))

	return nil
}