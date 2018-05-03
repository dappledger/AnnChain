package node

import (
	"bytes"

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	cvtools "github.com/dappledger/AnnChain/src/tools"
)

// --- since restoreOrg* just replay tx already in the blocks, so we don't care any error

func (met *Metropolis) restoreOrgCancel(txBytes []byte) {
	cancelTx := &OrgCancelTx{}
	if err := cvtools.TxFromBytes(txBytes, cancelTx); err != nil {
		return
	}
	if ok, _ := cvtools.TxVerifySignature(cancelTx); !ok {
		return
	}
	if err := met.checkOrgCancel(cancelTx); err != nil {
		return
	}
	delete(met.PendingOrgTxs, string(cancelTx.TxHash))
}

func (met *Metropolis) restoreOrgConfirm(txBytes []byte) {
	confirmTx := &OrgConfirmTx{}
	if err := cvtools.TxFromBytes(txBytes, confirmTx); err != nil {
		return
	}
	if ok, _ := cvtools.TxVerifySignature(confirmTx); !ok {
		return
	}
	if err := met.checkOrgConfirm(confirmTx); err != nil {
		return
	}
	delete(met.PendingOrgTxs, string(confirmTx.TxHash))
}

// restoreOrg won't check Metropolis.OrgState, apart from that, everything works the same as execution orgtx
func (met *Metropolis) restoreOrg(txBytes []byte) {
	var (
		err    error
		pubkey crypto.PubKeyEd25519
		txHash []byte
		node   *OrgNode

		orgtx = &OrgTx{}
	)
	pubkey, _ = met.core.GetPublicKey()

	if err := cvtools.TxFromBytes(txBytes, orgtx); err != nil {
		met.logger.Error("restore failed deserialize", zap.Error(err))
		return
	}
	if ok, _ := cvtools.TxVerifySignature(orgtx); !ok {
		return
	}
	if err := met.checkOrgs(orgtx); err != nil {
		met.logger.Error("restore failed on checkOrg", zap.Error(err))
		return
	}

	if txHash, err = cvtools.TxHash(orgtx); err != nil {
		met.logger.Error("restore failed", zap.Error(err))
		return
	}

	met.PendingOrgTxs[string(txHash)] = orgtx

	if !bytes.Equal(orgtx.PubKey, pubkey[:]) {
		return
	}

	// here we don't care about orgstate, because it is exactly the shape we want
	switch orgtx.Act {
	case OrgCreate:
		if node, err = met.createOrgNode(orgtx); err != nil {
			met.logger.Error("restore create failed", zap.Error(err))
			return
		}
		met.SetOrg(orgtx.ChainID, orgtx.App, node)
	case OrgJoin:
		if node, err = met.createOrgNode(orgtx); err != nil {
			met.logger.Error("restore join failed", zap.Error(err))
			return
		}
		met.SetOrg(orgtx.ChainID, orgtx.App, node)
	case OrgLeave:
		if err = met.RemoveOrg(orgtx.ChainID); err != nil {
			return
		}
	}
}
