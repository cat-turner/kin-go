package testutil

import (
	"bytes"
	"crypto/ed25519"
	"sort"
	"testing"

	"github.com/kinecosystem/agora-common/kin"
	"github.com/kinecosystem/go/keypair"
	hProtocol "github.com/kinecosystem/go/protocols/horizon"
	"github.com/kinecosystem/go/protocols/horizon/base"
	"github.com/kinecosystem/go/strkey"
	"github.com/pkg/errors"
	"github.com/stellar/go/xdr"
	"github.com/stretchr/testify/require"
)

func StellarAccountIDFromString(address string) (id xdr.AccountId, err error) {
	k, err := strkey.Decode(strkey.VersionByteAccountID, address)
	if err != nil {
		return id, errors.New("failed to decode provided address")
	}
	var v xdr.Uint256
	copy(v[:], k)
	return xdr.AccountId{
		Type:    xdr.PublicKeyTypePublicKeyTypeEd25519,
		Ed25519: &v,
	}, nil
}

func GenerateHorizonAccount(accountID string, nativeBalance string, sequence string) *hProtocol.Account {
	return &hProtocol.Account{
		HistoryAccount: hProtocol.HistoryAccount{
			ID: accountID,
		},
		Balances: []hProtocol.Balance{{
			Balance: nativeBalance,
			Asset:   base.Asset{Type: "native"},
		}},
		Sequence: sequence,
	}
}

func GenerateKin2HorizonAccount(accountID string, balance string, sequence string) *hProtocol.Account {
	return &hProtocol.Account{
		HistoryAccount: hProtocol.HistoryAccount{
			ID: accountID,
		},
		Balances: []hProtocol.Balance{{
			Balance: balance,
			Asset: base.Asset{
				Type:   "credit_alphanum4",
				Code:   kin.KinAssetCode,
				Issuer: kin.Kin2TestIssuer,
			},
		}},
		Sequence: sequence,
	}
}

func GenerateAccountID(t *testing.T) (*keypair.Full, xdr.AccountId) {
	kp, err := keypair.Random()
	require.NoError(t, err)

	pubKey, err := strkey.Decode(strkey.VersionByteAccountID, kp.Address())
	require.NoError(t, err)
	var senderPubKey xdr.Uint256
	copy(senderPubKey[:], pubKey)

	return kp, xdr.AccountId{
		Type:    xdr.PublicKeyTypePublicKeyTypeEd25519,
		Ed25519: &senderPubKey,
	}
}

func GenerateAccountIDs(t *testing.T, n int) []xdr.AccountId {
	accounts := make([]xdr.AccountId, n)
	for i := 0; i < n; i++ {
		_, accountID := GenerateAccountID(t)
		accounts[i] = accountID
	}
	return accounts
}

func SortKeys(src []ed25519.PublicKey) {
	sort.Slice(src, func(i, j int) bool { return bytes.Compare(src[i], src[j]) < 0 })
}

func GenerateTransactionEnvelope(src xdr.AccountId, seqNum int, operations []xdr.Operation) xdr.TransactionEnvelope {
	return xdr.TransactionEnvelope{
		Tx: xdr.Transaction{
			SourceAccount: src,
			SeqNum:        xdr.SequenceNumber(seqNum),
			Operations:    operations,
		},
	}
}

func GenerateTransactionResult(code xdr.TransactionResultCode, results []xdr.OperationResult) xdr.TransactionResult {
	return xdr.TransactionResult{
		Result: xdr.TransactionResultResult{
			Code:    code,
			Results: &results,
		},
	}
}

func GenerateCreateOperation(src *xdr.AccountId, dest xdr.AccountId) xdr.Operation {
	return xdr.Operation{
		SourceAccount: src,
		Body: xdr.OperationBody{
			Type:            xdr.OperationTypeCreateAccount,
			CreateAccountOp: &xdr.CreateAccountOp{Destination: dest},
		},
	}
}

func GeneratePaymentOperation(src *xdr.AccountId, dest xdr.AccountId) xdr.Operation {
	return xdr.Operation{
		SourceAccount: src,
		Body: xdr.OperationBody{
			Type: xdr.OperationTypePayment,
			PaymentOp: &xdr.PaymentOp{
				Destination: dest,
				Amount:      10,
			},
		},
	}
}

func GenerateKin2PaymentOperation(src *xdr.AccountId, dest xdr.AccountId, issuer xdr.AccountId) xdr.Operation {
	assetCode := [4]byte{}
	copy(assetCode[:], "KIN")

	return xdr.Operation{
		SourceAccount: src,
		Body: xdr.OperationBody{
			Type: xdr.OperationTypePayment,
			PaymentOp: &xdr.PaymentOp{
				Destination: dest,
				Amount:      1000, // equivalent to 10 quarks
				Asset: xdr.Asset{
					Type: xdr.AssetTypeAssetTypeCreditAlphanum4,
					AlphaNum4: &xdr.AssetAlphaNum4{
						AssetCode: assetCode,
						Issuer:    issuer,
					},
				},
			},
		},
	}
}

func GenerateMergeOperation(src *xdr.AccountId, dest xdr.AccountId) xdr.Operation {
	return xdr.Operation{
		SourceAccount: src,
		Body: xdr.OperationBody{
			Type:        xdr.OperationTypeAccountMerge,
			Destination: &dest,
		},
	}
}

func GenerateLEC(lecType xdr.LedgerEntryChangeType, id xdr.AccountId, seqNum xdr.SequenceNumber, balance xdr.Int64) xdr.LedgerEntryChange {
	lec := xdr.LedgerEntryChange{
		Type: lecType,
	}
	entry := &xdr.LedgerEntry{
		Data: xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: id,
				SeqNum:    seqNum,
				Balance:   balance,
			},
		},
	}

	switch lecType {
	case xdr.LedgerEntryChangeTypeLedgerEntryCreated:
		lec.Created = entry
	case xdr.LedgerEntryChangeTypeLedgerEntryUpdated:
		lec.Updated = entry
	case xdr.LedgerEntryChangeTypeLedgerEntryRemoved:
		lec.Removed = &xdr.LedgerKey{
			Type:    xdr.LedgerEntryTypeAccount,
			Account: &xdr.LedgerKeyAccount{AccountId: id},
		}
	case xdr.LedgerEntryChangeTypeLedgerEntryState:
		lec.State = entry
	}
	return lec
}

func GenerateTransactionMeta(v int32, operations []xdr.OperationMeta) xdr.TransactionMeta {
	m := xdr.TransactionMeta{V: v}
	switch v {
	case 1:
		m.V1 = &xdr.TransactionMetaV1{
			Operations: operations,
		}
	default:
		m.Operations = &operations
	}
	return m
}
